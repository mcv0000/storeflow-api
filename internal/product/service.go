package product

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/mcv0000/storeflow-api/internal/store"
	"github.com/redis/go-redis/v9"
)

type Service struct {
	repo      Repository
	storeRepo store.Repository
	cache     *redis.Client
}

func NewService(repo Repository, storeRepo store.Repository, cache *redis.Client) *Service {
	return &Service{
		repo:      repo,
		storeRepo: storeRepo,
		cache:     cache,
	}
}

func (s *Service) Create(ctx context.Context, userID, storeID, name string, description *string, priceCents int, currency string, inventoryCount int) (*Product, error) {
	store, err := s.storeRepo.GetByID(ctx, storeID)
	if err != nil {
		return nil, fmt.Errorf("store not found")
	}

	if store.OwnerID != userID {
		return nil, fmt.Errorf("unauthorized: you don't own this store")
	}

	product, err := s.repo.Create(ctx, storeID, name, description, priceCents, currency, inventoryCount)
	if err != nil {
		return nil, err
	}

	s.InvalidateStoreCache(ctx, storeID)

	return product, nil
}

func (s *Service) GetByStoreID(ctx context.Context, storeID string, activeOnly bool) ([]*Product, error) {
	cacheKey := s.getCacheKey(storeID, activeOnly)

	cached, err := s.cache.Get(ctx, cacheKey).Result()
	if err == nil {
		var products []*Product
		if err := json.Unmarshal([]byte(cached), &products); err == nil {
			return products, nil
		}
	}

	products, err := s.repo.GetByStoreID(ctx, storeID, activeOnly)
	if err != nil {
		return nil, err
	}

	data, err := json.Marshal(products)
	if err == nil {
		s.cache.Set(ctx, cacheKey, data, 60*time.Second)
	}

	return products, nil
}

func (s *Service) InvalidateStoreCache(ctx context.Context, storeID string) {
	s.cache.Del(ctx, s.getCacheKey(storeID, true))
	s.cache.Del(ctx, s.getCacheKey(storeID, false))
}

func (s *Service) getCacheKey(storeID string, activeOnly bool) string {
	return fmt.Sprintf("store:%s:products:active:%t", storeID, activeOnly)
}
