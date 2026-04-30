package docker

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// CLIClient implements Client using docker CLI
type CLIClient struct{}

// PullImage pulls an image using docker CLI, output goes directly to stdout
func (c *CLIClient) PullImage(ctx context.Context, imageRef string) error {
	cmd := exec.CommandContext(ctx, "docker", "pull", imageRef)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("docker pull failed: %w", err)
	}

	fmt.Println("Pull complete")
	return nil
}

// TagImage tags an image using docker CLI
func (c *CLIClient) TagImage(ctx context.Context, sourceRef, targetRef string) error {
	cmd := exec.CommandContext(ctx, "docker", "tag", sourceRef, targetRef)
	_, err := cmd.CombinedOutput()
	return err
}

// GetImageSize returns image size using docker CLI
func (c *CLIClient) GetImageSize(ctx context.Context, imageRef string) (int64, error) {
	cmd := exec.CommandContext(ctx, "docker", "image", "inspect", imageRef, "--format", "{{.Size}}")
	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}
	sizeStr := strings.TrimSpace(string(output))
	var size int64
	_, _ = fmt.Sscanf(sizeStr, "%d", &size)
	return size, nil
}
