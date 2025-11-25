package events

import (
	"context"
	"time"

	"go.uber.org/zap"
)

// Worker processes events from the event channel
type Worker struct {
	eventChan chan *OrderCreatedEvent
	logger    *zap.Logger
	stopChan  chan struct{}
}

// NewWorker creates a new event worker
func NewWorker(eventChan chan *OrderCreatedEvent, logger *zap.Logger) *Worker {
	return &Worker{
		eventChan: eventChan,
		logger:    logger,
		stopChan:  make(chan struct{}),
	}
}

// Start starts the worker to process events
func (w *Worker) Start(ctx context.Context) {
	w.logger.Info("Event worker started")
	
	for {
		select {
		case event := <-w.eventChan:
			w.processEvent(ctx, event)
		case <-ctx.Done():
			w.logger.Info("Event worker stopping due to context cancellation")
			return
		case <-w.stopChan:
			w.logger.Info("Event worker stopping")
			return
		}
	}
}

// Stop stops the worker
func (w *Worker) Stop() {
	close(w.stopChan)
}

// processEvent processes a single OrderCreated event
func (w *Worker) processEvent(ctx context.Context, event *OrderCreatedEvent) {
	w.logger.Info("Processing OrderCreated event",
		zap.String("order_id", event.OrderID),
		zap.String("customer_id", event.CustomerID),
		zap.Int("quantity", event.Quantity),
		zap.Float64("total_price", event.TotalPrice),
	)

	// Simulate some processing work
	time.Sleep(100 * time.Millisecond)

	// In a real system, this would:
	// - Send notifications
	// - Update inventory
	// - Trigger downstream services
	// - etc.

	w.logger.Info("OrderCreated event processed successfully",
		zap.String("order_id", event.OrderID),
	)
}

