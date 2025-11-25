package models

import (
	"time"
)

// Order represents an order in the system
type Order struct {
	ID         string    `json:"id" db:"id"`
	CustomerID string    `json:"customer_id" db:"customer_id"`
	ProductID  string    `json:"product_id" db:"product_id"`
	Quantity   int       `json:"quantity" db:"quantity"`
	TotalPrice float64   `json:"total_price" db:"total_price"`
	Status     string    `json:"status" db:"status"`
	OrderTime  time.Time `json:"order_time" db:"order_time"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
}

// CreateOrderRequest represents the request to create an order
type CreateOrderRequest struct {
	CustomerID     string    `json:"customer_id" binding:"required"`
	ProductID      string    `json:"product_id" binding:"required"`
	Quantity       int       `json:"quantity" binding:"required,min=1"`
	TotalPrice     float64   `json:"total_price" binding:"required,min=0"`
	OrderTime      time.Time `json:"order_time,omitempty" binding:"required`
	IdempotencyKey string    `json:"idempotency_key" binding:"required"`
}
