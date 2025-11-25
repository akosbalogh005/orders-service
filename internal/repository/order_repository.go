package repository

import (
	"context"
	"database/sql"
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

// GetOrderByIdempotencyKey retrieves an order by its idempotency key
func (r *OrderRepository) GetOrderByIdempotencyKey(ctx context.Context, key string) (*models.Order, error) {
	query := `
		SELECT o.id, o.customer_id, o.product_id, o.quantity, o.total_price, o.status, o.order_time, o.created_at, o.updated_at
		FROM orders o
		INNER JOIN idempotency_keys ik ON o.id = ik.order_id
		WHERE ik.key = $1
	`

	order := &models.Order{}
	err := r.db.QueryRowContext(ctx, query, key).Scan(
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
		r.logger.Error("Failed to get order by idempotency key",
			zap.Error(err),
			zap.String("idempotency_key", key),
		)
		return nil, err
	}

	return order, nil
}

// StoreIdempotencyKey stores an idempotency key mapping
func (r *OrderRepository) StoreIdempotencyKey(ctx context.Context, key string, orderID string) error {
	query := `
		INSERT INTO idempotency_keys (key, order_id, created_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (key) DO NOTHING
	`

	_, err := r.db.ExecContext(ctx, query, key, orderID, time.Now())
	if err != nil {
		r.logger.Error("Failed to store idempotency key",
			zap.Error(err),
			zap.String("idempotency_key", key),
			zap.String("order_id", orderID),
		)
		return err
	}

	return nil
}
