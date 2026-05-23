package store

import "time"

type Store struct {
	ID        string    `json:"id"`
	OwnerID   string    `json:"owner_id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
