package user

import (
	"context"
	"fmt"

	"github.com/mcv0000/storeflow-api/internal/auth"
)

type Service struct {
	repo      Repository
	jwtSecret string
}

func NewService(repo Repository, jwtSecret string) *Service {
	return &Service{
		repo:      repo,
		jwtSecret: jwtSecret,
	}
}

func (s *Service) Register(ctx context.Context, email, password, name string) (*User, string, error) {
	// Check if user already exists
	existingUser, _ := s.repo.GetByEmail(ctx, email)
	if existingUser != nil {
		return nil, "", fmt.Errorf("user with email %s already exists", email)
	}

	// Hash password
	passwordHash, err := auth.HashPassword(password)
	if err != nil {
		return nil, "", fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	user, err := s.repo.Create(ctx, email, passwordHash, name)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create user: %w", err)
	}

	// Generate token
	token, err := auth.GenerateToken(user.ID, s.jwtSecret)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate token: %w", err)
	}

	return user, token, nil
}

func (s *Service) Login(ctx context.Context, email, password string) (*User, string, error) {
	// Get user by email
	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		return nil, "", fmt.Errorf("invalid credentials")
	}

	// Check password
	if !auth.CheckPassword(password, user.PasswordHash) {
		return nil, "", fmt.Errorf("invalid credentials")
	}

	// Generate token
	token, err := auth.GenerateToken(user.ID, s.jwtSecret)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate token: %w", err)
	}

	return user, token, nil
}

func (s *Service) GetByID(ctx context.Context, id string) (*User, error) {
	return s.repo.GetByID(ctx, id)
}
