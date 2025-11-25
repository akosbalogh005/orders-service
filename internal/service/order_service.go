package service

import (
	"context"
	"encoding/json"
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
func (s *OrderService) CreateOrder(ctx context.Context, endpointName, endpointScheme string, req *models.CreateOrderRequest) (*models.Order, error) {
	// Check idempotency - if valid record exists, return saved response
	savedResponse, err := s.repo.GetIdempotencyResponse(ctx, endpointName, endpointScheme, req.IdempotencyKey)
	if err == nil && savedResponse != nil {
		s.logger.Info("Idempotent request detected, returning saved response",
			zap.String("endpoint_name", endpointName),
			zap.String("endpoint_scheme", endpointScheme),
			zap.String("idempotency_key", req.IdempotencyKey),
		)

		var order models.Order
		if err := json.Unmarshal(savedResponse, &order); err != nil {
			s.logger.Warn("Failed to unmarshal saved idempotency response, proceeding with new request",
				zap.Error(err),
			)
			// Continue with normal flow if unmarshaling fails
		} else {
			return &order, nil
		}
	} else if err != nil && err != repository.ErrIdempotencyNotFound {
		s.logger.Warn("Error checking idempotency, proceeding with new request",
			zap.Error(err),
		)
		// Continue with normal flow if there's an error (but not "not found")
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

	// Store idempotency response (default validity: 10 minutes)
	validityDuration := 10 * time.Minute
	if err := s.repo.StoreIdempotencyResponse(ctx, endpointName, endpointScheme, req.IdempotencyKey, order, validityDuration); err != nil {
		s.logger.Warn("Failed to store idempotency response",
			zap.Error(err),
			zap.String("endpoint_name", endpointName),
			zap.String("endpoint_scheme", endpointScheme),
			zap.String("idempotency_key", req.IdempotencyKey),
		)
		// Don't fail the request if idempotency storage fails
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
