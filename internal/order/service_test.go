package order

import (
	"context"
	"testing"
)

type fakeOrderRepo struct {
	createCalled bool
}

func (f *fakeOrderRepo) Create(ctx context.Context, params CreateParams) (*Order, error) {
	f.createCalled = true
	return &Order{ID: "order-1", StoreID: params.StoreID}, nil
}

func (f *fakeOrderRepo) GetByID(ctx context.Context, id string) (*Order, error) {
	return nil, nil
}

func (f *fakeOrderRepo) GetByStoreID(ctx context.Context, storeID string, status *Status) ([]*Order, error) {
	return nil, nil
}

func (f *fakeOrderRepo) UpdateStatus(ctx context.Context, id string, status Status) (*Order, error) {
	return nil, nil
}

func TestStatusIsValid(t *testing.T) {
	validStatuses := []Status{
		StatusPending,
		StatusPaid,
		StatusShipped,
		StatusCancelled,
	}

	for _, status := range validStatuses {
		if !status.IsValid() {
			t.Fatalf("expected status %s to be valid", status)
		}
	}

	invalidStatus := Status("REFUNDED")
	if invalidStatus.IsValid() {
		t.Fatal("expected unknown status to be invalid")
	}
}

func TestServiceCreateRejectsEmptyItems(t *testing.T) {
	repo := &fakeOrderRepo{}
	service := NewService(repo, nil, nil)

	_, err := service.Create(context.Background(), CreateParams{
		StoreID:       "store-1",
		CustomerEmail: "customer@example.com",
		Items:         []CreateItemParams{},
	})

	if err == nil {
		t.Fatal("expected error for empty order items")
	}

	if repo.createCalled {
		t.Fatal("repository should not be called when order items are empty")
	}
}

func TestServiceCreateRejectsInvalidQuantity(t *testing.T) {
	repo := &fakeOrderRepo{}
	service := NewService(repo, nil, nil)

	_, err := service.Create(context.Background(), CreateParams{
		StoreID:       "store-1",
		CustomerEmail: "customer@example.com",
		Items: []CreateItemParams{
			{
				ProductID: "product-1",
				Quantity:  0,
			},
		},
	})

	if err == nil {
		t.Fatal("expected error for invalid quantity")
	}

	if repo.createCalled {
		t.Fatal("repository should not be called when quantity is invalid")
	}
}

func TestServiceCreateRejectsMissingCustomerEmail(t *testing.T) {
	repo := &fakeOrderRepo{}
	service := NewService(repo, nil, nil)

	_, err := service.Create(context.Background(), CreateParams{
		StoreID:       "store-1",
		CustomerEmail: "",
		Items: []CreateItemParams{
			{
				ProductID: "product-1",
				Quantity:  1,
			},
		},
	})

	if err == nil {
		t.Fatal("expected error for missing customer email")
	}

	if repo.createCalled {
		t.Fatal("repository should not be called when customer email is missing")
	}
}
