package order

import (
	"context"
	"fmt"
	"strings"

	"github.com/mcv0000/storeflow-api/internal/product"
	"github.com/mcv0000/storeflow-api/internal/store"
)

type Service struct {
	repo           Repository
	storeRepo      store.Repository
	productService *product.Service
}

func NewService(repo Repository, storeRepo store.Repository, productService *product.Service) *Service {
	return &Service{
		repo:           repo,
		storeRepo:      storeRepo,
		productService: productService,
	}
}

func (s *Service) Create(ctx context.Context, params CreateParams) (*Order, error) {
	params.CustomerEmail = strings.TrimSpace(params.CustomerEmail)
	params.StoreID = strings.TrimSpace(params.StoreID)

	if params.StoreID == "" {
		return nil, fmt.Errorf("store id is required")
	}

	if params.CustomerEmail == "" {
		return nil, fmt.Errorf("customer email is required")
	}

	if len(params.Items) == 0 {
		return nil, fmt.Errorf("order must contain at least one item")
	}

	for _, item := range params.Items {
		if strings.TrimSpace(item.ProductID) == "" {
			return nil, fmt.Errorf("product id is required")
		}

		if item.Quantity <= 0 {
			return nil, fmt.Errorf("quantity must be greater than zero")
		}
	}

	order, err := s.repo.Create(ctx, params)
	if err != nil {
		return nil, err
	}

	s.productService.InvalidateStoreCache(ctx, params.StoreID)

	return order, nil
}

func (s *Service) GetByID(ctx context.Context, userID, orderID string) (*Order, error) {
	order, err := s.repo.GetByID(ctx, orderID)
	if err != nil {
		return nil, err
	}

	if err := s.ensureStoreOwner(ctx, userID, order.StoreID); err != nil {
		return nil, err
	}

	return order, nil
}

func (s *Service) GetByStoreID(ctx context.Context, userID, storeID string, status *Status) ([]*Order, error) {
	if err := s.ensureStoreOwner(ctx, userID, storeID); err != nil {
		return nil, err
	}

	return s.repo.GetByStoreID(ctx, storeID, status)
}

func (s *Service) UpdateStatus(ctx context.Context, userID, orderID string, status Status) (*Order, error) {
	if !status.IsValid() {
		return nil, fmt.Errorf("invalid order status")
	}

	existingOrder, err := s.repo.GetByID(ctx, orderID)
	if err != nil {
		return nil, err
	}

	if err := s.ensureStoreOwner(ctx, userID, existingOrder.StoreID); err != nil {
		return nil, err
	}

	return s.repo.UpdateStatus(ctx, orderID, status)
}

func (s *Service) ensureStoreOwner(ctx context.Context, userID, storeID string) error {
	store, err := s.storeRepo.GetByID(ctx, storeID)
	if err != nil {
		return fmt.Errorf("store not found")
	}

	if store.OwnerID != userID {
		return fmt.Errorf("unauthorized: you do not own this store")
	}

	return nil
}
