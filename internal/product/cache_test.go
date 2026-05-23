package product

import (
	"testing"
)

func TestGetCacheKey(t *testing.T) {
	tests := []struct {
		name       string
		storeID    string
		activeOnly bool
		expected   string
	}{
		{
			name:       "active products",
			storeID:    "store-123",
			activeOnly: true,
			expected:   "store:store-123:products:active:true",
		},
		{
			name:       "all products",
			storeID:    "store-123",
			activeOnly: false,
			expected:   "store:store-123:products:active:false",
		},
		{
			name:       "different store active",
			storeID:    "store-456",
			activeOnly: true,
			expected:   "store:store-456:products:active:true",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock service to access the getCacheKey method
			s := &Service{}
			result := s.getCacheKey(tt.storeID, tt.activeOnly)

			if result != tt.expected {
				t.Errorf("getCacheKey(%s, %v) = %s; want %s", tt.storeID, tt.activeOnly, result, tt.expected)
			}
		})
	}
}
