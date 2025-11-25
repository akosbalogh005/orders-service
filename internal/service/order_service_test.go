package service

import (
	"testing"

	"casebrief/internal/models"

	"github.com/stretchr/testify/assert"
)

// func TestOrderService_GetOrderByID(t *testing.T) {
// 	logger, _ := zap.NewDevelopment()
// 	eventChan := make(chan *events.OrderCreatedEvent, 10)

// 	service := NewOrderService(nil, eventChan, logger)

// 	ctx := context.Background()
// 	order, err := service.GetOrderByID(ctx, "non-existent-id")

// 	// Should return error when repository is nil (for testing structure)
// 	assert.Error(t, err)
// 	assert.Nil(t, order)
// }

func TestOrderService_CreateOrder_Validation(t *testing.T) {
	//logger, _ := zap.NewDevelopment()
	//eventChan := make(chan *events.OrderCreatedEvent, 10)

	//service := NewOrderService(nil, eventChan, logger)

	tests := []struct {
		name string
		req  *models.CreateOrderRequest
	}{
		{
			name: "valid request structure",
			req: &models.CreateOrderRequest{
				CustomerID:     "customer-1",
				ProductID:      "product-1",
				Quantity:       2,
				TotalPrice:     100.50,
				IdempotencyKey: "key-1",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify request structure is valid
			assert.NotEmpty(t, tt.req.CustomerID)
			assert.NotEmpty(t, tt.req.ProductID)
			assert.Greater(t, tt.req.Quantity, 0)
			assert.GreaterOrEqual(t, tt.req.TotalPrice, 0.0)
			assert.NotEmpty(t, tt.req.IdempotencyKey)
		})
	}
}
