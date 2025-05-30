# Authentication Strategy

This document outlines the authentication approach for the Trip Planning Platform, which follows a "freemium" model where core features are available without authentication.

## Overview

The platform implements a dual-access model:
- **Guest Mode**: Full access to core features using browser cache storage
- **Authenticated Mode**: Additional collaboration features with cloud storage

## Guest Mode (No Authentication Required)

### Features Available
- ✅ Interactive map exploration
- ✅ Place search and discovery
- ✅ Save locations to collections
- ✅ Create and plan routes
- ✅ Trip planning and itinerary creation
- ✅ Local data persistence via browser cache

### Data Storage
- **Storage Method**: Browser LocalStorage and IndexedDB
- **Persistence**: Data persists until user clears browser cache
- **Limitations**: 
  - Data not synced across devices
  - Data lost if cache is cleared
  - No sharing capabilities

### Implementation Details
```typescript
// Example local storage structure
interface LocalTripData {
  trips: Trip[];
  collections: Collection[];
  places: Place[];
  routes: Route[];
  preferences: UserPreferences;
}

// Storage keys
const STORAGE_KEYS = {
  TRIPS: 'newmap_trips',
  COLLECTIONS: 'newmap_collections', 
  PLACES: 'newmap_places',
  ROUTES: 'newmap_routes',
  PREFERENCES: 'newmap_preferences'
};
```

## Authenticated Mode

### Additional Features
- ✅ Cloud storage and sync
- ✅ Real-time collaboration
- ✅ Trip sharing with other users
- ✅ Role-based access control
- ✅ Cross-device synchronization
- ✅ Advanced analytics
- ✅ Backup and restore

### Migration from Guest to Authenticated
When a user signs up or logs in, their local data should be migrated to the cloud:

```typescript
async function migrateLocalDataToCloud(authToken: string) {
  const localData = getLocalStorageData();
  
  // Upload local trips, collections, etc. to backend
  await Promise.all([
    uploadTrips(localData.trips, authToken),
    uploadCollections(localData.collections, authToken),
    uploadPlaces(localData.places, authToken),
    uploadRoutes(localData.routes, authToken)
  ]);
  
  // Clear local storage after successful migration
  clearLocalStorage();
}
```

## Implementation Strategy

### 1. Service Layer Abstraction
Create a service abstraction that handles both local and cloud storage:

```typescript
interface DataService {
  getTrips(): Promise<Trip[]>;
  saveTrip(trip: Trip): Promise<void>;
  deleteTrip(tripId: string): Promise<void>;
  // ... other methods
}

class LocalDataService implements DataService {
  // Implementation using localStorage/IndexedDB
}

class CloudDataService implements DataService {
  // Implementation using API calls
}

// Factory pattern based on auth status
function createDataService(isAuthenticated: boolean): DataService {
  return isAuthenticated ? new CloudDataService() : new LocalDataService();
}
```

### 2. State Management
Use Redux to manage the dual storage approach:

```typescript
interface AppState {
  auth: {
    isAuthenticated: boolean;
    user: User | null;
  };
  trips: {
    items: Trip[];
    isLocal: boolean; // Track if data is local or cloud
  };
  // ... other slices
}
```

### 3. UI Indicators
Show users their current mode and data persistence status:

```typescript
// Component to show current mode
function DataModeIndicator() {
  const isAuthenticated = useSelector(state => state.auth.isAuthenticated);
  
  return (
    <div className="data-mode-indicator">
      {isAuthenticated ? (
        <span>☁️ Cloud sync active</span>
      ) : (
        <span>💾 Saved locally (sign in to sync)</span>
      )}
    </div>
  );
}
```

## Security Considerations

### Guest Mode Security
- No sensitive data storage in local storage
- All API calls for public data only
- Rate limiting on public endpoints
- Input validation and sanitization

### Authentication Security
- JWT tokens with short expiration
- Refresh token rotation
- Secure HTTP-only cookies for tokens
- CORS protection
- Rate limiting on authenticated endpoints

## Benefits of This Approach

### User Experience
- **Immediate Access**: Users can start using the platform instantly
- **No Friction**: No registration required for core features
- **Progressive Enhancement**: Natural upgrade path to authenticated features

### Business Benefits
- **Lower Barrier to Entry**: More users will try the platform
- **Conversion Funnel**: Users experience value before committing to registration
- **Retention**: Users with saved local data are more likely to register

### Technical Benefits
- **Reduced Server Load**: Guest users don't consume database resources
- **Scalability**: Core features scale independently of user accounts
- **Performance**: Local storage provides instant access to saved data

## Migration Path

### Phase 1: Current State
- All features require authentication
- Standard user registration/login flow

### Phase 2: Implement Local Storage
- Create local storage service layer
- Implement guest mode for core features
- Add data mode indicators

### Phase 3: Migration Flow
- Implement local-to-cloud data migration
- Add seamless authentication upgrade prompts
- Create data sync conflict resolution

### Phase 4: Optimization
- Optimize local storage performance
- Add offline capabilities
- Implement smart sync strategies

## Technical Implementation Notes

### Browser Storage Limits
- LocalStorage: ~5-10MB per origin
- IndexedDB: Much larger, suitable for complex trip data
- Consider using IndexedDB for large datasets

### Data Synchronization
- Implement conflict resolution for overlapping data
- Use timestamps for last-modified tracking
- Consider operational transformation for real-time sync

### Performance Considerations
- Lazy load cloud data when possible
- Cache frequently accessed data locally even when authenticated
- Implement background sync for better UX