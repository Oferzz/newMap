#!/bin/bash

# Script to run integration tests for the Trip Platform API

set -e

echo "🧪 Running Trip Platform Integration Tests"
echo "========================================"

# Check if PostgreSQL is running
if ! pg_isready -h localhost -p 5432 > /dev/null 2>&1; then
    echo "❌ PostgreSQL is not running on localhost:5432"
    echo "Please start PostgreSQL before running integration tests"
    exit 1
fi

# Check if Redis is running
if ! redis-cli ping > /dev/null 2>&1; then
    echo "❌ Redis is not running on localhost:6379"
    echo "Please start Redis before running integration tests"
    exit 1
fi

# Create test database if it doesn't exist
echo "📦 Setting up test database..."
createdb -h localhost -p 5432 -U postgres trip_platform_test 2>/dev/null || true

# Run migrations on test database
echo "🔄 Running database migrations..."
export DATABASE_URL="postgres://postgres:postgres@localhost:5432/trip_platform_test?sslmode=disable"
migrate -path ../migrations -database "$DATABASE_URL" up

# Set test environment variables
export APP_ENV=test
export RUN_INTEGRATION_TESTS=true
export JWT_SECRET=test-secret-key-for-integration-tests
export REDIS_URL=redis://localhost:6379/1

# Run the tests
echo "🚀 Running integration tests..."
cd ..
go test -v -tags=integration ./cmd/server -count=1

# Clean up test database
echo "🧹 Cleaning up..."
dropdb -h localhost -p 5432 -U postgres trip_platform_test 2>/dev/null || true

echo "✅ Integration tests completed!"