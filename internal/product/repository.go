package product

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	Create(ctx context.Context, storeID, name string, description *string, priceCents int, currency string, inventoryCount int) (*Product, error)
	GetByStoreID(ctx context.Context, storeID string, activeOnly bool) ([]*Product, error)
}

type PostgresRepository struct {
	db *pgxpool.Pool
}

func NewPostgresRepository(db *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) Create(ctx context.Context, storeID, name string, description *string, priceCents int, currency string, inventoryCount int) (*Product, error) {
	query := `
		INSERT INTO products (store_id, name, description, price_cents, currency, inventory_count)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, store_id, name, description, price_cents, currency, inventory_count, active, created_at, updated_at
	`

	product := &Product{}
	err := r.db.QueryRow(ctx, query, storeID, name, description, priceCents, currency, inventoryCount).Scan(
		&product.ID,
		&product.StoreID,
		&product.Name,
		&product.Description,
		&product.PriceCents,
		&product.Currency,
		&product.InventoryCount,
		&product.Active,
		&product.CreatedAt,
		&product.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create product: %w", err)
	}

	return product, nil
}

func (r *PostgresRepository) GetByStoreID(ctx context.Context, storeID string, activeOnly bool) ([]*Product, error) {
	query := `
		SELECT id, store_id, name, description, price_cents, currency, inventory_count, active, created_at, updated_at
		FROM products
		WHERE store_id = $1
	`

	if activeOnly {
		query += " AND active = true"
	}

	query += " ORDER BY created_at DESC"

	rows, err := r.db.Query(ctx, query, storeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get products by store id: %w", err)
	}
	defer rows.Close()

	var products []*Product
	for rows.Next() {
		product := &Product{}
		err := rows.Scan(
			&product.ID,
			&product.StoreID,
			&product.Name,
			&product.Description,
			&product.PriceCents,
			&product.Currency,
			&product.InventoryCount,
			&product.Active,
			&product.CreatedAt,
			&product.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan product: %w", err)
		}
		products = append(products, product)
	}

	return products, nil
}
