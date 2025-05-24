# Trip Planning Platform - Complete Architecture Plan

## Table of Contents
1. [System Overview](#system-overview)
2. [High-Level Architecture](#high-level-architecture)
3. [Database Design](#database-design)
4. [Backend Architecture](#backend-architecture)
5. [Frontend Architecture](#frontend-architecture)
6. [API Design](#api-design)
7. [Security & RBAC Architecture](#security--rbac-architecture)
8. [Real-time Architecture](#real-time-architecture)
9. [Caching Strategy](#caching-strategy)
10. [Deployment Architecture](#deployment-architecture)
11. [CI/CD Pipeline](#cicd-pipeline)
12. [Performance Optimization](#performance-optimization)
13. [Future ML Integration](#future-ml-integration)

## System Overview

### Architecture Principles
- **Microservices-ready monolith**: Start with a well-structured monolith that can be split later
- **Event-driven architecture**: For real-time features and decoupling
- **Domain-driven design**: Clear boundaries between business domains
- **CQRS pattern**: Separate read and write operations for scalability
- **API-first approach**: Well-documented RESTful APIs with GraphQL consideration

### Tech Stack Summary
```yaml
Frontend:
  - React 18+ with TypeScript
  - Mapbox GL JS v3
  - Redux Toolkit for state management
  - Socket.io client for real-time
  - Tailwind CSS for styling
  - Vite for build tooling

Backend:
  - Go 1.21+ with Gin framework
  - MongoDB Atlas for primary database
  - Redis for caching and sessions
  - MinIO/S3 for media storage
  - Socket.io for WebSockets
  - JWT for authentication

Infrastructure:
  - Render for deployment
  - Cloudflare CDN
  - GitHub Actions for CI/CD
  - Sentry for error tracking
  - DataDog for monitoring
```

## High-Level Architecture

### System Architecture Diagram
```
┌─────────────────────────────────────────────────────────────────┐
│                        Client Applications                       │
├─────────────────┬─────────────────┬─────────────────────────────┤
│   Web App (React)│  Mobile Web     │    Future Native Apps       │
└────────┬────────┴────────┬────────┴──────────┬─────────────────┘
         │                 │                    │
         └─────────────────┴────────────────────┘
                           │
                    ┌──────▼──────┐
                    │  Cloudflare  │
                    │     CDN      │
                    └──────┬──────┘
                           │
         ┌─────────────────┴─────────────────┐
         │          Load Balancer            │
         │          (Render.com)             │
         └─────────────────┬─────────────────┘
                           │
    ┌──────────────────────┴──────────────────────┐
    │              API Gateway Layer              │
    │        (Authentication, Rate Limiting)      │
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
    │          Core Infrastructure               │
    ├─────────────┬─────────────┬────────────────┤
    │  MongoDB    │   Redis     │  MinIO/S3      │
    │  Atlas      │   Cache     │  Storage       │
    └─────────────┴─────────────┴────────────────┘
```

### Monorepo Structure
```
trip-planner/
├── .github/
│   └── workflows/
│       ├── ci.yml
│       ├── deploy-api.yml
│       └── deploy-web.yml
├── apps/
│   ├── api/                 # Go backend
│   │   ├── cmd/
│   │   │   └── server/
│   │   ├── internal/
│   │   │   ├── auth/
│   │   │   ├── trips/
│   │   │   ├── places/
│   │   │   ├── users/
│   │   │   ├── media/
│   │   │   ├── suggestions/
│   │   │   └── realtime/
│   │   ├── pkg/
│   │   │   ├── database/
│   │   │   ├── cache/
│   │   │   ├── storage/
│   │   │   └── utils/
│   │   ├── go.mod
│   │   └── Dockerfile
│   └── web/                 # React frontend
│       ├── src/
│       │   ├── components/
│       │   ├── features/
│       │   ├── hooks/
│       │   ├── services/
│       │   ├── store/
│       │   └── utils/
│       ├── package.json
│       └── Dockerfile
├── packages/               # Shared packages
│   ├── types/             # TypeScript types
│   ├── validators/        # Shared validation
│   └── constants/         # Shared constants
├── scripts/               # Build and deploy scripts
├── docker-compose.yml     # Local development
└── README.md
```

## Database Design

### MongoDB Collections Schema

#### Users Collection
```javascript
{
  _id: ObjectId,
  email: String,
  username: String,
  password_hash: String,
  profile: {
    display_name: String,
    avatar_url: String,
    bio: String,
    location: String
  },
  roles: [String], // ['user', 'admin', 'moderator']
  preferences: {
    privacy: {
      profile_visibility: String, // 'public', 'friends', 'private'
      location_sharing: Boolean,
      trip_default_privacy: String
    },
    notifications: {
      email: Boolean,
      push: Boolean,
      suggestions: Boolean,
      trip_invites: Boolean
    }
  },
  friends: [ObjectId], // References to other users
  blocked_users: [ObjectId],
  created_at: Date,
  updated_at: Date,
  last_active: Date,
  status: String // 'active', 'suspended', 'deleted'
}
```

#### Trips Collection
```javascript
{
  _id: ObjectId,
  title: String,
  description: String,
  owner_id: ObjectId,
  collaborators: [{
    user_id: ObjectId,
    role: String, // 'admin', 'editor', 'viewer'
    permissions: {
      can_edit: Boolean,
      can_delete: Boolean,
      can_invite: Boolean,
      can_moderate_suggestions: Boolean
    },
    invited_at: Date,
    joined_at: Date
  }],
  route: {
    waypoints: [{
      place_id: ObjectId,
      order: Number,
      arrival_time: Date,
      departure_time: Date,
      notes: String
    }],
    total_distance: Number,
    estimated_duration: Number
  },
  schedule: {
    start_date: Date,
    end_date: Date,
    timezone: String
  },
  privacy: String, // 'public', 'friends', 'private', 'invite_only'
  tags: [String],
  cover_image: String,
  status: String, // 'planning', 'active', 'completed', 'cancelled'
  statistics: {
    views: Number,
    shares: Number,
    suggestions: Number
  },
  created_at: Date,
  updated_at: Date,
  deleted_at: Date
}
```

#### Places Collection
```javascript
{
  _id: ObjectId,
  name: String,
  description: String,
  type: String, // 'poi', 'area', 'region'
  parent_id: ObjectId, // For hierarchical organization
  location: {
    type: "Point",
    coordinates: [longitude, latitude]
  },
  bounds: { // For areas/regions
    type: "Polygon",
    coordinates: [[[]]]
  },
  address: {
    street: String,
    city: String,
    state: String,
    country: String,
    postal_code: String
  },
  created_by: ObjectId,
  collaborators: [{
    user_id: ObjectId,
    role: String,
    permissions: Object
  }],
  media: [{
    type: String, // 'photo', 'video'
    url: String,
    thumbnail_url: String,
    caption: String,
    uploaded_by: ObjectId,
    uploaded_at: Date
  }],
  attributes: {
    category: [String],
    tags: [String],
    opening_hours: Object,
    contact: Object,
    amenities: [String]
  },
  ratings: {
    average: Number,
    count: Number
  },
  privacy: String,
  status: String, // 'active', 'pending', 'archived'
  created_at: Date,
  updated_at: Date
}
```

#### Suggestions Collection
```javascript
{
  _id: ObjectId,
  target_type: String, // 'trip', 'place'
  target_id: ObjectId,
  suggested_by: ObjectId,
  type: String, // 'edit', 'addition', 'deletion', 'comment'
  status: String, // 'pending', 'approved', 'rejected', 'implemented'
  content: {
    field: String, // Which field to modify
    current_value: Mixed,
    suggested_value: Mixed,
    reason: String
  },
  moderation: {
    reviewed_by: ObjectId,
    reviewed_at: Date,
    decision: String,
    notes: String
  },
  discussion: [{
    user_id: ObjectId,
    message: String,
    timestamp: Date
  }],
  created_at: Date,
  updated_at: Date
}
```

#### Media Collection
```javascript
{
  _id: ObjectId,
  filename: String,
  original_name: String,
  mime_type: String,
  size: Number,
  storage_path: String,
  cdn_url: String,
  thumbnails: {
    small: String,
    medium: String,
    large: String
  },
  metadata: {
    width: Number,
    height: Number,
    duration: Number, // For videos
    exif: Object,
    location: {
      type: "Point",
      coordinates: [longitude, latitude]
    }
  },
  uploaded_by: ObjectId,
  used_in: [{
    type: String, // 'trip', 'place', 'profile'
    id: ObjectId
  }],
  tags: [String],
  created_at: Date
}
```

### Database Indexes
```javascript
// Geospatial indexes
db.places.createIndex({ "location": "2dsphere" })
db.places.createIndex({ "bounds": "2dsphere" })
db.media.createIndex({ "metadata.location": "2dsphere" })

// Performance indexes
db.users.createIndex({ "email": 1 }, { unique: true })
db.users.createIndex({ "username": 1 }, { unique: true })
db.trips.createIndex({ "owner_id": 1, "status": 1 })
db.trips.createIndex({ "collaborators.user_id": 1 })
db.places.createIndex({ "created_by": 1 })
db.places.createIndex({ "parent_id": 1 })
db.suggestions.createIndex({ "target_id": 1, "status": 1 })
db.suggestions.createIndex({ "suggested_by": 1 })

// Text search indexes
db.trips.createIndex({ "title": "text", "description": "text" })
db.places.createIndex({ "name": "text", "description": "text" })
```

## Backend Architecture

### Go Project Structure
```
api/
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── auth/
│   │   ├── handler.go
│   │   ├── service.go
│   │   ├── middleware.go
│   │   └── jwt.go
│   ├── trips/
│   │   ├── handler.go
│   │   ├── service.go
│   │   ├── repository.go
│   │   └── models.go
│   ├── places/
│   │   ├── handler.go
│   │   ├── service.go
│   │   ├── repository.go
│   │   └── models.go
│   ├── users/
│   │   ├── handler.go
│   │   ├── service.go
│   │   ├── repository.go
│   │   └── models.go
│   ├── media/
│   │   ├── handler.go
│   │   ├── service.go
│   │   ├── storage.go
│   │   └── processor.go
│   ├── suggestions/
│   │   ├── handler.go
│   │   ├── service.go
│   │   ├── repository.go
│   │   └── workflow.go
│   ├── permissions/
│   │   ├── rbac.go
│   │   ├── policies.go
│   │   └── middleware.go
│   └── realtime/
│       ├── hub.go
│       ├── client.go
│       └── events.go
├── pkg/
│   ├── database/
│   │   ├── mongodb.go
│   │   └── migrations.go
│   ├── cache/
│   │   ├── redis.go
│   │   └── strategies.go
│   ├── storage/
│   │   ├── s3.go
│   │   └── cdn.go
│   ├── mapbox/
│   │   └── client.go
│   ├── email/
│   │   └── sender.go
│   └── utils/
│       ├── validator.go
│       ├── logger.go
│       └── errors.go
├── config/
│   └── config.go
├── go.mod
└── go.sum
```

### Service Layer Architecture
```go
// Example Trip Service
type TripService struct {
    repo       TripRepository
    placeRepo  PlaceRepository
    cache      cache.Cache
    events     events.Publisher
    mapbox     mapbox.Client
    rbac       permissions.RBAC
}

func (s *TripService) CreateTrip(ctx context.Context, userId string, input CreateTripInput) (*Trip, error) {
    // Validate input
    if err := input.Validate(); err != nil {
        return nil, err
    }
    
    // Create trip
    trip := &Trip{
        ID:        primitive.NewObjectID(),
        Title:     input.Title,
        OwnerID:   userId,
        CreatedAt: time.Now(),
    }
    
    // Set default permissions
    trip.Collaborators = []Collaborator{{
        UserID: userId,
        Role:   "admin",
        Permissions: permissions.AdminPermissions(),
    }}
    
    // Save to database
    if err := s.repo.Create(ctx, trip); err != nil {
        return nil, err
    }
    
    // Publish event
    s.events.Publish("trip.created", trip)
    
    // Invalidate cache
    s.cache.Delete(fmt.Sprintf("user_trips:%s", userId))
    
    return trip, nil
}
```

### Repository Pattern
```go
type TripRepository interface {
    Create(ctx context.Context, trip *Trip) error
    GetByID(ctx context.Context, id string) (*Trip, error)
    Update(ctx context.Context, id string, updates map[string]interface{}) error
    Delete(ctx context.Context, id string) error
    ListByUser(ctx context.Context, userId string, filters TripFilters) ([]*Trip, error)
    AddCollaborator(ctx context.Context, tripId string, collaborator Collaborator) error
    RemoveCollaborator(ctx context.Context, tripId, userId string) error
}
```

## Frontend Architecture

### Component Structure
```
src/
├── components/
│   ├── common/
│   │   ├── Layout/
│   │   ├── Navigation/
│   │   ├── Modal/
│   │   └── LoadingSpinner/
│   ├── map/
│   │   ├── MapContainer/
│   │   ├── MarkerCluster/
│   │   ├── RouteLayer/
│   │   └── PlacePopup/
│   └── forms/
│       ├── TripForm/
│       ├── PlaceForm/
│       └── SuggestionForm/
├── features/
│   ├── trips/
│   │   ├── TripList/
│   │   ├── TripDetail/
│   │   ├── TripPlanner/
│   │   └── TripShare/
│   ├── places/
│   │   ├── PlaceExplorer/
│   │   ├── PlaceDetail/
│   │   └── PlaceGallery/
│   ├── suggestions/
│   │   ├── SuggestionList/
│   │   ├── SuggestionReview/
│   │   └── SuggestionModal/
│   └── auth/
│       ├── Login/
│       ├── Register/
│       └── Profile/
├── hooks/
│   ├── useAuth.ts
│   ├── useRealtime.ts
│   ├── useMapbox.ts
│   └── usePermissions.ts
├── services/
│   ├── api/
│   │   ├── client.ts
│   │   ├── trips.ts
│   │   ├── places.ts
│   │   └── suggestions.ts
│   └── realtime/
│       └── socket.ts
├── store/
│   ├── index.ts
│   ├── auth/
│   ├── trips/
│   ├── places/
│   └── ui/
└── utils/
    ├── permissions.ts
    ├── validators.ts
    └── formatters.ts
```

### State Management (Redux Toolkit)
```typescript
// store/trips/slice.ts
import { createSlice, createAsyncThunk } from '@reduxjs/toolkit';

export const fetchTrips = createAsyncThunk(
  'trips/fetchTrips',
  async (filters: TripFilters) => {
    const response = await api.trips.list(filters);
    return response.data;
  }
);

const tripsSlice = createSlice({
  name: 'trips',
  initialState: {
    items: [],
    currentTrip: null,
    loading: false,
    error: null,
  },
  reducers: {
    setCurrentTrip: (state, action) => {
      state.currentTrip = action.payload;
    },
    updateTripLocally: (state, action) => {
      const index = state.items.findIndex(t => t.id === action.payload.id);
      if (index !== -1) {
        state.items[index] = action.payload;
      }
    },
  },
  extraReducers: (builder) => {
    builder
      .addCase(fetchTrips.pending, (state) => {
        state.loading = true;
      })
      .addCase(fetchTrips.fulfilled, (state, action) => {
        state.loading = false;
        state.items = action.payload;
      })
      .addCase(fetchTrips.rejected, (state, action) => {
        state.loading = false;
        state.error = action.error.message;
      });
  },
});
```

### Real-time Hook
```typescript
// hooks/useRealtime.ts
export function useRealtime() {
  const dispatch = useAppDispatch();
  const { user } = useAuth();
  
  useEffect(() => {
    if (!user) return;
    
    const socket = io(WEBSOCKET_URL, {
      auth: { token: getAuthToken() },
    });
    
    socket.on('trip.updated', (data) => {
      dispatch(updateTripLocally(data));
    });
    
    socket.on('suggestion.created', (data) => {
      if (hasPermission(user, data.targetId, 'moderate')) {
        dispatch(addSuggestion(data));
        showNotification('New suggestion received');
      }
    });
    
    socket.on('place.shared', (data) => {
      dispatch(addSharedPlace(data));
    });
    
    return () => {
      socket.disconnect();
    };
  }, [user, dispatch]);
}
```

## API Design

### RESTful Endpoints

#### Authentication Endpoints
```
POST   /api/auth/register
POST   /api/auth/login
POST   /api/auth/refresh
POST   /api/auth/logout
GET    /api/auth/me
```

#### Trip Endpoints
```
GET    /api/trips                    # List trips (with filters)
POST   /api/trips                    # Create trip
GET    /api/trips/:id                # Get trip details
PUT    /api/trips/:id                # Update trip (admin only)
DELETE /api/trips/:id                # Delete trip (admin only)
POST   /api/trips/:id/collaborators  # Add collaborator (admin only)
DELETE /api/trips/:id/collaborators/:userId  # Remove collaborator
POST   /api/trips/:id/waypoints      # Add waypoint
PUT    /api/trips/:id/waypoints/:waypointId  # Update waypoint
DELETE /api/trips/:id/waypoints/:waypointId  # Remove waypoint
POST   /api/trips/:id/share          # Share trip
GET    /api/trips/:id/suggestions    # Get trip suggestions
```

#### Place Endpoints
```
GET    /api/places                   # List places (with geospatial queries)
POST   /api/places                   # Create place
GET    /api/places/:id               # Get place details
PUT    /api/places/:id               # Update place (admin only)
DELETE /api/places/:id               # Delete place (admin only)
POST   /api/places/:id/media         # Upload media
DELETE /api/places/:id/media/:mediaId # Remove media
GET    /api/places/search            # Search places
GET    /api/places/nearby            # Get nearby places
POST   /api/places/:id/suggestions   # Submit suggestion (viewer)
```

#### Suggestion Endpoints
```
GET    /api/suggestions              # List suggestions (with filters)
POST   /api/suggestions              # Create suggestion
GET    /api/suggestions/:id          # Get suggestion details
PUT    /api/suggestions/:id/status   # Update suggestion status (admin only)
POST   /api/suggestions/:id/comments # Add comment
```

#### User Endpoints
```
GET    /api/users/profile            # Get current user profile
PUT    /api/users/profile            # Update profile
GET    /api/users/:id                # Get user public profile
POST   /api/users/friends            # Send friend request
PUT    /api/users/friends/:id        # Accept/reject friend request
DELETE /api/users/friends/:id        # Remove friend
```

### API Response Format
```json
{
  "success": true,
  "data": {
    // Response data
  },
  "meta": {
    "page": 1,
    "limit": 20,
    "total": 100,
    "hasMore": true
  },
  "error": null
}
```

### Error Response Format
```json
{
  "success": false,
  "data": null,
  "error": {
    "code": "UNAUTHORIZED",
    "message": "You don't have permission to perform this action",
    "details": {
      "required_role": "admin",
      "current_role": "viewer"
    }
  }
}
```

## Security & RBAC Architecture

### Role-Based Access Control Implementation

#### Permission Model
```go
type Permission string

const (
    // Trip permissions
    TripView   Permission = "trip.view"
    TripEdit   Permission = "trip.edit"
    TripDelete Permission = "trip.delete"
    TripShare  Permission = "trip.share"
    TripInvite Permission = "trip.invite"
    
    // Place permissions
    PlaceView   Permission = "place.view"
    PlaceEdit   Permission = "place.edit"
    PlaceDelete Permission = "place.delete"
    PlaceMedia  Permission = "place.media"
    
    // Suggestion permissions
    SuggestionCreate   Permission = "suggestion.create"
    SuggestionModerate Permission = "suggestion.moderate"
)

type Role struct {
    Name        string
    Permissions []Permission
}

var DefaultRoles = map[string]Role{
    "admin": {
        Name: "admin",
        Permissions: []Permission{
            TripView, TripEdit, TripDelete, TripShare, TripInvite,
            PlaceView, PlaceEdit, PlaceDelete, PlaceMedia,
            SuggestionCreate, SuggestionModerate,
        },
    },
    "editor": {
        Name: "editor",
        Permissions: []Permission{
            TripView, TripEdit, TripShare,
            PlaceView, PlaceEdit, PlaceMedia,
            SuggestionCreate,
        },
    },
    "viewer": {
        Name: "viewer",
        Permissions: []Permission{
            TripView, PlaceView, SuggestionCreate,
        },
    },
}
```

#### RBAC Middleware
```go
func RequirePermission(permission Permission) gin.HandlerFunc {
    return func(c *gin.Context) {
        userId := c.GetString("userId")
        resourceId := c.Param("id")
        resourceType := getResourceType(c.FullPath())
        
        hasPermission, err := rbac.CheckPermission(
            c.Request.Context(),
            userId,
            resourceId,
            resourceType,
            permission,
        )
        
        if err != nil || !hasPermission {
            c.JSON(403, gin.H{
                "error": "Insufficient permissions",
            })
            c.Abort()
            return
        }
        
        c.Next()
    }
}
```

### JWT Token Structure
```json
{
  "sub": "user_id",
  "email": "user@example.com",
  "roles": ["user"],
  "exp": 1234567890,
  "iat": 1234567890
}
```

## Real-time Architecture

### WebSocket Event System

#### Event Types
```typescript
enum EventType {
  // Trip events
  TRIP_CREATED = 'trip.created',
  TRIP_UPDATED = 'trip.updated',
  TRIP_DELETED = 'trip.deleted',
  TRIP_SHARED = 'trip.shared',
  
  // Place events
  PLACE_CREATED = 'place.created',
  PLACE_UPDATED = 'place.updated',
  PLACE_MEDIA_ADDED = 'place.media.added',
  
  // Suggestion events
  SUGGESTION_CREATED = 'suggestion.created',
  SUGGESTION_UPDATED = 'suggestion.updated',
  SUGGESTION_APPROVED = 'suggestion.approved',
  
  // Collaboration events
  USER_JOINED = 'user.joined',
  USER_LEFT = 'user.left',
  USER_TYPING = 'user.typing',
}
```

#### WebSocket Hub Implementation
```go
type Hub struct {
    clients    map[string]*Client
    rooms      map[string]map[*Client]bool
    broadcast  chan Message
    register   chan *Client
    unregister chan *Client
    rbac       permissions.RBAC
}

func (h *Hub) HandleMessage(client *Client, message Message) {
    // Check permissions before broadcasting
    if !h.canReceiveEvent(client, message) {
        return
    }
    
    switch message.Type {
    case "join_room":
        h.joinRoom(client, message.Room)
    case "leave_room":
        h.leaveRoom(client, message.Room)
    case "broadcast":
        h.broadcastToRoom(message.Room, message)
    }
}

func (h *Hub) canReceiveEvent(client *Client, message Message) bool {
    // Check if user has permission to receive this event
    permission := getRequiredPermission(message.Type)
    return h.rbac.HasPermission(
        client.UserID,
        message.ResourceID,
        permission,
    )
}
```

## Caching Strategy

### Redis Cache Implementation

#### Cache Keys Structure
```
# User cache
user:{userId}                    # User profile
user:{userId}:trips             # User's trips list
user:{userId}:places            # User's places
user:{userId}:permissions       # User's permissions cache

# Trip cache
trip:{tripId}                   # Trip details
trip:{tripId}:collaborators     # Trip collaborators
trip:{tripId}:waypoints         # Trip waypoints
trip:{tripId}:suggestions       # Trip suggestions

# Place cache
place:{placeId}                 # Place details
place:{placeId}:media           # Place media
places:nearby:{lat}:{lng}:{radius}  # Nearby places
places:search:{query}           # Search results

# Session cache
session:{sessionId}             # User session
```

#### Cache Strategy
```go
type CacheStrategy struct {
    redis *redis.Client
    ttl   map[string]time.Duration
}

func NewCacheStrategy(redis *redis.Client) *CacheStrategy {
    return &CacheStrategy{
        redis: redis,
        ttl: map[string]time.Duration{
            "user":         1 * time.Hour,
            "trip":         30 * time.Minute,
            "place":        1 * time.Hour,
            "search":       5 * time.Minute,
            "nearby":       15 * time.Minute,
            "permissions":  30 * time.Minute,
        },
    }
}

func (c *CacheStrategy) GetOrSet(
    ctx context.Context,
    key string,
    fn func() (interface{}, error),
) (interface{}, error) {
    // Try to get from cache
    val, err := c.redis.Get(ctx, key).Result()
    if err == nil {
        var result interface{}
        if err := json.Unmarshal([]byte(val), &result); err == nil {
            return result, nil
        }
    }
    
    // Get from source
    result, err := fn()
    if err != nil {
        return nil, err
    }
    
    // Set in cache
    data, _ := json.Marshal(result)
    ttl := c.getTTL(key)
    c.redis.Set(ctx, key, data, ttl)
    
    return result, nil
}
```

## Deployment Architecture

### Render.com Configuration

#### render.yaml
```yaml
services:
  # API Service
  - type: web
    name: trip-planner-api
    env: docker
    dockerfilePath: ./apps/api/Dockerfile
    dockerContext: .
    envVars:
      - key: PORT
        value: 8080
      - key: MONGODB_URI
        fromDatabase:
          name: trip-planner-db
          property: connectionString
      - key: REDIS_URL
        fromService:
          name: trip-planner-redis
          type: pserv
          property: connectionString
    autoDeploy: true
    healthCheckPath: /health

  # Web App
  - type: web
    name: trip-planner-web
    env: static
    buildCommand: cd apps/web && npm install && npm run build
    staticPublishPath: ./apps/web/dist
    pullRequestPreviewsEnabled: true
    headers:
      - path: /*
        name: X-Frame-Options
        value: DENY
      - path: /*
        name: X-Content-Type-Options
        value: nosniff

  # Redis Cache
  - type: pserv
    name: trip-planner-redis
    env: docker
    dockerfilePath: ./docker/redis.Dockerfile
    disk:
      name: redis-data
      mountPath: /data
      sizeGB: 10

databases:
  - name: trip-planner-db
    databaseName: trip_planner
    user: trip_planner_user
```

### Environment Variables
```env
# API Environment
NODE_ENV=production
PORT=8080

# Database
MONGODB_URI=mongodb+srv://...
REDIS_URL=redis://...

# Auth
JWT_SECRET=...
JWT_EXPIRY=7d
REFRESH_TOKEN_EXPIRY=30d

# Storage
S3_BUCKET=trip-planner-media
S3_REGION=us-east-1
S3_ACCESS_KEY=...
S3_SECRET_KEY=...
CDN_URL=https://cdn.trip-planner.com

# External APIs
MAPBOX_API_KEY=...
SENDGRID_API_KEY=...

# Monitoring
SENTRY_DSN=...
DATADOG_API_KEY=...
```

## CI/CD Pipeline

### GitHub Actions Workflow

#### .github/workflows/ci.yml
```yaml
name: CI

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main]

jobs:
  test-api:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: 1.21
      
      - name: Cache Go modules
        uses: actions/cache@v3
        with:
          path: ~/go/pkg/mod
          key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}
      
      - name: Run tests
        working-directory: ./apps/api
        run: |
          go test -v -cover ./...
          go test -race -coverprofile=coverage.txt ./...
      
      - name: Upload coverage
        uses: codecov/codecov-action@v3

  test-web:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Setup Node.js
        uses: actions/setup-node@v3
        with:
          node-version: 18
          cache: 'npm'
          cache-dependency-path: apps/web/package-lock.json
      
      - name: Install dependencies
        working-directory: ./apps/web
        run: npm ci
      
      - name: Run tests
        working-directory: ./apps/web
        run: npm run test:ci
      
      - name: Build
        working-directory: ./apps/web
        run: npm run build

  deploy:
    needs: [test-api, test-web]
    runs-on: ubuntu-latest
    if: github.ref == 'refs/heads/main'
    steps:
      - uses: actions/checkout@v3
      
      - name: Deploy to Render
        env:
          RENDER_API_KEY: ${{ secrets.RENDER_API_KEY }}
        run: |
          curl -X POST \
            -H "Authorization: Bearer $RENDER_API_KEY" \
            -H "Content-Type: application/json" \
            -d '{"clearCache": true}' \
            https://api.render.com/v1/services/${{ secrets.RENDER_SERVICE_ID }}/deploys
```

## Performance Optimization

### Backend Optimizations

1. **Database Query Optimization**
   - Use projection to limit returned fields
   - Implement cursor-based pagination
   - Aggregate pipelines for complex queries
   - Connection pooling

2. **Caching Strategy**
   - Cache frequently accessed data
   - Implement cache warming
   - Use cache tags for invalidation
   - Edge caching for static assets

3. **API Optimizations**
   - Response compression (gzip)
   - Request batching
   - Partial responses with field selection
   - HTTP/2 support

### Frontend Optimizations

1. **Bundle Optimization**
   - Code splitting by route
   - Lazy loading components
   - Tree shaking
   - Minimize bundle size

2. **Map Performance**
   - Cluster markers for large datasets
   - Viewport-based loading
   - Tile caching
   - Vector tiles for better performance

3. **State Management**
   - Normalize data structure
   - Memoization for expensive computations
   - Virtual scrolling for lists
   - Debounce/throttle user inputs

## Future ML Integration

### Architecture for ML Features

#### ML Service Architecture
```
┌─────────────────┐     ┌─────────────────┐
│   API Gateway   │────▶│   ML Gateway    │
└─────────────────┘     └────────┬────────┘
                                 │
                    ┌────────────┴────────────┐
                    │                         │
            ┌───────▼────────┐      ┌────────▼────────┐
            │ Recommendation │      │ Route Optimizer │
            │    Service     │      │     Service     │
            └───────┬────────┘      └────────┬────────┘
                    │                         │
            ┌───────▼────────┐      ┌────────▼────────┐
            │   ML Models    │      │   ML Models     │
            │  (TensorFlow)  │      │   (OR-Tools)   │
            └────────────────┘      └─────────────────┘
```

#### ML Features Roadmap

1. **Phase 1: Basic Recommendations**
   - Collaborative filtering for place recommendations
   - Content-based trip suggestions
   - Popular routes analysis

2. **Phase 2: Advanced Features**
   - Natural language trip planning
   - Automatic photo categorization
   - Sentiment analysis for reviews

3. **Phase 3: Predictive Features**
   - Traffic prediction for routes
   - Crowd prediction for popular places
   - Personalized itinerary optimization

### ML Data Pipeline
```python
# Example ML pipeline for recommendations
class PlaceRecommendationPipeline:
    def __init__(self):
        self.preprocessor = DataPreprocessor()
        self.feature_extractor = FeatureExtractor()
        self.model = load_model('place_recommender_v1')
    
    def recommend(self, user_id, context):
        # Get user history
        user_data = self.fetch_user_data(user_id)
        
        # Extract features
        features = self.feature_extractor.extract(
            user_data,
            context
        )
        
        # Generate recommendations
        predictions = self.model.predict(features)
        
        # Post-process and rank
        recommendations = self.rank_predictions(
            predictions,
            context.filters
        )
        
        return recommendations
```

## Monitoring and Observability

### Monitoring Stack
- **Application Monitoring**: DataDog APM
- **Error Tracking**: Sentry
- **Log Aggregation**: DataDog Logs
- **Uptime Monitoring**: Render built-in + StatusPage
- **Performance Monitoring**: Core Web Vitals tracking

### Key Metrics to Track
1. **API Metrics**
   - Request latency (p50, p95, p99)
   - Error rates by endpoint
   - Database query performance
   - Cache hit rates

2. **User Metrics**
   - Active users (DAU/MAU)
   - Feature adoption rates
   - User journey completion
   - Suggestion submission rates

3. **Business Metrics**
   - Trip creation rate
   - Collaboration frequency
   - Media upload volume
   - Social sharing metrics

This architecture provides a solid foundation for building a scalable, collaborative trip planning platform with robust role-based permissions and real-time features. The modular design allows for easy expansion and the addition of ML capabilities in the future.