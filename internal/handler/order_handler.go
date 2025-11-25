package handler

import (
	"net/http"

	"casebrief/internal/models"
	"casebrief/internal/repository"
	"casebrief/internal/service"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// OrderHandler handles HTTP requests for orders
type OrderHandler struct {
	service *service.OrderService
	logger  *zap.Logger
}

// NewOrderHandler creates a new order handler
func NewOrderHandler(service *service.OrderService, logger *zap.Logger) *OrderHandler {
	return &OrderHandler{
		service: service,
		logger:  logger,
	}
}

// CreateOrder handles POST /orders
// @Summary Create a new order
// @Description Create a new order with idempotency support
// @Tags orders
// @Accept json
// @Produce json
// @Param order body models.CreateOrderRequest true "Order creation request"
// @Success 201 {object} models.Order
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /orders [post]
func (h *OrderHandler) CreateOrder(c *gin.Context) {
	ctx := c.Request.Context()
	span := trace.SpanFromContext(ctx)
	defer span.End()

	var req models.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid request body",
			zap.Error(err),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	// Extract endpoint name and scheme for idempotency
	endpointName := c.Request.URL.Path // e.g., "/orders"
	endpointScheme := c.Request.Method // e.g., "POST"

	order, err := h.service.CreateOrder(ctx, endpointName, endpointScheme, &req)
	if err != nil {
		h.logger.Error("Failed to create order",
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create order"})
		return
	}

	h.logger.Info("Order created successfully",
		zap.String("order_id", order.ID),
		zap.String("customer_id", order.CustomerID),
	)
	c.JSON(http.StatusCreated, order)
}

// GetOrderByID handles GET /orders/{id}
// @Summary Get order by ID
// @Description Retrieve an order by its ID
// @Tags orders
// @Produce json
// @Param id path string true "Order ID"
// @Success 200 {object} models.Order
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /orders/{id} [get]
func (h *OrderHandler) GetOrderByID(c *gin.Context) {
	ctx := c.Request.Context()
	span := trace.SpanFromContext(ctx)
	defer span.End()

	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Order ID is required"})
		return
	}

	order, err := h.service.GetOrderByID(ctx, id)
	if err != nil {
		if err == repository.ErrOrderNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
			return
		}
		h.logger.Error("Failed to get order",
			zap.Error(err),
			zap.String("order_id", id),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve order"})
		return
	}

	c.JSON(http.StatusOK, order)
}
