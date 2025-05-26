package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/Oferzz/newMap/apps/api/internal/config"
)

// PostgresDB wraps the sqlx database connection
type PostgresDB struct {
	*sqlx.DB
	pool *pgxpool.Pool
}

// NewPostgresDB creates a new PostgreSQL database connection
func NewPostgresDB(cfg *config.DatabaseConfig) (*PostgresDB, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Parse the connection string for pgx
	connConfig, err := pgxpool.ParseConfig(cfg.URI)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database URI: %w", err)
	}

	// Configure connection pool
	connConfig.MaxConns = int32(cfg.MaxPoolSize)
	connConfig.MinConns = int32(cfg.MinPoolSize)
	connConfig.MaxConnLifetime = time.Duration(cfg.MaxIdleTime) * time.Minute
	connConfig.MaxConnIdleTime = time.Duration(cfg.MaxIdleTime) * time.Minute

	// Create the connection pool
	pool, err := pgxpool.NewWithConfig(ctx, connConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Test the connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Create sqlx DB from pgx pool for easier query building
	connStr := stdlib.RegisterConnConfig(connConfig.ConnConfig)
	db, err := sqlx.Open("pgx", connStr)
	if err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to create sqlx connection: %w", err)
	}

	// Configure sqlx connection pool to match pgx settings
	db.SetMaxOpenConns(cfg.MaxPoolSize)
	db.SetMaxIdleConns(cfg.MinPoolSize)
	db.SetConnMaxLifetime(time.Duration(cfg.MaxIdleTime) * time.Minute)
	db.SetConnMaxIdleTime(time.Duration(cfg.MaxIdleTime) * time.Minute)

	return &PostgresDB{
		DB:   db,
		pool: pool,
	}, nil
}

// RunMigrations runs database migrations
func (db *PostgresDB) RunMigrations(migrationsPath string) error {
	driver, err := postgres.WithInstance(db.DB.DB, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create migration driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", migrationsPath),
		"postgres",
		driver,
	)
	if err != nil {
		return fmt.Errorf("failed to create migration instance: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

// CreateExtensions creates necessary PostgreSQL extensions
func (db *PostgresDB) CreateExtensions(ctx context.Context) error {
	extensions := []string{
		"CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"",
		"CREATE EXTENSION IF NOT EXISTS \"postgis\"",
		"CREATE EXTENSION IF NOT EXISTS \"pg_trgm\"",
	}

	for _, ext := range extensions {
		if _, err := db.DB.ExecContext(ctx, ext); err != nil {
			return fmt.Errorf("failed to create extension: %w", err)
		}
	}

	return nil
}

// Close closes both sqlx and pgx connections
func (db *PostgresDB) Close() error {
	db.pool.Close()
	return db.DB.Close()
}

// GetPool returns the underlying pgxpool for advanced operations
func (db *PostgresDB) GetPool() *pgxpool.Pool {
	return db.pool
}

// BeginTx starts a new transaction with context
func (db *PostgresDB) BeginTx(ctx context.Context) (*sql.Tx, error) {
	return db.DB.BeginTx(ctx, nil)
}

// Transaction executes a function within a database transaction
func (db *PostgresDB) Transaction(ctx context.Context, fn func(*sqlx.Tx) error) error {
	tx, err := db.DB.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("tx error: %w, rollback error: %v", err, rbErr)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// HealthCheck performs a health check on the database
func (db *PostgresDB) HealthCheck(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := db.DB.PingContext(ctx); err != nil {
		return fmt.Errorf("database health check failed: %w", err)
	}

	// Check if we can execute a simple query
	var result int
	if err := db.DB.GetContext(ctx, &result, "SELECT 1"); err != nil {
		return fmt.Errorf("database query health check failed: %w", err)
	}

	return nil
}