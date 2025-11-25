package events

import "casebrief/internal/models"

// OrderCreatedEvent represents an event emitted when an order is created
type OrderCreatedEvent struct {
	OrderID    string
	CustomerID string
	ProductID  string
	Quantity   int
	TotalPrice float64
	Timestamp  int64
}

// ToOrder converts event to order model
func (e *OrderCreatedEvent) ToOrder() *models.Order {
	return &models.Order{
		ID:          e.OrderID,
		CustomerID: e.CustomerID,
		ProductID:  e.ProductID,
		Quantity:   e.Quantity,
		TotalPrice: e.TotalPrice,
		Status:     "created",
	}
}

