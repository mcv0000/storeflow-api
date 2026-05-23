package graphql

import (
	orderdomain "github.com/mcv0000/storeflow-api/internal/order"
	"github.com/mcv0000/storeflow-api/internal/product"
	"github.com/mcv0000/storeflow-api/internal/store"
	"github.com/mcv0000/storeflow-api/internal/user"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	userService    *user.Service
	storeService   *store.Service
	productService *product.Service
	orderService   *orderdomain.Service
}

func NewResolver(userService *user.Service, storeService *store.Service, productService *product.Service, orderService *orderdomain.Service) *Resolver {
	return &Resolver{
		userService:    userService,
		storeService:   storeService,
		productService: productService,
		orderService:   orderService,
	}
}
