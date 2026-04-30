package docker

import (
	"context"

	"github.com/moby/moby/client"
)

// Client defines the interface for Docker operations
type Client interface {
	PullImage(ctx context.Context, imageRef string) error
	TagImage(ctx context.Context, sourceRef, targetRef string) error
	GetImageSize(ctx context.Context, imageRef string) (int64, error)
}

// NewClient creates a Docker client based on mode
func NewClient(cliMode bool, host string) (Client, error) {
	if cliMode {
		return &CLIClient{}, nil
	}

	opts := []client.Opt{client.FromEnv}
	if host != "" {
		opts = []client.Opt{client.WithHost(host)}
	}

	apiCli, err := client.New(opts...)
	if err != nil {
		return nil, err
	}

	return &APIClient{cli: apiCli}, nil
}
