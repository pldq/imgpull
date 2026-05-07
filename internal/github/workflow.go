package github

import (
	"context"
	"errors"
	"fmt"
	"time"

	gh "github.com/google/go-github/v85/github"
)

// TriggerWorkflow triggers a workflow_dispatch event
func (g *Github) TriggerWorkflow(ctx context.Context, workflowName, branchName, tag string) (int64, error) {
	returnRunDetails := true
	event := gh.CreateWorkflowDispatchEventRequest{
		Ref: branchName,
		Inputs: map[string]interface{}{
			"tag": tag,
		},
		ReturnRunDetails: &returnRunDetails,
	}

	detail, _, err := g.client.Actions.CreateWorkflowDispatchEventByFileName(
		ctx, g.owner, g.repo, workflowName, event)
	if err != nil {
		return 0, fmt.Errorf("failed to trigger workflow: %w", err)
	}
	return *detail.WorkflowRunID, nil
}

// GetLatestWorkflowRunByBranch gets the latest workflow run ID for a workflow on a branch
func (g *Github) GetLatestWorkflowRunByBranch(ctx context.Context, workflowName, branchName string) (int64, error) {
	runs, _, err := g.client.Actions.ListWorkflowRunsByFileName(
		ctx, g.owner, g.repo, workflowName,
		&gh.ListWorkflowRunsOptions{
			Branch: branchName,
			ListOptions: gh.ListOptions{
				PerPage: 1,
			},
		})
	if err != nil {
		return 0, fmt.Errorf("failed to list workflow runs: %w", err)
	}

	if len(runs.WorkflowRuns) == 0 {
		return 0, fmt.Errorf("no workflow runs found")
	}

	if runs.WorkflowRuns[0].ID == nil {
		return 0, fmt.Errorf("workflow run ID is nil")
	}

	return *runs.WorkflowRuns[0].ID, nil
}

// WaitForWorkflow waits for a workflow run to complete by run ID
func (g *Github) WaitForWorkflow(ctx context.Context, runID int64, callback WaitWorkFlowCallback) error {
	pollInterval := 2 * time.Second

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		run, _, err := g.client.Actions.GetWorkflowRunByID(ctx, g.owner, g.repo, runID)
		if err != nil {
			return fmt.Errorf("failed to get workflow run: %w", err)
		}

		status := mapWorkflowStatus(run)

		if callback != nil {
			finished, err := callback(status, run)
			if finished || err != nil {
				return err
			}
		} else {
			return errors.New("WaitWorkFlowCallback is required")
		}
		time.Sleep(pollInterval)
	}
}

func (g *Github) CancelWorkflowRun(ctx context.Context, runID int64) error {
	response, err := g.client.Actions.CancelWorkflowRunByID(ctx, g.owner, g.repo, runID)
	if response != nil && response.StatusCode == 202 {
		return nil
	}
	return err
}

func mapWorkflowStatus(run *gh.WorkflowRun) WorkflowStatus {
	switch run.GetStatus() {
	case "queued", "pending":
		return StatusPending
	case "in_progress", "waiting":
		return StatusInProgress
	case "completed":
		switch run.GetConclusion() {
		case "success":
			return StatusSuccess
		case "failure", "timed_out":
			return StatusFailed
		case "cancelled":
			return StatusCancelled
		default:
			return StatusFailed
		}
	default:
		return StatusPending
	}
}
