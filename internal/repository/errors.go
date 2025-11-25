package repository

import "errors"

var (
	// ErrOrderNotFound is returned when an order is not found
	ErrOrderNotFound = errors.New("order not found")
	// ErrIdempotencyNotFound is returned when an idempotency record is not found or expired
	ErrIdempotencyNotFound = errors.New("idempotency record not found")
)

