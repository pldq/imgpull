package github

import (
	"context"
	"fmt"

	"github.com/bmatcuk/doublestar/v4"
	gh "github.com/google/go-github/v85/github"
)

// BranchExists checks if a branch exists
func (g *Github) BranchExists(ctx context.Context, branchName string) (bool, error) {
	_, response, err := g.client.Repositories.GetBranch(ctx, g.owner, g.repo, branchName, 1)
	if response != nil && response.StatusCode == 404 {
		return false, nil

	}
	if err != nil {
		return false, fmt.Errorf("failed to check branch: %w", err)
	}
	return true, nil
}

// CreateBranch creates a new branch from the default branch
func (g *Github) CreateBranch(ctx context.Context, branchName string) error {
	repo, _, err := g.client.Repositories.Get(ctx, g.owner, g.repo)
	if err != nil {
		return fmt.Errorf("failed to get repository: %w", err)
	}

	ref, _, err := g.client.Git.GetRef(ctx, g.owner, g.repo, "refs/heads/"+*repo.DefaultBranch)
	if err != nil {
		return fmt.Errorf("failed to get default branch ref: %w", err)
	}

	newRef := gh.CreateRef{
		Ref: "refs/heads/" + branchName,
		SHA: *ref.Object.SHA,
	}

	_, _, err = g.client.Git.CreateRef(ctx, g.owner, g.repo, newRef)
	if err != nil {
		return fmt.Errorf("failed to create branch: %w", err)
	}

	return nil
}

// ListBranches returns all branches in the repository
func (g *Github) ListBranches(ctx context.Context, excludeBranches []string) ([]string, error) {
	branches, _, err := pageGetAll(func(page int) ([]*gh.Branch, *gh.Response, error) {
		return g.client.Repositories.ListBranches(ctx, g.owner, g.repo, &gh.BranchListOptions{
			ListOptions: gh.ListOptions{PerPage: 100, Page: page},
		})
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list branches: %w", err)
	}

	names := make([]string, 0)
loop:
	for _, branch := range branches {
		name := branch.GetName()
		for _, pattern := range excludeBranches {
			matched, err := doublestar.Match(pattern, name)
			if err != nil {
				return names, err
			}
			if matched {
				continue loop
			}
		}
		names = append(names, name)
	}
	return names, nil
}

// DeleteBranch deletes a branch by name
func (g *Github) DeleteBranch(ctx context.Context, branchName string) error {
	_, err := g.client.Git.DeleteRef(ctx, g.owner, g.repo, "heads/"+branchName)
	if err != nil {
		return fmt.Errorf("failed to delete branch %s: %w", branchName, err)
	}
	return nil
}
