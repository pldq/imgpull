package github

import (
	"context"
	"fmt"
	"strings"

	gh "github.com/google/go-github/v85/github"
)

// WorkflowStatus represents the status of a workflow run
type WorkflowStatus string

const (
	StatusPending    WorkflowStatus = "pending"
	StatusInProgress WorkflowStatus = "in_progress"
	StatusSuccess    WorkflowStatus = "success"
	StatusFailed     WorkflowStatus = "failed"
	StatusCancelled  WorkflowStatus = "cancelled"
)

type WaitWorkFlowCallback func(status WorkflowStatus, run *gh.WorkflowRun) (bool, error)

// Client defines the interface for GitHub operations
type Client interface {
	BranchExists(ctx context.Context, branchName string) (bool, error)
	CreateBranch(ctx context.Context, branchName string) error
	ListBranches(ctx context.Context, excludeBranches []string) ([]string, error)
	DeleteBranch(ctx context.Context, branchName string) error
	ListContainerPackages(ctx context.Context) ([]string, error)
	DeleteContainerPackage(ctx context.Context, packageName string) error
	PackageTagExists(ctx context.Context, packageName, tag string) (bool, error)
	TriggerWorkflow(ctx context.Context, workflowName, branchName, tag string) (int64, error)
	CancelWorkflowRun(ctx context.Context, runID int64) error
	GetLatestWorkflowRunByBranch(ctx context.Context, workflowName, branchName string) (int64, error)
	WaitForWorkflow(ctx context.Context, runID int64, callback WaitWorkFlowCallback) error
	GetRepoOwner() string
}

// Github handles GitHub API operations
type Github struct {
	client *gh.Client
	owner  string
	repo   string
}

// GetRepoOwner returns the repository owner
func (g *Github) GetRepoOwner() string {
	return g.owner
}

type pageRequest[T any] func(page int) ([]T, *gh.Response, error)

func pageGetAll[T any](request pageRequest[T]) (total []T, response *gh.Response, err error) {
	page := 1
	for {
		data, response, err := request(page)
		page++
		if err != nil {
			return total, response, err
		}
		total = append(total, data...)
		if page > response.GetLastPage() {
			return total, response, err
		}
	}
}

// NewClient creates a new GitHub client
func NewClient(token, repo string) (Client, error) {
	if token == "" {
		return nil, fmt.Errorf("GitHub token is required")
	}

	parts := strings.Split(repo, "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid repo format, expected 'owner/repo'")
	}

	client := gh.NewClient(nil).WithAuthToken(token)
	return &Github{
		client: client,
		owner:  parts[0],
		repo:   parts[1],
	}, nil
}
