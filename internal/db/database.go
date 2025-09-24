package db

import (
	_config "bitbucket.org/edts/go-task-management/config"
	_logger "bitbucket.org/edts/go-task-management/pkg/logger"
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

var logs = _logger.GetContextLoggerf(nil)

// Database struct wraps the connection pool
type Database struct {
	Pool *pgxpool.Pool
}

// NewDatabase initializes and returns a new database connection
func NewDatabase(cfg *_config.DatabaseConfig) (*Database, error) {
	config, err := pgxpool.ParseConfig(cfg.URL)
	if err != nil {
		logs.Errorf("unable to parse database URL: %w", err)
		return nil, fmt.Errorf("unable to parse database URL: %w", err)
	}

	config.MaxConns = cfg.MaxConnections
	config.MinConns = cfg.MinConnections
	config.MaxConnIdleTime = cfg.MaxIdleTime
	config.HealthCheckPeriod = cfg.HealthCheckPeriod

	// Initialize database connection pool
	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		logs.Errorf("unable to connect to database: %v", err)
		return nil, fmt.Errorf("unable to connect to database: %w", err)
	}

	// Check connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err = pool.Ping(ctx); err != nil {
		pool.Close()
		logs.Errorf("database ping failed: %v", err)
		return nil, fmt.Errorf("database ping failed: %w", err)
	}

	logs.Info("✅ Connected to PostgreSQL successfully!")
	return &Database{Pool: pool}, nil
}

// Close closes the database connection
func (d *Database) Close() {
	if d.Pool != nil {
		d.Pool.Close()
		fmt.Println("Database connection closed.")
	}
}
