package user

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	Create(ctx context.Context, email, passwordHash, name string) (*User, error)
	GetByID(ctx context.Context, id string) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
}

type PostgresRepository struct {
	db *pgxpool.Pool
}

func NewPostgresRepository(db *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) Create(ctx context.Context, email, passwordHash, name string) (*User, error) {
	query := `
		INSERT INTO users (email, password_hash, name)
		VALUES ($1, $2, $3)
		RETURNING id, email, password_hash, name, created_at, updated_at
	`

	user := &User{}
	err := r.db.QueryRow(ctx, query, email, passwordHash, name).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Name,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

func (r *PostgresRepository) GetByID(ctx context.Context, id string) (*User, error) {
	query := `
		SELECT id, email, password_hash, name, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	user := &User{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Name,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by id: %w", err)
	}

	return user, nil
}

func (r *PostgresRepository) GetByEmail(ctx context.Context, email string) (*User, error) {
	query := `
		SELECT id, email, password_hash, name, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	user := &User{}
	err := r.db.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Name,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return user, nil
}
