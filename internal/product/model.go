package product

import "time"

type Product struct {
	ID             string    `json:"id"`
	StoreID        string    `json:"store_id"`
	Name           string    `json:"name"`
	Description    *string   `json:"description"`
	PriceCents     int       `json:"price_cents"`
	Currency       string    `json:"currency"`
	InventoryCount int       `json:"inventory_count"`
	Active         bool      `json:"active"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}
