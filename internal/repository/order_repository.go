package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"casebrief/internal/models"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

// OrderRepository handles database operations for orders
type OrderRepository struct {
	db     *sql.DB
	logger *zap.Logger
}

// NewOrderRepository creates a new order repository
func NewOrderRepository(db *sql.DB, logger *zap.Logger) *OrderRepository {
	return &OrderRepository{
		db:     db,
		logger: logger,
	}
}

// CreateOrder creates a new order in the database
func (r *OrderRepository) CreateOrder(ctx context.Context, order *models.Order) error {
	query := `
		INSERT INTO orders (id, customer_id, product_id, quantity, total_price, status, order_time, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	now := time.Now()
	order.ID = uuid.New().String()
	if order.OrderTime.IsZero() {
		order.OrderTime = now
	}
	order.CreatedAt = now
	order.UpdatedAt = now
	order.Status = "created"

	_, err := r.db.ExecContext(ctx, query,
		order.ID,
		order.CustomerID,
		order.ProductID,
		order.Quantity,
		order.TotalPrice,
		order.Status,
		order.OrderTime,
		order.CreatedAt,
		order.UpdatedAt,
	)

	if err != nil {
		r.logger.Error("Failed to create order",
			zap.Error(err),
			zap.String("customer_id", order.CustomerID),
		)
		return err
	}

	r.logger.Info("Order created successfully",
		zap.String("order_id", order.ID),
	)
	return nil
}

// GetOrderByID retrieves an order by its ID
func (r *OrderRepository) GetOrderByID(ctx context.Context, id string) (*models.Order, error) {
	query := `
		SELECT id, customer_id, product_id, quantity, total_price, status, order_time, created_at, updated_at
		FROM orders
		WHERE id = $1
	`

	order := &models.Order{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&order.ID,
		&order.CustomerID,
		&order.ProductID,
		&order.Quantity,
		&order.TotalPrice,
		&order.Status,
		&order.OrderTime,
		&order.CreatedAt,
		&order.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, ErrOrderNotFound
	}

	if err != nil {
		r.logger.Error("Failed to get order by ID",
			zap.Error(err),
			zap.String("order_id", id),
		)
		return nil, err
	}

	return order, nil
}

// GetIdempotencyResponse retrieves a saved response by endpoint and idempotency key if still valid
func (r *OrderRepository) GetIdempotencyResponse(ctx context.Context, endpointName, endpointScheme, key string) ([]byte, error) {
	query := `
		SELECT response
		FROM idempotency_keys
		WHERE endpoint_name = $1 AND endpoint_scheme = $2 AND key = $3 AND valid_to > NOW()
	`

	var response []byte
	err := r.db.QueryRowContext(ctx, query, endpointName, endpointScheme, key).Scan(&response)

	if err == sql.ErrNoRows {
		return nil, ErrIdempotencyNotFound
	}

	if err != nil {
		r.logger.Error("Failed to get idempotency response",
			zap.Error(err),
			zap.String("endpoint_name", endpointName),
			zap.String("endpoint_scheme", endpointScheme),
			zap.String("idempotency_key", key),
		)
		return nil, err
	}

	return response, nil
}

// StoreIdempotencyResponse stores an idempotency key with endpoint info and response
func (r *OrderRepository) StoreIdempotencyResponse(ctx context.Context, endpointName, endpointScheme, key string, response interface{}, validityDuration time.Duration) error {
	responseJSON, err := json.Marshal(response)
	if err != nil {
		r.logger.Error("Failed to marshal response for idempotency",
			zap.Error(err),
		)
		return err
	}

	validTo := time.Now().Add(validityDuration)
	query := `
		INSERT INTO idempotency_keys (endpoint_name, endpoint_scheme, key, response, valid_to, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (endpoint_name, endpoint_scheme, key) DO UPDATE
		SET response = EXCLUDED.response, valid_to = EXCLUDED.valid_to
	`

	_, err = r.db.ExecContext(ctx, query, endpointName, endpointScheme, key, responseJSON, validTo, time.Now())
	if err != nil {
		r.logger.Error("Failed to store idempotency response",
			zap.Error(err),
			zap.String("endpoint_name", endpointName),
			zap.String("endpoint_scheme", endpointScheme),
			zap.String("idempotency_key", key),
		)
		return err
	}

	return nil
}
