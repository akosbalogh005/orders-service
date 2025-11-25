package repository

import "errors"

var (
	// ErrOrderNotFound is returned when an order is not found
	ErrOrderNotFound = errors.New("order not found")
)

