# Testing Guide for Trip Platform API

This document describes how to run tests for the Trip Platform API.

## Prerequisites

- Go 1.21 or later
- PostgreSQL 16 with PostGIS extension
- Redis 7.0 or later
- golang-migrate CLI tool (for running migrations)

## Test Structure

The API has three levels of testing:

### 1. Unit Tests
Unit tests test individual functions and methods in isolation using mocks.

**Location**: `*_test.go` files alongside the source code
**Examples**:
- `internal/domain/users/service_test.go`
- `internal/domain/trips/repository_pg_test.go`
- `internal/utils/jwt_test.go`

### 2. Handler Tests
Handler tests test the HTTP handlers with mocked services.

**Location**: `internal/domain/*/handler_test.go`
**Examples**:
- `internal/domain/users/handler_test.go`
- `internal/domain/trips/handler_test.go`
- `internal/domain/places/handler_test.go`

### 3. Integration Tests
Integration tests test the full API with real database connections.

**Location**: `cmd/server/main_test.go`

## Running Tests

### Run All Unit Tests
```bash
cd apps/api
go test -v ./...
```

### Run Tests with Coverage
```bash
cd apps/api
go test -v -cover ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

### Run Tests for Specific Package
```bash
cd apps/api
go test -v ./internal/domain/users/...
go test -v ./internal/domain/trips/...
```

### Run Integration Tests
```bash
cd apps/api
./scripts/run-integration-tests.sh
```

Or manually:
```bash
# Start PostgreSQL and Redis
# Create test database
createdb trip_platform_test

# Run migrations
export DATABASE_URL="postgres://localhost/trip_platform_test?sslmode=disable"
migrate -path migrations -database "$DATABASE_URL" up

# Run integration tests
export RUN_INTEGRATION_TESTS=true
go test -v -tags=integration ./cmd/server

# Clean up
dropdb trip_platform_test
```

### Run Tests with Race Detection
```bash
cd apps/api
go test -race ./...
```

## Test Database Setup

For integration tests, you need a test PostgreSQL database:

```sql
-- Create test database
CREATE DATABASE trip_platform_test;

-- Connect to test database
\c trip_platform_test;

-- Enable PostGIS extension
CREATE EXTENSION IF NOT EXISTS postgis;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
```

## Mocking

We use [testify/mock](https://github.com/stretchr/testify) for mocking dependencies.

Example mock:
```go
type MockService struct {
    mock.Mock
}

func (m *MockService) Create(ctx context.Context, input *CreateUserInput) (*User, error) {
    args := m.Called(ctx, input)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*User), args.Error(1)
}
```

## Writing Tests

### Unit Test Example
```go
func TestService_Create(t *testing.T) {
    mockRepo := new(MockRepository)
    service := NewService(mockRepo)

    input := &CreateUserInput{
        Email:    "test@example.com",
        Username: "testuser",
    }

    mockRepo.On("Create", mock.Anything, mock.Anything).Return(nil)

    user, err := service.Create(context.Background(), input)
    
    assert.NoError(t, err)
    assert.NotNil(t, user)
    mockRepo.AssertExpectations(t)
}
```

### Handler Test Example
```go
func TestHandler_GetUser(t *testing.T) {
    gin.SetMode(gin.TestMode)
    
    mockService := new(MockService)
    handler := NewHandler(mockService)
    
    router := gin.New()
    router.GET("/users/:id", handler.GetUser)
    
    user := &User{ID: "123", Email: "test@example.com"}
    mockService.On("GetByID", mock.Anything, "123").Return(user, nil)
    
    req := httptest.NewRequest("GET", "/users/123", nil)
    rec := httptest.NewRecorder()
    
    router.ServeHTTP(rec, req)
    
    assert.Equal(t, http.StatusOK, rec.Code)
    mockService.AssertExpectations(t)
}
```

## Continuous Integration

Tests are automatically run on every pull request via GitHub Actions. See `.github/workflows/ci.yml` for the CI configuration.

## Test Data

Test data should be:
- Self-contained within each test
- Cleaned up after test completion
- Use unique identifiers to avoid conflicts

## Troubleshooting

### PostgreSQL Connection Issues
- Ensure PostgreSQL is running: `pg_isready`
- Check connection string in environment variables
- Verify user permissions for test database

### Redis Connection Issues
- Ensure Redis is running: `redis-cli ping`
- Check Redis URL in environment variables
- Use different Redis database for tests (e.g., database 1)

### Test Failures
- Run tests with `-v` flag for verbose output
- Check test logs for detailed error messages
- Ensure all dependencies are properly mocked
- Verify database migrations are up to date