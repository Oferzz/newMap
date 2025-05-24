# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview
This is a collaborative trip planning platform with real-time features, role-based permissions, and map-based interactions. The architecture follows a microservices-ready monolith approach with a Go backend and React frontend.

## Tech Stack
- **Backend**: Go 1.21+ with Gin framework
- **Frontend**: React 18+ with TypeScript, Redux Toolkit, Mapbox GL JS
- **Database**: MongoDB Atlas (primary), Redis (caching)
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
npm run test:ci  # For CI environment

# Build for production
npm run build

# Run linting
npm run lint

# Run type checking (if configured)
npm run typecheck
```

### Docker Development
```bash
# Run full stack locally
docker-compose up

# Rebuild containers
docker-compose build
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
- **Real-time Hooks**: Custom hooks for WebSocket connections
- **Mapbox Integration**: Map-based interactions with performance optimizations

### Database Design
- **MongoDB Collections**: Users, Trips, Places, Suggestions, Media
- **Geospatial Indexes**: For location-based queries
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
- Frontend: Jest/React Testing Library for component tests
- Integration tests for API endpoints
- E2E tests for critical user flows

## Security Considerations
- JWT authentication with refresh tokens
- Role-based access control on all endpoints
- Input validation and sanitization
- Rate limiting on API endpoints
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