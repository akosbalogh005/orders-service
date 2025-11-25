package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"casebrief/docs"
	"casebrief/internal/config"
	"casebrief/internal/db"
	"casebrief/internal/events"
	"casebrief/internal/handler"
	"casebrief/internal/logger"
	"casebrief/internal/middleware"
	"casebrief/internal/repository"
	"casebrief/internal/service"
	"casebrief/internal/tracing"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.uber.org/zap"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Initialize logger
	appLogger, err := logger.NewLogger(cfg)
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}
	defer appLogger.Sync()

	appLogger.Info("Starting Orders microservice",
		zap.String("port", cfg.ServerPort),
	)

	// Initialize tracing
	var shutdownTracing func()
	if cfg.OTelEnabled {
		shutdownTracing, err = tracing.InitTracing("orders-service", appLogger)
		if err != nil {
			appLogger.Warn("Failed to initialize tracing", zap.Error(err))
		} else {
			defer shutdownTracing()
		}
	}

	// Connect to database
	db, err := db.ConnectDB(cfg, appLogger)
	if err != nil {
		appLogger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	// Create event channel
	eventChan := make(chan *events.OrderCreatedEvent, cfg.EventQueueSize)

	// Initialize repository
	orderRepo := repository.NewOrderRepository(db, appLogger)

	// Initialize service
	orderService := service.NewOrderService(orderRepo, eventChan, appLogger)

	// Initialize handlers
	orderHandler := handler.NewOrderHandler(orderService, appLogger)
	healthHandler := handler.NewHealthHandler()

	// Create event worker
	worker := events.NewWorker(eventChan, appLogger)

	// Start event worker
	workerCtx, workerCancel := context.WithCancel(context.Background())
	defer workerCancel()
	go worker.Start(workerCtx)

	// Setup router
	router := setupRouter(cfg, orderHandler, healthHandler, appLogger)

	// Create HTTP server
	srv := &http.Server{
		Addr:    ":" + cfg.ServerPort,
		Handler: router,
	}

	// Start server in goroutine
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			appLogger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	appLogger.Info("Server started successfully",
		zap.String("port", cfg.ServerPort),
	)

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	appLogger.Info("Shutting down server...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Stop event worker
	worker.Stop()
	workerCancel()

	// Shutdown HTTP server
	if err := srv.Shutdown(ctx); err != nil {
		appLogger.Error("Server forced to shutdown", zap.Error(err))
	}

	appLogger.Info("Server exited")
}

func setupRouter(cfg *config.Config, orderHandler *handler.OrderHandler, healthHandler *handler.HealthHandler, logger *zap.Logger) *gin.Engine {
	router := gin.New()

	// Use zap logger and recovery middleware
	router.Use(middleware.ZapLogger(logger))
	router.Use(middleware.ZapRecovery(logger))

	// Initialize Swagger docs
	docs.SwaggerInfo.Host = cfg.Hostname + ":" + cfg.ServerPort

	// Add OpenTelemetry middleware if enabled
	if cfg.OTelEnabled {
		router.Use(otelgin.Middleware("orders-service"))
	}

	// Health check
	router.GET("/healthz", healthHandler.HealthCheck)

	// Swagger documentation
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API routes
	router.POST("/orders", orderHandler.CreateOrder)
	router.GET("/orders/:id", orderHandler.GetOrderByID)

	return router
}
