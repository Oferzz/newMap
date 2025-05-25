# Trip Planning Platform - Render.com Deployment Architecture

## Table of Contents
1. [Updated System Overview](#updated-system-overview)
2. [Render Services Architecture](#render-services-architecture)
3. [Database Design with PostgreSQL](#database-design-with-postgresql)
4. [Redis Cache Implementation](#redis-cache-implementation)
5. [Media Storage Strategy](#media-storage-strategy)
6. [Deployment Configuration](#deployment-configuration)
7. [Environment Setup](#environment-setup)
8. [Cost Optimization](#cost-optimization)
9. [Migration from MongoDB to PostgreSQL](#migration-from-mongodb-to-postgresql)

## Updated System Overview

### Architecture Changes for Render
- **PostgreSQL** instead of MongoDB Atlas for primary database
- **Render Key-Value** (Redis-compatible) for caching
- **Render Persistent Disks** for media storage with CDN integration
- **Render Web Services** for both API and static site hosting
- **Built-in autoscaling** and zero-downtime deploys

### Updated Tech Stack
```yaml
Frontend:
  - React 18+ with TypeScript
  - Mapbox GL JS v3
  - Redux Toolkit
  - Socket.io client
  - Tailwind CSS
  - Vite build tool

Backend:
  - Go 1.21+ with Gin framework
  - PostgreSQL (Render Postgres)
  - Redis (Render Key-Value)
  - Persistent Disk + CDN for media
  - WebSockets via Socket.io
  - JWT authentication

Infrastructure:
  - Render.com for all services
  - Cloudflare CDN integration
  - GitHub Actions CI/CD
  - Sentry error tracking
  - Built-in Render metrics
```

## Render Services Architecture

### System Architecture on Render
```
┌─────────────────────────────────────────────────────────────┐
│                    Client Applications                       │
├─────────────────┬─────────────────┬─────────────────────────┤
│   Web App       │  Mobile Web     │    Future Native Apps   │
│ (Render Static) │                 │                         │
└────────┬────────┴────────┬────────┴──────────┬──────────────┘
         │                 │                    │
         └─────────────────┴────────────────────┘
                           │
                    ┌──────▼──────┐
                    │  Cloudflare  │
                    │     CDN      │
                    └──────┬──────┘
                           │
         ┌─────────────────┴─────────────────┐
         │      Render Load Balancer         │
         │     (Automatic & Managed)         │
         └─────────────────┬─────────────────┘
                           │
    ┌──────────────────────┴──────────────────────┐
    │         Render Web Service (API)            │
    │            - Auto-scaling                   │
    │            - Health checks                  │
    │            - Zero downtime deploys          │
    └──────────────────────┬──────────────────────┘
                           │
    ┌──────────────────────┴──────────────────────┐
    │           Application Services              │
    ├──────────────┬──────────────┬───────────────┤
    │ Trip Service │ Place Service│ User Service  │
    ├──────────────┼──────────────┼───────────────┤
    │ Media Service│ Share Service│ Suggestion Svc│
    └──────┬───────┴──────┬───────┴───────┬───────┘
           │              │               │
    ┌──────▼──────────────▼───────────────▼──────┐
    │         Render Infrastructure              │
    ├─────────────┬─────────────┬────────────────┤
    │  PostgreSQL │ Key-Value   │ Persistent     │
    │  (Managed)  │ (Redis)     │ Disks          │
    └─────────────┴─────────────┴────────────────┘
```

### Service Types on Render

1. **Web Services (API)**
   - Docker-based deployment
   - Auto-scaling capabilities
   - Health check monitoring
   - Zero-downtime deploys
   - Private networking between services

2. **Static Sites (Frontend)**
   - Global CDN distribution
   - Automatic HTTPS
   - Pull request previews
   - Custom headers support

3. **PostgreSQL Database**
   - Managed database with backups
   - Read replicas for scaling
   - High availability option
   - Connection pooling
   - Extensions support (PostGIS, pgvector)

4. **Key-Value Store (Redis)**
   - Valkey 8 (Redis-compatible)
   - Persistent storage for paid instances
   - Internal networking for low latency
   - Configurable eviction policies

5. **Persistent Disks**
   - SSD storage for media files
   - Encrypted at rest
   - Daily snapshots
   - Mountable to web services

## Database Design with PostgreSQL

### PostgreSQL Schema Design

#### Users Table
```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    username VARCHAR(100) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    display_name VARCHAR(255),
    avatar_url TEXT,
    bio TEXT,
    location VARCHAR(255),
    roles TEXT[] DEFAULT ARRAY['user'],
    profile_visibility VARCHAR(50) DEFAULT 'public',
    location_sharing BOOLEAN DEFAULT false,
    trip_default_privacy VARCHAR(50) DEFAULT 'private',
    email_notifications BOOLEAN DEFAULT true,
    push_notifications BOOLEAN DEFAULT true,
    suggestion_notifications BOOLEAN DEFAULT true,
    trip_invite_notifications BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    last_active TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    status VARCHAR(50) DEFAULT 'active'
);

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_status ON users(status);
```

#### Trips Table
```sql
CREATE TABLE trips (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(255) NOT NULL,
    description TEXT,
    owner_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    cover_image TEXT,
    privacy VARCHAR(50) DEFAULT 'private',
    status VARCHAR(50) DEFAULT 'planning',
    start_date DATE,
    end_date DATE,
    timezone VARCHAR(100),
    tags TEXT[],
    view_count INTEGER DEFAULT 0,
    share_count INTEGER DEFAULT 0,
    suggestion_count INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_trips_owner ON trips(owner_id);
CREATE INDEX idx_trips_status ON trips(status);
CREATE INDEX idx_trips_privacy ON trips(privacy);
CREATE INDEX idx_trips_dates ON trips(start_date, end_date);
CREATE INDEX idx_trips_search ON trips USING gin(to_tsvector('english', title || ' ' || COALESCE(description, '')));
```

#### Places Table with PostGIS
```sql
CREATE EXTENSION IF NOT EXISTS postgis;
CREATE EXTENSION IF NOT EXISTS postgis_topology;

CREATE TABLE places (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    type VARCHAR(50) NOT NULL, -- 'poi', 'area', 'region'
    parent_id UUID REFERENCES places(id) ON DELETE SET NULL,
    location GEOGRAPHY(POINT, 4326),
    bounds GEOGRAPHY(POLYGON, 4326),
    street_address VARCHAR(255),
    city VARCHAR(100),
    state VARCHAR(100),
    country VARCHAR(100),
    postal_code VARCHAR(20),
    created_by UUID NOT NULL REFERENCES users(id),
    category TEXT[],
    tags TEXT[],
    opening_hours JSONB,
    contact_info JSONB,
    amenities TEXT[],
    average_rating DECIMAL(3,2),
    rating_count INTEGER DEFAULT 0,
    privacy VARCHAR(50) DEFAULT 'public',
    status VARCHAR(50) DEFAULT 'active',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_places_location ON places USING GIST(location);
CREATE INDEX idx_places_bounds ON places USING GIST(bounds);
CREATE INDEX idx_places_created_by ON places(created_by);
CREATE INDEX idx_places_parent ON places(parent_id);
CREATE INDEX idx_places_search ON places USING gin(to_tsvector('english', name || ' ' || COALESCE(description, '')));
```

#### Trip Collaborators Table
```sql
CREATE TABLE trip_collaborators (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    trip_id UUID NOT NULL REFERENCES trips(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(50) NOT NULL, -- 'admin', 'editor', 'viewer'
    can_edit BOOLEAN DEFAULT false,
    can_delete BOOLEAN DEFAULT false,
    can_invite BOOLEAN DEFAULT false,
    can_moderate_suggestions BOOLEAN DEFAULT false,
    invited_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    joined_at TIMESTAMP WITH TIME ZONE,
    UNIQUE(trip_id, user_id)
);

CREATE INDEX idx_collaborators_trip ON trip_collaborators(trip_id);
CREATE INDEX idx_collaborators_user ON trip_collaborators(user_id);
```

#### Trip Waypoints Table
```sql
CREATE TABLE trip_waypoints (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    trip_id UUID NOT NULL REFERENCES trips(id) ON DELETE CASCADE,
    place_id UUID NOT NULL REFERENCES places(id),
    order_position INTEGER NOT NULL,
    arrival_time TIMESTAMP WITH TIME ZONE,
    departure_time TIMESTAMP WITH TIME ZONE,
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(trip_id, order_position)
);

CREATE INDEX idx_waypoints_trip ON trip_waypoints(trip_id);
CREATE INDEX idx_waypoints_place ON trip_waypoints(place_id);
```

#### Media Table
```sql
CREATE TABLE media (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    filename VARCHAR(255) NOT NULL,
    original_name VARCHAR(255) NOT NULL,
    mime_type VARCHAR(100) NOT NULL,
    size_bytes BIGINT NOT NULL,
    storage_path TEXT NOT NULL,
    cdn_url TEXT,
    thumbnail_small TEXT,
    thumbnail_medium TEXT,
    thumbnail_large TEXT,
    width INTEGER,
    height INTEGER,
    duration_seconds INTEGER, -- for videos
    location GEOGRAPHY(POINT, 4326),
    uploaded_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_media_uploaded_by ON media(uploaded_by);
CREATE INDEX idx_media_location ON media USING GIST(location);
```

#### Suggestions Table
```sql
CREATE TABLE suggestions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    target_type VARCHAR(50) NOT NULL, -- 'trip', 'place'
    target_id UUID NOT NULL,
    suggested_by UUID NOT NULL REFERENCES users(id),
    type VARCHAR(50) NOT NULL, -- 'edit', 'addition', 'deletion', 'comment'
    status VARCHAR(50) DEFAULT 'pending',
    field_name VARCHAR(100),
    current_value TEXT,
    suggested_value TEXT,
    reason TEXT,
    reviewed_by UUID REFERENCES users(id),
    reviewed_at TIMESTAMP WITH TIME ZONE,
    decision VARCHAR(50),
    review_notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_suggestions_target ON suggestions(target_type, target_id);
CREATE INDEX idx_suggestions_user ON suggestions(suggested_by);
CREATE INDEX idx_suggestions_status ON suggestions(status);
```

## Redis Cache Implementation

### Render Key-Value Configuration

```go
// cache/redis.go
package cache

import (
    "context"
    "encoding/json"
    "time"
    "github.com/redis/go-redis/v9"
)

type RedisCache struct {
    client *redis.Client
    ttls   map[string]time.Duration
}

func NewRedisCache(internalURL string) (*RedisCache, error) {
    // Use internal URL for better performance
    opt, err := redis.ParseURL(internalURL)
    if err != nil {
        return nil, err
    }
    
    client := redis.NewClient(opt)
    
    // Test connection
    ctx := context.Background()
    if err := client.Ping(ctx).Err(); err != nil {
        return nil, err
    }
    
    return &RedisCache{
        client: client,
        ttls: map[string]time.Duration{
            "user":         1 * time.Hour,
            "trip":         30 * time.Minute,
            "place":        1 * time.Hour,
            "search":       5 * time.Minute,
            "nearby":       15 * time.Minute,
            "permissions":  30 * time.Minute,
            "session":      24 * time.Hour,
        },
    }, nil
}

// Configure maxmemory policy for caching
func (r *RedisCache) ConfigureForCaching(ctx context.Context) error {
    // Set eviction policy to allkeys-lru for caching
    return r.client.ConfigSet(ctx, "maxmemory-policy", "allkeys-lru").Err()
}
```

### Cache Key Strategy for PostgreSQL
```go
// Cache keys adapted for PostgreSQL UUIDs
const (
    UserKey         = "user:%s"              // user:uuid
    UserTripsKey    = "user:%s:trips"       // user:uuid:trips
    UserPlacesKey   = "user:%s:places"      // user:uuid:places
    TripKey         = "trip:%s"              // trip:uuid
    TripWaypointsKey = "trip:%s:waypoints"   // trip:uuid:waypoints
    PlaceKey        = "place:%s"             // place:uuid
    NearbyPlacesKey = "places:nearby:%f:%f:%d" // lat:lng:radius
    SearchResultsKey = "search:%s:%d:%d"     // query:offset:limit
)
```

## Media Storage Strategy

### Using Persistent Disks with CDN

```go
// media/storage.go
package media

import (
    "fmt"
    "io"
    "os"
    "path/filepath"
    "github.com/google/uuid"
)

type DiskStorage struct {
    basePath string
    cdnURL   string
}

func NewDiskStorage(mountPath, cdnURL string) *DiskStorage {
    return &DiskStorage{
        basePath: mountPath,
        cdnURL:   cdnURL,
    }
}

func (s *DiskStorage) Upload(file io.Reader, contentType string) (*MediaFile, error) {
    // Generate unique filename
    fileID := uuid.New().String()
    ext := getExtension(contentType)
    filename := fmt.Sprintf("%s%s", fileID, ext)
    
    // Create directory structure (year/month/day)
    now := time.Now()
    dirPath := filepath.Join(
        s.basePath,
        fmt.Sprintf("%d", now.Year()),
        fmt.Sprintf("%02d", now.Month()),
        fmt.Sprintf("%02d", now.Day()),
    )
    
    if err := os.MkdirAll(dirPath, 0755); err != nil {
        return nil, err
    }
    
    // Save file
    fullPath := filepath.Join(dirPath, filename)
    dst, err := os.Create(fullPath)
    if err != nil {
        return nil, err
    }
    defer dst.Close()
    
    size, err := io.Copy(dst, file)
    if err != nil {
        return nil, err
    }
    
    // Generate CDN URL
    relativePath := filepath.Join(
        fmt.Sprintf("%d", now.Year()),
        fmt.Sprintf("%02d", now.Month()),
        fmt.Sprintf("%02d", now.Day()),
        filename,
    )
    
    return &MediaFile{
        ID:          fileID,
        Filename:    filename,
        StoragePath: fullPath,
        CDNUrl:      fmt.Sprintf("%s/%s", s.cdnURL, relativePath),
        Size:        size,
    }, nil
}
```

### Image Processing
```go
// media/processor.go
func (p *ImageProcessor) GenerateThumbnails(inputPath string) (*Thumbnails, error) {
    // Use imaging library to create thumbnails
    src, err := imaging.Open(inputPath)
    if err != nil {
        return nil, err
    }
    
    thumbnails := &Thumbnails{}
    
    // Small thumbnail (150x150)
    small := imaging.Thumbnail(src, 150, 150, imaging.Lanczos)
    thumbnails.Small = p.saveThumbnail(small, inputPath, "small")
    
    // Medium thumbnail (400x400)
    medium := imaging.Thumbnail(src, 400, 400, imaging.Lanczos)
    thumbnails.Medium = p.saveThumbnail(medium, inputPath, "medium")
    
    // Large thumbnail (800x800)
    large := imaging.Thumbnail(src, 800, 800, imaging.Lanczos)
    thumbnails.Large = p.saveThumbnail(large, inputPath, "large")
    
    return thumbnails, nil
}
```

## Deployment Configuration

### render.yaml Configuration
```yaml
services:
  # Go API Service
  - type: web
    name: trip-planner-api
    runtime: docker
    dockerfilePath: ./apps/api/Dockerfile
    dockerContext: .
    repo: https://github.com/yourusername/trip-planner
    buildFilter:
      paths:
        - apps/api/**
        - packages/**
    envVars:
      - key: PORT
        value: 8080
      - key: DATABASE_URL
        fromDatabase:
          name: trip-planner-db
          property: connectionString
      - key: REDIS_URL
        fromService:
          name: trip-planner-cache
          type: redis
          property: connectionString
      - key: INTERNAL_REDIS_URL
        fromService:
          name: trip-planner-cache
          type: redis
          property: internalConnectionString
      - key: MEDIA_PATH
        value: /data/media
      - key: CDN_URL
        fromService:
          name: trip-planner-web
          type: web
          envVarKey: RENDER_EXTERNAL_URL
    disk:
      name: media-storage
      mountPath: /data
      sizeGB: 100
    healthCheckPath: /api/health
    autoDeploy: true
    scaling:
      minInstances: 1
      maxInstances: 10
      targetMemoryPercent: 80
      targetCPUPercent: 70

  # React Frontend
  - type: web
    name: trip-planner-web
    runtime: static
    buildCommand: cd apps/web && npm install && npm run build
    staticPublishPath: apps/web/dist
    pullRequestPreviewsEnabled: true
    headers:
      - path: /*
        name: Cache-Control
        value: public, max-age=31536000, immutable
      - path: /index.html
        name: Cache-Control
        value: no-cache, no-store, must-revalidate
      - path: /api/*
        name: X-Frame-Options
        value: DENY
    routes:
      - type: rewrite
        source: /media/*
        destination: https://trip-planner-api.onrender.com/media/*

  # Redis Cache Service
  - type: redis
    name: trip-planner-cache
    plan: starter # or standard/pro based on needs
    maxmemoryPolicy: allkeys-lru
    ipAllowList: [] # Configure if using external access

databases:
  # PostgreSQL Database
  - name: trip-planner-db
    plan: starter # or standard/pro
    databaseName: trip_planner
    user: trip_planner_user
    postgresMajorVersion: 15
```

### Dockerfile for API
```dockerfile
# apps/api/Dockerfile
FROM golang:1.21-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

COPY apps/api/go.mod apps/api/go.sum ./
RUN go mod download

COPY apps/api/ .
COPY packages/ ../packages/

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main cmd/server/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

COPY --from=builder /app/main .

# Create directory for media storage
RUN mkdir -p /data/media

EXPOSE 8080

CMD ["./main"]
```

## Environment Setup

### Production Environment Variables
```env
# API Service Environment
NODE_ENV=production
PORT=8080
API_VERSION=v1

# Database (from Render PostgreSQL)
DATABASE_URL=postgresql://user:pass@host:5432/trip_planner?sslmode=require
DB_MAX_CONNECTIONS=100
DB_IDLE_CONNECTIONS=10

# Redis Cache (from Render Key-Value)
REDIS_URL=redis://red-xxx.redis.render.com:6379
INTERNAL_REDIS_URL=redis://red-xxx:6379
REDIS_MAX_RETRIES=3
REDIS_POOL_SIZE=10

# Authentication
JWT_SECRET=your-secure-jwt-secret
JWT_EXPIRY=7d
REFRESH_TOKEN_EXPIRY=30d
BCRYPT_COST=12

# Media Storage
MEDIA_PATH=/data/media
MAX_UPLOAD_SIZE=52428800
ALLOWED_MIME_TYPES=image/jpeg,image/png,image/webp,video/mp4
CDN_URL=https://trip-planner-web.onrender.com
THUMBNAIL_QUALITY=85

# External Services
MAPBOX_API_KEY=pk.xxx
MAPBOX_STYLE_URL=mapbox://styles/mapbox/streets-v11

# Email Service (using Render's built-in)
SMTP_HOST=smtp.render.com
SMTP_PORT=587
SMTP_USER=your-email@render.com
SMTP_FROM=noreply@yourdomain.com

# Monitoring
SENTRY_DSN=https://xxx@xxx.ingest.sentry.io/xxx
LOG_LEVEL=info

# Security
CORS_ORIGINS=https://trip-planner-web.onrender.com,https://yourdomain.com
RATE_LIMIT_REQUESTS=100
RATE_LIMIT_WINDOW=1m
SESSION_LIFETIME=86400

# Feature Flags
ENABLE_SOCIAL_LOGIN=false
ENABLE_ML_RECOMMENDATIONS=false
ENABLE_REAL_TIME=true
```

### Frontend Environment Variables
```env
# apps/web/.env.production
VITE_API_URL=https://trip-planner-api.onrender.com/api/v1
VITE_WS_URL=wss://trip-planner-api.onrender.com
VITE_MAPBOX_TOKEN=pk.xxx
VITE_SENTRY_DSN=https://xxx@xxx.ingest.sentry.io/xxx
VITE_GA_ID=G-XXXXXXXXXX
```

## Cost Optimization

### Render Pricing Considerations

1. **Database Optimization**
   - Start with Starter plan ($7/month)
   - Monitor connections and storage
   - Use read replicas only when needed
   - Enable connection pooling

2. **Redis Optimization**
   - Use Starter plan initially
   - Configure appropriate eviction policy
   - Monitor memory usage
   - Consider data compression

3. **Web Service Optimization**
   - Use autoscaling to handle traffic
   - Start with 512MB RAM instances
   - Monitor CPU and memory metrics
   - Use health checks to prevent over-provisioning

4. **Storage Optimization**
   - Compress images before storage
   - Implement cleanup policies
   - Use appropriate thumbnail sizes
   - Consider external CDN for high traffic

### Performance Monitoring
```go
// monitoring/metrics.go
package monitoring

import (
    "github.com/gin-gonic/gin"
    "time"
)

func MetricsMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        
        c.Next()
        
        // Log to Render's built-in metrics
        duration := time.Since(start)
        status := c.Writer.Status()
        path := c.Request.URL.Path
        
        // These will appear in Render dashboard
        log.Printf("metrics: path=%s status=%d duration=%dms",
            path, status, duration.Milliseconds())
    }
}
```

## Migration from MongoDB to PostgreSQL

### Key Schema Differences

1. **Document → Relational**
   - Embedded documents become separate tables
   - Arrays become junction tables or PostgreSQL arrays
   - Use JSONB for flexible fields

2. **Indexes**
   - Convert 2dsphere indexes to PostGIS
   - Text indexes to PostgreSQL full-text search
   - Compound indexes work similarly

3. **Queries**
   - Aggregation pipelines → SQL with CTEs
   - Geospatial queries → PostGIS functions
   - Text search → PostgreSQL ts_vector

### Migration Script Example
```go
// migrations/001_initial_schema.sql
BEGIN;

-- Enable extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "postgis";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";

-- Create all tables
CREATE TABLE users (...);
CREATE TABLE trips (...);
CREATE TABLE places (...);
-- ... rest of schema

-- Create indexes
CREATE INDEX CONCURRENTLY ...;

-- Add foreign key constraints
ALTER TABLE trips ADD CONSTRAINT ...;

COMMIT;
```

This architecture leverages Render's managed services to simplify deployment while maintaining scalability and performance. The PostgreSQL + PostGIS combination provides powerful geospatial capabilities, while Render's built-in features handle many operational concerns automatically.