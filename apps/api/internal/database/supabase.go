package database

import (
	"context"
	"fmt"
	"log"

	"github.com/Oferzz/newMap/apps/api/internal/config"
)

type SupabaseDB struct {
	*PostgresDB
	projectURL string
	serviceKey string
}

// NewSupabaseDB creates a new Supabase database connection
func NewSupabaseDB(url, serviceKey string) (*SupabaseDB, error) {
	// Parse Supabase URL to get database connection string
	dbURL := convertSupabaseURLToPostgresURL(url, serviceKey)
	
	// Create a config object for PostgresDB
	cfg := &config.DatabaseConfig{
		URI:            dbURL,
		Name:           "postgres",
		MaxPoolSize:    100,
		MinPoolSize:    10,
		MaxIdleTime:    10,
		MigrationsPath: "./migrations",
		SSLMode:        "require",
	}
	
	// Create PostgreSQL connection using existing PostgresDB
	pgDB, err := NewPostgresDB(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create postgres connection: %w", err)
	}
	
	return &SupabaseDB{
		PostgresDB: pgDB,
		projectURL: url,
		serviceKey: serviceKey,
	}, nil
}

// convertSupabaseURLToPostgresURL converts a Supabase project URL to a PostgreSQL connection string
func convertSupabaseURLToPostgresURL(projectURL, serviceKey string) string {
	// Extract project reference from URL (e.g., https://xrzjkhivkbcjdfirunyz.supabase.co)
	var projectRef string
	if _, err := fmt.Sscanf(projectURL, "https://%s.supabase.co", &projectRef); err != nil {
		log.Printf("Failed to parse Supabase URL: %v", err)
		// Fallback to using the known project reference
		projectRef = "xrzjkhivkbcjdfirunyz"
	}
	
	log.Printf("Using project reference: %s", projectRef)
	
	// Try connection pooler first (should work better on Render)
	return fmt.Sprintf(
		"postgresql://postgres.%s:%s@aws-0-us-east-1.pooler.supabase.com:5432/postgres?sslmode=require",
		projectRef,
		serviceKey,
	)
}

// GetServiceKey returns the service key for Supabase API calls
func (db *SupabaseDB) GetServiceKey() string {
	return db.serviceKey
}

// GetProjectURL returns the project URL
func (db *SupabaseDB) GetProjectURL() string {
	return db.projectURL
}

// EnableRLS enables Row Level Security on a table
func (db *SupabaseDB) EnableRLS(ctx context.Context, tableName string) error {
	query := fmt.Sprintf("ALTER TABLE %s ENABLE ROW LEVEL SECURITY", tableName)
	_, err := db.pool.Exec(ctx, query)
	return err
}

// CreatePolicy creates a Row Level Security policy
func (db *SupabaseDB) CreatePolicy(ctx context.Context, policyName, tableName, operation, expression string) error {
	query := fmt.Sprintf(
		"CREATE POLICY %s ON %s FOR %s USING (%s)",
		policyName,
		tableName,
		operation,
		expression,
	)
	_, err := db.pool.Exec(ctx, query)
	return err
}

// HealthCheck performs a health check on the Supabase database
func (db *SupabaseDB) HealthCheck(ctx context.Context) error {
	// First check the PostgreSQL connection
	if err := db.PostgresDB.HealthCheck(ctx); err != nil {
		return err
	}
	
	// Additional Supabase-specific checks could go here
	// For example, checking if RLS is properly configured
	
	return nil
}