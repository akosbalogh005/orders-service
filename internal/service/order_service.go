package service

import (
	"context"
	"time"

	"casebrief/internal/events"
	"casebrief/internal/models"
	"casebrief/internal/repository"

	"go.uber.org/zap"
)

// OrderService handles business logic for orders
type OrderService struct {
	repo      *repository.OrderRepository
	eventChan chan *events.OrderCreatedEvent
	logger    *zap.Logger
}

// NewOrderService creates a new order service
func NewOrderService(repo *repository.OrderRepository, eventChan chan *events.OrderCreatedEvent, logger *zap.Logger) *OrderService {
	return &OrderService{
		repo:      repo,
		eventChan: eventChan,
		logger:    logger,
	}
}

// CreateOrder creates a new order and emits an event
func (s *OrderService) CreateOrder(ctx context.Context, req *models.CreateOrderRequest) (*models.Order, error) {
	// Check idempotency
	existingOrder, err := s.repo.GetOrderByIdempotencyKey(ctx, req.IdempotencyKey)
	if err == nil && existingOrder != nil {
		s.logger.Info("Order already exists for idempotency key",
			zap.String("idempotency_key", req.IdempotencyKey),
			zap.String("order_id", existingOrder.ID),
		)
		return existingOrder, nil
	}

	// Create order
	order := &models.Order{
		CustomerID: req.CustomerID,
		ProductID:  req.ProductID,
		Quantity:   req.Quantity,
		TotalPrice: req.TotalPrice,
		OrderTime:  req.OrderTime,
	}

	if err := s.repo.CreateOrder(ctx, order); err != nil {
		return nil, err
	}

	// Store idempotency key
	if err := s.repo.StoreIdempotencyKey(ctx, req.IdempotencyKey, order.ID); err != nil {
		s.logger.Warn("Failed to store idempotency key",
			zap.Error(err),
			zap.String("idempotency_key", req.IdempotencyKey),
		)
		// Don't fail the request if idempotency key storage fails
	}

	// Emit event
	event := &events.OrderCreatedEvent{
		OrderID:    order.ID,
		CustomerID: order.CustomerID,
		ProductID:  order.ProductID,
		Quantity:   order.Quantity,
		TotalPrice: order.TotalPrice,
		Timestamp:  time.Now().Unix(),
	}

	select {
	case s.eventChan <- event:
		s.logger.Info("OrderCreated event emitted",
			zap.String("order_id", order.ID),
		)
	case <-ctx.Done():
		s.logger.Warn("Context cancelled before event could be emitted",
			zap.String("order_id", order.ID),
		)
	default:
		s.logger.Warn("Event channel full, event not emitted",
			zap.String("order_id", order.ID),
		)
	}

	return order, nil
}

// GetOrderByID retrieves an order by ID
func (s *OrderService) GetOrderByID(ctx context.Context, id string) (*models.Order, error) {
	return s.repo.GetOrderByID(ctx, id)
}
