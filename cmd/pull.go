package cmd

import (
	"context"
	"errors"
	"fmt"
	"image-poller/pkg/display"
	"os"
	"strconv"
	"time"

	"image-poller/internal/config"
	"image-poller/internal/db"
	"image-poller/internal/docker"
	"image-poller/internal/github"
	"image-poller/pkg/image"
	"image-poller/utils"

	gh "github.com/google/go-github/v85/github"
	"github.com/spf13/cobra"
)

var (
	githubClient github.Client
	dockerClient docker.Client
)

var pullCmd = &cobra.Command{
	Use:   "pull <image>",
	Short: "Pull a Docker image",
	Long: `Pull a Docker image from Docker Hub via GitHub Actions.

Examples:
  imgpull pull nginx                    # Pull nginx:latest
  imgpull pull nginx:1.21               # Pull nginx:1.21
  imgpull pull prom/prometheus:v2.45.0  # Pull prom/prometheus:v2.45.0
  imgpull pull nginx --force            # Force re-pull even if branch exists`,
	Args: cobra.ExactArgs(1),
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if err := cfg.Validate(); err != nil {
			return err
		}

		var err error
		githubClient, err = github.NewClient(cfg.GitHubToken, cfg.GitHubRepo)
		if err != nil {
			return fmt.Errorf("failed to create GitHub client: %w", err)
		}

		dockerClient, err = docker.NewClient(cfg.DockerMode == config.DockerModeCLI, cfg.DockerHost)
		if err != nil {
			return fmt.Errorf("failed to create Docker client: %w", err)
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		imageRef := args[0]
		force, _ := cmd.Flags().GetBool("force")

		ref, err := image.Parse(imageRef)
		if err != nil {
			return fmt.Errorf("invalid image reference: %w", err)
		}

		ctx := context.Background()
		originalRef := ref.OriginalReference()

		fmt.Printf("Image: %s\n", ref.OriginalReference())
		fmt.Printf("Branch: %s\n", ref.BranchName())
		fmt.Printf("Target: %s\n", ref.GHCRReference(githubClient.GetRepoOwner()))

		sw := utils.NewStopWatch()

		puller := &ImagePuller{
			ctx:      ctx,
			ref:      *ref,
			onlyPull: !force,
			gh:       githubClient,
			docker:   dockerClient,
		}

		if err := puller.Run(sw); err != nil {
			return err
		}

		imageSize, _ := dockerClient.GetImageSize(ctx, originalRef)
		record := &db.ImageRecord{
			ImageName: ref.Name,
			Tag:       ref.Tag,
			Size:      imageSize,
			PullTime:  sw.StartTime(),
			Duration:  sw.Total().Milliseconds(),
		}
		if err := database.UpsertRecord(record); err != nil {
			fmt.Printf("Warning: failed to save record: %v\n", err)
		}

		fmt.Printf("\nSuccessfully pulled %s (took %s)\n", originalRef, sw.Total().Round(time.Second))
		return nil
	},
}

func init() {
	pullCmd.Flags().Bool("force", false, "Force re-pull even if branch exists")
	rootCmd.AddCommand(pullCmd)
}

type ImagePuller struct {
	ctx          context.Context
	ref          image.Reference
	onlyPull     bool
	branchExists bool
	gh           github.Client
	docker       docker.Client
}

func (p *ImagePuller) Run(sw *utils.StopWatch) error {
	// Check if package tag already exists (skip for latest tag)
	tagExists := p.checkPackageTagExisted()

	steps := []struct {
		name        string
		fn          func() error
		skipOnExist bool // skip this step when package tag already exists
	}{
		{name: "Checking branch", fn: p.ensureBranch, skipOnExist: true},
		{name: "Running workflow", fn: p.runWorkflow, skipOnExist: true},
		{name: "Pulling image from ghcr.io", fn: p.pullImage},
		{name: "Renaming image", fn: p.renameImage},
	}

	for i, step := range steps {
		// Skip steps marked with skipOnTag if tag already exists
		if tagExists && step.skipOnExist {
			fmt.Printf("\n[%d/%d] %s... (skipped, tag already exists)\n", i+1, len(steps), step.name)
			continue
		}
		fmt.Printf("\n[%d/%d] %s...\n", i+1, len(steps), step.name)
		if err := sw.Run(step.name, step.fn); err != nil {
			return err
		}
	}

	return nil
}

func (p *ImagePuller) checkPackageTagExisted() bool {
	if p.ref.Tag == "latest" || !p.onlyPull {
		return false
	}

	tagExists, err := p.gh.PackageTagExists(p.ctx, p.ref.BranchName(), p.ref.Tag)
	if err != nil {
		fmt.Printf("Warning: failed to check package tag: %v\n", err)
		return false
	}

	if tagExists {
		fmt.Printf("Package '%s' with tag '%s' already exists in ghcr.io\n", p.ref.BranchName(), p.ref.Tag)
		return true
	}

	return false
}

func (p *ImagePuller) ensureBranch() (err error) {
	p.branchExists, err = p.gh.BranchExists(p.ctx, p.ref.BranchName())
	if err != nil {
		return fmt.Errorf("failed to check branch: %w", err)
	}

	if p.branchExists {
		fmt.Printf("Branch '%s' already exists\n", p.ref.BranchName())
		return nil
	}

	// Get last runID before creating branch (to exclude old workflow)
	lastRunID, _ := p.gh.GetLatestWorkflowRunByBranch(p.ctx, "trans-image.yml", p.ref.BranchName())
	fmt.Printf("Creating branch '%s'...\n", p.ref.BranchName())
	if err := p.gh.CreateBranch(p.ctx, p.ref.BranchName()); err != nil {
		return fmt.Errorf("failed to create branch: %w", err)
	}
	fmt.Printf("Branch '%s' created\n", p.ref.BranchName())

	retry := utils.NewRetry[int64](10)
	newRunID, err := retry.Do(func() (int64, error) {
		runID, err := p.gh.GetLatestWorkflowRunByBranch(p.ctx, "trans-image.yml", p.ref.BranchName())
		if err != nil {
			return 0, err
		}
		if runID <= lastRunID {
			return 0, errors.New("workflow not triggered yet")
		}

		return runID, nil
	})
	if err != nil {
		return fmt.Errorf("failed to get new workflow run ID: %w", err)
	}
	if p.ref.Tag == "latest" {
		fmt.Printf("Waiting for auto-triggered workflow: trans-image.yml (runID: %d)\r\n", newRunID)
		return p.waitForWorkflow(newRunID)
	}

	fmt.Printf("Cancel for auto-triggered workflow: trans-image.yml (runID: %d)\r\n", newRunID)
	return p.gh.CancelWorkflowRun(p.ctx, newRunID)
}

func (p *ImagePuller) runWorkflow() error {
	// Skip if new branch + latest (already waited in ensureBranch)
	if !p.branchExists && p.ref.Tag == "latest" {
		return nil
	}

	// Need to manually trigger pull-image-with-tag.yml
	workflowFile := "pull-image-with-tag.yml"
	runID, err := p.gh.TriggerWorkflow(p.ctx, workflowFile, p.ref.BranchName(), p.ref.Tag)
	if err != nil {
		return fmt.Errorf("failed to trigger workflow: %w", err)
	}
	fmt.Printf("Workflow triggered: %s (runID: %d)\n", workflowFile, runID)

	return p.waitForWorkflow(runID)
}

func (p *ImagePuller) waitForWorkflow(runID int64) error {
	eventProcessor := display.Full(os.Stdout)
	event := &display.Event{
		ID:     strconv.FormatInt(runID, 10),
		Text:   fmt.Sprintf("%d", runID),
		Status: display.Working,
	}
	eventProcessor.Start(p.ctx, *event)
	defer eventProcessor.Done()
	err := p.gh.WaitForWorkflow(p.ctx, runID, func(status github.WorkflowStatus, run *gh.WorkflowRun) (bool, error) {
		switch status {
		case github.StatusPending:
			eventProcessor.OnDetails(event, "Queued")
		case github.StatusInProgress:
			eventProcessor.OnDetails(event, "Running")
		case github.StatusSuccess:
			event.Details = "Success"
			event.Status = display.Done
			eventProcessor.On(*event)
			return true, nil
		case github.StatusFailed:
			event.Details = fmt.Sprintf("%s %s", run.GetConclusion(), run.GetHTMLURL())
			event.Status = display.Error
			eventProcessor.On(*event)
		}

		if status == github.StatusFailed || status == github.StatusCancelled {
			return true, errors.New(fmt.Sprintf("Workflow running failed with conclusion: %s", run.GetConclusion()))
		}

		return false, nil
	})

	return err
}

func (p *ImagePuller) pullImage() error {
	ghcrRef := p.ref.GHCRReference(p.gh.GetRepoOwner())
	if err := p.docker.PullImage(p.ctx, ghcrRef); err != nil {
		return fmt.Errorf("failed to pull image: %w", err)
	}
	return nil
}

func (p *ImagePuller) renameImage() error {
	ghcrRef := p.ref.GHCRReference(p.gh.GetRepoOwner())
	originalRef := p.ref.OriginalReference()

	if err := p.docker.TagImage(p.ctx, ghcrRef, originalRef); err != nil {
		return fmt.Errorf("failed to tag image: %w", err)
	}
	fmt.Printf("Image renamed to: %s\n", originalRef)
	return nil
}
