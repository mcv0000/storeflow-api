package store

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	Create(ctx context.Context, ownerID, name, slug string) (*Store, error)
	GetByID(ctx context.Context, id string) (*Store, error)
	GetByOwnerID(ctx context.Context, ownerID string) ([]*Store, error)
}

type PostgresRepository struct {
	db *pgxpool.Pool
}

func NewPostgresRepository(db *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) Create(ctx context.Context, ownerID, name, slug string) (*Store, error) {
	query := `
		INSERT INTO stores (owner_id, name, slug)
		VALUES ($1, $2, $3)
		RETURNING id, owner_id, name, slug, created_at, updated_at
	`

	store := &Store{}
	err := r.db.QueryRow(ctx, query, ownerID, name, slug).Scan(
		&store.ID,
		&store.OwnerID,
		&store.Name,
		&store.Slug,
		&store.CreatedAt,
		&store.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create store: %w", err)
	}

	return store, nil
}

func (r *PostgresRepository) GetByID(ctx context.Context, id string) (*Store, error) {
	query := `
		SELECT id, owner_id, name, slug, created_at, updated_at
		FROM stores
		WHERE id = $1
	`

	store := &Store{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&store.ID,
		&store.OwnerID,
		&store.Name,
		&store.Slug,
		&store.CreatedAt,
		&store.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get store by id: %w", err)
	}

	return store, nil
}

func (r *PostgresRepository) GetByOwnerID(ctx context.Context, ownerID string) ([]*Store, error) {
	query := `
		SELECT id, owner_id, name, slug, created_at, updated_at
		FROM stores
		WHERE owner_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query, ownerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get stores by owner id: %w", err)
	}
	defer rows.Close()

	var stores []*Store
	for rows.Next() {
		store := &Store{}
		err := rows.Scan(
			&store.ID,
			&store.OwnerID,
			&store.Name,
			&store.Slug,
			&store.CreatedAt,
			&store.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan store: %w", err)
		}
		stores = append(stores, store)
	}

	return stores, nil
}
