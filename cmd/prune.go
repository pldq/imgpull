package cmd

import (
	"context"
	"fmt"
	"image-poller/internal/github"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

// BranchPackage represents a branch and its associated package
type BranchPackage struct {
	Branch  string
	Package string
}

// Delete deletes the branch and package, returns counts of deleted resources
func (bp *BranchPackage) Delete(ctx context.Context, client github.Client) (int, int) {
	deletedBranches := 0
	deletedPackages := 0

	if bp.Branch != "" {
		if err := client.DeleteBranch(ctx, bp.Branch); err != nil {
			fmt.Printf("  ❌ Failed to delete branch %s: %v\n", bp.Branch, err)
		} else {
			fmt.Printf("  ✅ Deleted branch: %s\n", bp.Branch)
			deletedBranches++
		}
	}

	if bp.Package != "" {
		if err := client.DeleteContainerPackage(ctx, bp.Package); err != nil {
			fmt.Printf("  ❌ Failed to delete package %s: %v\n", bp.Package, err)
		} else {
			fmt.Printf("  ✅ Deleted package: %s\n", bp.Package)
			deletedPackages++
		}
	}

	return deletedBranches, deletedPackages
}

var pruneCmd = &cobra.Command{
	Use:   "prune",
	Short: "Clean up repository branches and packages",
	Long: `⚠ DANGER: This command will permanently delete resources!

This command performs the following destructive operations:
  • Delete ALL branches except 'master' (or the default branch)
  • Delete ALL container packages in the repository

These actions are IRREVERSIBLE. Make sure you understand the impact
before running this command.

Use --dry-run to preview what would be deleted without actually deleting.`,
	Args: cobra.NoArgs,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if err := cfg.Validate(); err != nil {
			return err
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		githubClient, err := github.NewClient(cfg.GitHubToken, cfg.GitHubRepo)
		if err != nil {
			return fmt.Errorf("failed to create GitHub client: %w", err)
		}

		ctx := context.Background()

		branchPackages, err := listResources(ctx, githubClient)
		if err != nil {
			return err
		}

		printResourceTable(branchPackages)

		if dryRun {
			fmt.Println("✅ Dry run completed. No resources were deleted.")
			return nil
		}

		if len(branchPackages) > 0 {
			fmt.Println("🗑️  Deleting resources...")
			totalDeletedBranches := 0
			totalDeletedPackages := 0
			for _, bp := range branchPackages {
				db, dp := bp.Delete(ctx, githubClient)
				totalDeletedBranches += db
				totalDeletedPackages += dp
			}

			fmt.Printf("✅ Prune completed: deleted %d branches, %d packages\n", totalDeletedBranches, totalDeletedPackages)
		}
		return nil
	},
}

func listResources(ctx context.Context, client github.Client) ([]BranchPackage, error) {
	fmt.Println("📋 Listing branches...")
	branches, err := client.ListBranches(ctx, []string{"master", "main", "dependabot/**"})
	if err != nil {
		return nil, fmt.Errorf("failed to list branches: %w", err)
	}

	fmt.Println("📦 Listing container packages...")
	packages, err := client.ListContainerPackages(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list packages: %w", err)
	}

	// Build association map
	branchSet := make(map[string]bool)
	for _, branch := range branches {
		branchSet[branch] = true
	}

	packageSet := make(map[string]bool)
	for _, pkg := range packages {
		packageSet[pkg] = true
	}

	// Create BranchPackage list with associations
	result := make([]BranchPackage, 0)

	// Branches with matching packages
	for _, branch := range branches {
		bp := BranchPackage{Branch: branch}
		if packageSet[branch] {
			bp.Package = branch
		}
		result = append(result, bp)
	}

	// Packages without matching branches
	for _, pkg := range packages {
		if !branchSet[pkg] {
			result = append(result, BranchPackage{Package: pkg})
		}
	}

	return result, nil
}

func printResourceTable(branchPackages []BranchPackage) {
	branchCount := 0
	packageCount := 0

	if len(branchPackages) > 0 {
		fmt.Println()
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		_, _ = fmt.Fprintln(w, "Branch\tPackage\t")
		_, _ = fmt.Fprintln(w, "------\t-------\t")
		for _, bp := range branchPackages {
			if bp.Branch != "" {
				branchCount++
			}
			if bp.Package != "" {
				packageCount++
			}
			_, _ = fmt.Fprintf(w, "%s\t%s\t\n", bp.Branch, bp.Package)
		}
		_ = w.Flush()
	}

	fmt.Printf("\nSummary: %d branches, %d packages to delete\n", branchCount, packageCount)
}

func init() {
	pruneCmd.Flags().Bool("dry-run", false, "Preview what would be deleted without actually deleting")
	rootCmd.AddCommand(pruneCmd)
}
