package order

import (
	"fmt"
	"io"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/mcv0000/storeflow-api/internal/product"
)

type Status string

const (
	StatusPending   Status = "PENDING"
	StatusPaid      Status = "PAID"
	StatusShipped   Status = "SHIPPED"
	StatusCancelled Status = "CANCELLED"
)

func (s Status) IsValid() bool {
	switch s {
	case StatusPending, StatusPaid, StatusShipped, StatusCancelled:
		return true
	default:
		return false
	}
}

func (s Status) String() string {
	return string(s)
}

func (s *Status) UnmarshalGQL(v interface{}) error {
	value, ok := v.(string)
	if !ok {
		return fmt.Errorf("order status must be a string")
	}

	status := Status(value)
	if !status.IsValid() {
		return fmt.Errorf("invalid order status: %s", value)
	}

	*s = status
	return nil
}

func (s Status) MarshalGQL(w io.Writer) {
	graphql.MarshalString(string(s)).MarshalGQL(w)
}

type Order struct {
	ID            string       `json:"id"`
	StoreID       string       `json:"store_id"`
	CustomerEmail string       `json:"customer_email"`
	Status        Status       `json:"status"`
	TotalCents    int          `json:"total_cents"`
	Currency      string       `json:"currency"`
	Items         []*OrderItem `json:"items"`
	CreatedAt     time.Time    `json:"created_at"`
	UpdatedAt     time.Time    `json:"updated_at"`
}

type OrderItem struct {
	ID             string           `json:"id"`
	OrderID        string           `json:"order_id"`
	ProductID      string           `json:"product_id"`
	Product        *product.Product `json:"product"`
	Quantity       int              `json:"quantity"`
	UnitPriceCents int              `json:"unit_price_cents"`
	TotalCents     int              `json:"total_cents"`
	CreatedAt      time.Time        `json:"created_at"`
}

type CreateParams struct {
	StoreID       string
	CustomerEmail string
	Items         []CreateItemParams
}

type CreateItemParams struct {
	ProductID string
	Quantity  int
}
