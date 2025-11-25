package db

import (
	"casebrief/internal/config"
	"context"
	"database/sql"
	"fmt"
	"time"

	"go.uber.org/zap"
)

func ConnectDB(cfg *config.Config, logger *zap.Logger) (*sql.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBSSLMode,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, err
	}

	logger.Info("Database connection established")
	return db, nil
}
