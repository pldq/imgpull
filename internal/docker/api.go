package docker

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/moby/moby/client"
)

// APIClient implements Client using Docker API
type APIClient struct {
	cli *client.Client
}

// PullImage pulls an image using Docker API, outputs progress to stdout
func (c *APIClient) PullImage(ctx context.Context, imageRef string) error {
	out, err := c.cli.ImagePull(ctx, imageRef, client.ImagePullOptions{})
	if err != nil {
		return err
	}
	defer out.Close()

	decoder := json.NewDecoder(out)
	for {
		var msg struct {
			Status   string `json:"status"`
			ID       string `json:"id"`
			Progress string `json:"progress"`
			Error    string `json:"error"`
		}
		if err := decoder.Decode(&msg); err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		if msg.Error != "" {
			return fmt.Errorf("pull error: %s", msg.Error)
		}

		if msg.ID != "" {
			if msg.Progress != "" {
				_, _ = fmt.Fprintf(os.Stdout, "%s: %s %s\n", msg.ID, msg.Status, msg.Progress)
			} else {
				_, _ = fmt.Fprintf(os.Stdout, "%s: %s\n", msg.ID, msg.Status)
			}
		} else if msg.Status != "" {
			_, _ = fmt.Fprintln(os.Stdout, msg.Status)
		}
	}

	_, _ = fmt.Fprintln(os.Stdout, "Pull complete")
	return nil
}

// TagImage tags an image using Docker API
func (c *APIClient) TagImage(ctx context.Context, sourceRef, targetRef string) error {
	_, err := c.cli.ImageTag(ctx, client.ImageTagOptions{
		Source: sourceRef,
		Target: targetRef,
	})
	return err
}

// GetImageSize returns image size using Docker API
func (c *APIClient) GetImageSize(ctx context.Context, imageRef string) (int64, error) {
	images, err := c.cli.ImageList(ctx, client.ImageListOptions{
		Filters: make(client.Filters).Add("reference", imageRef),
	})
	if err != nil {
		return 0, err
	}
	if len(images.Items) == 0 {
		return 0, fmt.Errorf("image not found: %s", imageRef)
	}
	return images.Items[0].Size, nil
}
