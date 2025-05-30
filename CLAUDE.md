# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview
This is a collaborative trip planning platform with real-time features, role-based permissions, and map-based interactions. The platform follows a "freemium" model where core features (save locations, create routes, search) are available without authentication using browser cache storage, while collaboration features require authentication for cloud storage and sharing. The architecture follows a microservices-ready monolith approach with a Go backend and React frontend.

## Tech Stack
- **Backend**: Go 1.23+ with Gin framework
- **Frontend**: React 18+ with TypeScript, Redux Toolkit, Mapbox GL JS
- **Database**: PostgreSQL with PostGIS (primary), MongoDB (legacy), Redis (caching)
- **Real-time**: Socket.io
- **Deployment**: Render.com with Cloudflare CDN

## Development Commands

### Backend (Go API)
```bash
# Navigate to API directory
cd apps/api

# Run tests
go test -v -cover ./...
go test -race -coverprofile=coverage.txt ./...

# Run linting (if golangci-lint is installed)
golangci-lint run

# Run the server locally
go run cmd/server/main.go

# Build the application
go build -o bin/server cmd/server/main.go
```

### Frontend (React Web App)
```bash
# Navigate to web app directory
cd apps/web

# Install dependencies
npm install

# Run development server
npm run dev

# Run tests
npm test
npm run test:coverage  # With coverage report
npm run test:ui  # UI for tests

# Build for production
npm run build

# Run linting
npm run lint

# Run type checking
npm run typecheck
```

### Docker Development
```bash
# Run full stack locally
docker-compose up

# Run in production mode
docker-compose -f docker-compose.prod.yml up

# Rebuild containers
docker-compose build

# Use deployment script for convenience
./deploy.sh start    # Start services
./deploy.sh stop     # Stop services
./deploy.sh logs     # View logs
./deploy.sh migrate  # Run database migrations
./deploy.sh backup   # Create database backup
```

## Architecture Overview

### Monorepo Structure
The project uses a monorepo structure with apps (api, web) and shared packages. Key directories:
- `apps/api/` - Go backend with domain-driven design
- `apps/web/` - React frontend with feature-based organization
- `packages/` - Shared TypeScript types, validators, and constants

### Backend Architecture
- **Service Layer Pattern**: Each domain (trips, places, users, etc.) has handler, service, repository, and models
- **RBAC Implementation**: Comprehensive role-based access control with admin, editor, and viewer roles
- **Event-Driven**: Publishes events for real-time updates
- **Caching Strategy**: Redis caching with TTL management for performance

### Frontend Architecture
- **Feature-Based Organization**: Components organized by feature (trips, places, suggestions)
- **Redux Toolkit**: State management with async thunks
- **Dual Storage Strategy**: Local storage for guest users, cloud storage for authenticated users
- **Real-time Hooks**: Custom hooks for WebSocket connections
- **Mapbox Integration**: Map-based interactions with performance optimizations

### Database Design
- **PostgreSQL Tables**: Users, Trips, Places, Suggestions, Media with PostGIS for geospatial data
- **MongoDB Collections**: Legacy support for existing data
- **Geospatial Indexes**: For location-based queries using PostGIS
- **Text Search**: Full-text search on trips and places
- **Hierarchical Data**: Places can have parent-child relationships

## Key Development Patterns

### API Response Format
All API responses follow this structure:
```json
{
  "success": true,
  "data": {},
  "meta": {
    "page": 1,
    "limit": 20,
    "total": 100,
    "hasMore": true
  },
  "error": null
}
```

### Permission Checking
Always check permissions before operations:
```go
RequirePermission(permission Permission) gin.HandlerFunc
```

### Real-time Events
Events follow the pattern: `resource.action` (e.g., `trip.created`, `place.updated`)

### Caching Keys
Redis cache keys follow the pattern: `resource:id:subresource`

## Testing Requirements
- All new features must have comprehensive tests
- Backend: Use Go's built-in testing with coverage reports
- Frontend: Vitest for unit tests, React Testing Library for component tests
- Integration tests for API endpoints
- E2E tests for critical user flows

## Security Considerations
- JWT authentication with refresh tokens for authenticated users
- Guest mode with secure local storage (no sensitive data)
- Role-based access control on protected endpoints
- Public endpoints for search and place discovery
- Input validation and sanitization on all inputs
- Rate limiting on both public and protected API endpoints
- Secure media upload with virus scanning

## Performance Guidelines
- Use database projections to limit returned fields
- Implement cursor-based pagination for large datasets
- Cache frequently accessed data with appropriate TTLs
- Use viewport-based loading for map markers
- Optimize bundle size with code splitting

## Deployment Process
- CI/CD via GitHub Actions
- Automated testing on all PRs
- Deploy to Render.com on main branch
- Environment-specific configurations
- Database migrations handled separately

## Local Environment Setup

### Required Environment Variables
```bash
# Backend (.env in apps/api/)
DATABASE_URL=postgresql://user:pass@localhost:5432/newMap
REDIS_URL=redis://localhost:6379
JWT_SECRET=your-secure-secret-key
MAPBOX_API_KEY=your-mapbox-api-key

# Frontend (.env in apps/web/)
VITE_API_URL=http://localhost:8080/api/v1
VITE_WS_URL=ws://localhost:8080
VITE_MAPBOX_TOKEN=your-mapbox-api-key
```

### Single Test Execution
```bash
# Backend - run specific test
cd apps/api
go test -v ./internal/domain/trips -run TestCreateTrip

# Frontend - run specific test file
cd apps/web
npm test -- SearchBar.test.tsx

# Frontend - run tests in watch mode
npm test -- --watch
```

## Additional Guidelines
- Database migrations are in `apps/api/migrations/` - use `./deploy.sh migrate` or `go run cmd/server/main.go migrate`
- When adding new API endpoints, follow the domain-driven structure in `internal/domain/`
- All API responses use the standardized response format via `pkg/response`
- Frontend state management follows Redux Toolkit patterns with thunks in `store/thunks/`
- Implement dual storage pattern: local storage for guest users, API calls for authenticated users
- Real-time features use WebSocket connections managed by `useWebSocket` hook
- Public endpoints should not require authentication (places search, health check)
- Collections and trip sharing require authentication
- If there is something you dont know, dont assume it, go read it in the render docs

## Authentication Strategy
- **Guest Mode**: Core features work without authentication using browser localStorage/IndexedDB
- **Authenticated Mode**: Additional collaboration features with cloud storage and real-time sync
- **Migration Path**: Seamless upgrade from guest to authenticated with local data migration
- See `AUTHENTICATION_STRATEGY.md` for detailed implementation guidelines

## Puppeteer Memory
- When using puppeteer, go to https://newmap-fe.onrender.com/ for the website url

## Development Memories
- create every feature on main branch, push everything after finishing implementing