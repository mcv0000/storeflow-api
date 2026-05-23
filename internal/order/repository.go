package order

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mcv0000/storeflow-api/internal/product"
)

type Repository interface {
	Create(ctx context.Context, params CreateParams) (*Order, error)
	GetByID(ctx context.Context, id string) (*Order, error)
	GetByStoreID(ctx context.Context, storeID string, status *Status) ([]*Order, error)
	UpdateStatus(ctx context.Context, id string, status Status) (*Order, error)
}

type PostgresRepository struct {
	db *pgxpool.Pool
}

func NewPostgresRepository(db *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) Create(ctx context.Context, params CreateParams) (*Order, error) {
	if strings.TrimSpace(params.CustomerEmail) == "" {
		return nil, fmt.Errorf("customer email is required")
	}

	if len(params.Items) == 0 {
		return nil, fmt.Errorf("order must contain at least one item")
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin order transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	type itemData struct {
		product        *product.Product
		quantity       int
		unitPriceCents int
		totalCents     int
	}

	seenProducts := make(map[string]bool)
	items := make([]itemData, 0, len(params.Items))
	orderTotal := 0
	orderCurrency := ""

	for _, inputItem := range params.Items {
		if inputItem.Quantity <= 0 {
			return nil, fmt.Errorf("quantity must be greater than zero")
		}

		if strings.TrimSpace(inputItem.ProductID) == "" {
			return nil, fmt.Errorf("product id is required")
		}

		if seenProducts[inputItem.ProductID] {
			return nil, fmt.Errorf("duplicate product in order: %s", inputItem.ProductID)
		}
		seenProducts[inputItem.ProductID] = true

		p, err := r.getProductForUpdate(ctx, tx, inputItem.ProductID)
		if err != nil {
			return nil, err
		}

		if p.StoreID != params.StoreID {
			return nil, fmt.Errorf("product %s does not belong to store %s", p.ID, params.StoreID)
		}

		if !p.Active {
			return nil, fmt.Errorf("product %s is not active", p.ID)
		}

		if p.InventoryCount < inputItem.Quantity {
			return nil, fmt.Errorf("insufficient inventory for product %s", p.ID)
		}

		if orderCurrency == "" {
			orderCurrency = p.Currency
		}

		if p.Currency != orderCurrency {
			return nil, fmt.Errorf("mixed currencies are not supported")
		}

		itemTotal := p.PriceCents * inputItem.Quantity
		orderTotal += itemTotal

		items = append(items, itemData{
			product:        p,
			quantity:       inputItem.Quantity,
			unitPriceCents: p.PriceCents,
			totalCents:     itemTotal,
		})
	}

	order := &Order{}
	err = tx.QueryRow(ctx, `
        INSERT INTO orders (store_id, customer_email, status, total_cents, currency)
        VALUES ($1, $2, $3::order_status, $4, $5)
        RETURNING id, store_id, customer_email, status::text, total_cents, currency, created_at, updated_at
    `, params.StoreID, params.CustomerEmail, StatusPending, orderTotal, orderCurrency).Scan(
		&order.ID,
		&order.StoreID,
		&order.CustomerEmail,
		&order.Status,
		&order.TotalCents,
		&order.Currency,
		&order.CreatedAt,
		&order.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	order.Items = make([]*OrderItem, 0, len(items))

	for _, item := range items {
		orderItem := &OrderItem{}
		err := tx.QueryRow(ctx, `
            INSERT INTO order_items (order_id, product_id, quantity, unit_price_cents, total_cents)
            VALUES ($1, $2, $3, $4, $5)
            RETURNING id, order_id, product_id, quantity, unit_price_cents, total_cents, created_at
        `, order.ID, item.product.ID, item.quantity, item.unitPriceCents, item.totalCents).Scan(
			&orderItem.ID,
			&orderItem.OrderID,
			&orderItem.ProductID,
			&orderItem.Quantity,
			&orderItem.UnitPriceCents,
			&orderItem.TotalCents,
			&orderItem.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create order item: %w", err)
		}

		_, err = tx.Exec(ctx, `
            UPDATE products
            SET inventory_count = inventory_count - $1,
                updated_at = NOW()
            WHERE id = $2
        `, item.quantity, item.product.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to decrement product inventory: %w", err)
		}

		item.product.InventoryCount -= item.quantity
		orderItem.Product = item.product
		order.Items = append(order.Items, orderItem)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit order transaction: %w", err)
	}

	return order, nil
}

func (r *PostgresRepository) GetByID(ctx context.Context, id string) (*Order, error) {
	order := &Order{}

	err := r.db.QueryRow(ctx, `
        SELECT id, store_id, customer_email, status::text, total_cents, currency, created_at, updated_at
        FROM orders
        WHERE id = $1
    `, id).Scan(
		&order.ID,
		&order.StoreID,
		&order.CustomerEmail,
		&order.Status,
		&order.TotalCents,
		&order.Currency,
		&order.CreatedAt,
		&order.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get order by id: %w", err)
	}

	items, err := r.getItemsByOrderID(ctx, r.db, order.ID)
	if err != nil {
		return nil, err
	}
	order.Items = items

	return order, nil
}

func (r *PostgresRepository) GetByStoreID(ctx context.Context, storeID string, status *Status) ([]*Order, error) {
	query := `
        SELECT id, store_id, customer_email, status::text, total_cents, currency, created_at, updated_at
        FROM orders
        WHERE store_id = $1
    `
	args := []interface{}{storeID}

	if status != nil {
		query += " AND status = $2::order_status"
		args = append(args, *status)
	}

	query += " ORDER BY created_at DESC"

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get orders by store id: %w", err)
	}
	defer rows.Close()

	var orders []*Order

	for rows.Next() {
		order := &Order{}
		err := rows.Scan(
			&order.ID,
			&order.StoreID,
			&order.CustomerEmail,
			&order.Status,
			&order.TotalCents,
			&order.Currency,
			&order.CreatedAt,
			&order.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan order: %w", err)
		}

		items, err := r.getItemsByOrderID(ctx, r.db, order.ID)
		if err != nil {
			return nil, err
		}
		order.Items = items

		orders = append(orders, order)
	}

	return orders, nil
}

func (r *PostgresRepository) UpdateStatus(ctx context.Context, id string, status Status) (*Order, error) {
	if !status.IsValid() {
		return nil, fmt.Errorf("invalid order status: %s", status)
	}

	order := &Order{}
	err := r.db.QueryRow(ctx, `
        UPDATE orders
        SET status = $2::order_status,
            updated_at = NOW()
        WHERE id = $1
        RETURNING id, store_id, customer_email, status::text, total_cents, currency, created_at, updated_at
    `, id, status).Scan(
		&order.ID,
		&order.StoreID,
		&order.CustomerEmail,
		&order.Status,
		&order.TotalCents,
		&order.Currency,
		&order.CreatedAt,
		&order.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update order status: %w", err)
	}

	items, err := r.getItemsByOrderID(ctx, r.db, order.ID)
	if err != nil {
		return nil, err
	}
	order.Items = items

	return order, nil
}

func (r *PostgresRepository) getProductForUpdate(ctx context.Context, tx pgx.Tx, productID string) (*product.Product, error) {
	p := &product.Product{}

	err := tx.QueryRow(ctx, `
        SELECT id, store_id, name, description, price_cents, currency, inventory_count, active, created_at, updated_at
        FROM products
        WHERE id = $1
        FOR UPDATE
    `, productID).Scan(
		&p.ID,
		&p.StoreID,
		&p.Name,
		&p.Description,
		&p.PriceCents,
		&p.Currency,
		&p.InventoryCount,
		&p.Active,
		&p.CreatedAt,
		&p.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get product for order: %w", err)
	}

	return p, nil
}

type itemQueryer interface {
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
}

func (r *PostgresRepository) getItemsByOrderID(ctx context.Context, q itemQueryer, orderID string) ([]*OrderItem, error) {
	rows, err := q.Query(ctx, `
        SELECT
            oi.id,
            oi.order_id,
            oi.product_id,
            oi.quantity,
            oi.unit_price_cents,
            oi.total_cents,
            oi.created_at,
            p.id,
            p.store_id,
            p.name,
            p.description,
            p.price_cents,
            p.currency,
            p.inventory_count,
            p.active,
            p.created_at,
            p.updated_at
        FROM order_items oi
        JOIN products p ON p.id = oi.product_id
        WHERE oi.order_id = $1
        ORDER BY oi.created_at ASC
    `, orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to get order items: %w", err)
	}
	defer rows.Close()

	var items []*OrderItem

	for rows.Next() {
		item := &OrderItem{}
		p := &product.Product{}

		err := rows.Scan(
			&item.ID,
			&item.OrderID,
			&item.ProductID,
			&item.Quantity,
			&item.UnitPriceCents,
			&item.TotalCents,
			&item.CreatedAt,
			&p.ID,
			&p.StoreID,
			&p.Name,
			&p.Description,
			&p.PriceCents,
			&p.Currency,
			&p.InventoryCount,
			&p.Active,
			&p.CreatedAt,
			&p.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan order item: %w", err)
		}

		item.Product = p
		items = append(items, item)
	}

	return items, nil
}
