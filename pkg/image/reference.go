package image

import (
	"fmt"
	"strings"
)

// Reference represents a parsed docker image reference
type Reference struct {
	// Registry is the registry host (e.g., "docker.io", "ghcr.io")
	Registry string
	// Name is the image name without tag (e.g., "prom/prometheus", "library/nginx")
	Name string
	// Tag is the image tag (e.g., "latest", "v2.45.0")
	Tag string
}

// Parse parses an image string into a Reference
// Supported formats:
//   - nginx -> docker.io/library/nginx:latest
//   - nginx:1.21 -> docker.io/library/nginx:1.21
//   - prom/prometheus -> docker.io/prom/prometheus:latest
//   - prom/prometheus:v2.45.0 -> docker.io/prom/prometheus:v2.45.0
//   - ghcr.io/pldq/prom/prometheus -> ghcr.io/pldq/prom/prometheus:latest
func Parse(image string) (*Reference, error) {
	if image == "" {
		return nil, fmt.Errorf("image name cannot be empty")
	}

	ref := &Reference{
		Tag: "latest",
	}

	// Parse registry if present
	parts := strings.SplitN(image, "/", 2)
	if len(parts) == 2 && (strings.Contains(parts[0], ".") || strings.Contains(parts[0], ":") || parts[0] == "localhost") {
		// Has registry prefix
		ref.Registry = parts[0]
		image = parts[1]
	} else {
		ref.Registry = "docker.io"
	}

	// Parse tag
	if idx := strings.LastIndex(image, ":"); idx > 0 {
		// Check if it's not part of a port (e.g., localhost:5000/image)
		if !strings.Contains(image[idx:], "/") {
			ref.Name = image[:idx]
			ref.Tag = image[idx+1:]
		} else {
			ref.Name = image
		}
	} else {
		ref.Name = image
	}

	// Handle docker.io official images (add library/ prefix)
	if ref.Registry == "docker.io" && !strings.Contains(ref.Name, "/") {
		ref.Name = "library/" + ref.Name
	}

	if ref.Name == "" {
		return nil, fmt.Errorf("invalid image name: %s", image)
	}

	return ref, nil
}

// String returns the full image reference string
func (r *Reference) String() string {
	return fmt.Sprintf("%s/%s:%s", r.Registry, r.Name, r.Tag)
}

// BranchName returns the branch name for GitHub (same as image name without tag)
func (r *Reference) BranchName() string {
	return r.Name
}

// GHCRName returns the full image name in ghcr.io
func (r *Reference) GHCRName(githubUser string) string {
	return fmt.Sprintf("ghcr.io/%s/%s", githubUser, r.Name)
}

// GHCRReference returns the full image reference in ghcr.io with tag
func (r *Reference) GHCRReference(githubUser string) string {
	return fmt.Sprintf("%s:%s", r.GHCRName(githubUser), r.Tag)
}

// OriginalReference returns the original docker hub reference
func (r *Reference) OriginalReference() string {
	// Remove library/ prefix for display
	name := r.Name
	if strings.HasPrefix(name, "library/") {
		name = strings.TrimPrefix(name, "library/")
	}
	return fmt.Sprintf("%s:%s", name, r.Tag)
}
