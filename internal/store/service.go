package store

import (
	"context"
	"fmt"
	"regexp"
	"strings"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, ownerID, name, slug string) (*Store, error) {
	// Generate slug if not provided
	if slug == "" {
		slug = generateSlug(name)
	}

	// Validate slug
	if !isValidSlug(slug) {
		return nil, fmt.Errorf("invalid slug format")
	}

	return s.repo.Create(ctx, ownerID, name, slug)
}

func (s *Service) GetByID(ctx context.Context, id string) (*Store, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) GetByOwnerID(ctx context.Context, ownerID string) ([]*Store, error) {
	return s.repo.GetByOwnerID(ctx, ownerID)
}

func generateSlug(name string) string {
	// Convert to lowercase
	slug := strings.ToLower(name)

	// Replace spaces with hyphens
	slug = strings.ReplaceAll(slug, " ", "-")

	// Remove non-alphanumeric characters except hyphens
	reg := regexp.MustCompile("[^a-z0-9-]+")
	slug = reg.ReplaceAllString(slug, "")

	// Remove consecutive hyphens
	reg = regexp.MustCompile("-+")
	slug = reg.ReplaceAllString(slug, "-")

	// Trim hyphens from start and end
	slug = strings.Trim(slug, "-")

	return slug
}

func isValidSlug(slug string) bool {
	// Slug must be lowercase alphanumeric with hyphens
	reg := regexp.MustCompile("^[a-z0-9]+(?:-[a-z0-9]+)*$")
	return reg.MatchString(slug)
}
