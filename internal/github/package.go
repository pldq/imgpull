package github

import (
	"context"
	"fmt"
	"net/url"

	gh "github.com/google/go-github/v85/github"
)

var (
	container = "container"
)

// ListContainerPackages returns all packages of a specific type in the repository
func (g *Github) ListContainerPackages(ctx context.Context) ([]string, error) {
	packages, _, err := pageGetAll(func(page int) ([]*gh.Package, *gh.Response, error) {
		return g.client.Users.ListPackages(ctx, g.owner, &gh.PackageListOptions{
			PackageType: gh.Ptr(container),
			ListOptions: gh.ListOptions{
				PerPage: 100,
				Page:    page,
			},
		})
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list packages: %w", err)
	}

	names := make([]string, 0)
	for _, p := range packages {
		repository := p.GetRepository()
		fullName := fmt.Sprintf("%s/%s", g.owner, g.repo)
		if repository != nil && repository.GetFullName() == fullName {
			names = append(names, p.GetName())
		}
	}
	return names, nil
}

// DeleteContainerPackage deletes a package by name
func (g *Github) DeleteContainerPackage(ctx context.Context, packageName string) error {
	_, err := g.client.Users.DeletePackage(ctx, g.owner, container, url.PathEscape(packageName))
	if err != nil {
		return err
	}
	return nil
}

// PackageTagExists checks if a specific tag exists in a container package
func (g *Github) PackageTagExists(ctx context.Context, packageName, tag string) (bool, error) {
	versions, _, err := pageGetAll(func(page int) ([]*gh.PackageVersion, *gh.Response, error) {
		return g.client.Users.ListPackageVersions(ctx, container, url.PathEscape(packageName), &gh.ListPackageVersionsOptions{
			ListOptions: gh.ListOptions{
				PerPage: 100,
				Page:    page,
			},
		})
	})

	if err != nil {
		return false, fmt.Errorf("failed to list package versions: %w", err)
	}

	for _, v := range versions {
		metadata, ok := v.GetMetadata()
		if !ok || metadata.Container == nil {
			continue
		}
		for _, t := range metadata.Container.Tags {
			if t == tag {
				return true, nil
			}
		}
	}

	return false, nil
}
