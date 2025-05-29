# Trip Planning Platform - UI Architecture

## Overview
The UI follows a minimalist, map-centric design where the map is the primary interface element, with a functional header containing search and essential controls.

## Layout Structure

```
┌─────────────────────────────────────────────────────────────┐
│                          HEADER                              │
│  [Logo] [Search Bar.....................] [User] [+ Trip]   │
└─────────────────────────────────────────────────────────────┘
│                                                              │
│                                                              │
│                                                              │
│                         MAP VIEW                             │
│                    (Full Screen Map)                         │
│                                                              │
│                                                              │
│                                                              │
│                                                              │
└──────────────────────────────────────────────────────────────┘
```

## Component Architecture

### 1. Header Component
```typescript
interface HeaderProps {
  onSearch: (query: string) => void;
  onCreateTrip: () => void;
  user: User | null;
}
```

**Features:**
- **Logo/Brand**: Clickable, returns to home/default view
- **Search Bar**: 
  - Omnisearch for places, trips, and users
  - Autocomplete suggestions
  - Search history
  - Filter options (dropdown)
- **User Menu**:
  - Profile avatar
  - Dropdown: Profile, My Trips, Settings, Logout
- **Action Buttons**:
  - Create New Trip (+)
  - Notifications (bell icon)
  - View toggle (map/list - optional)

### 2. Map Component (Main View)
```typescript
interface MapViewProps {
  center: [number, number];
  zoom: number;
  places: Place[];
  trips: Trip[];
  selectedItem: Place | Trip | null;
}
```

**Features:**
- **Full-screen Mapbox GL JS**
- **Interactive Elements**:
  - Place markers (clustered at zoom levels)
  - Trip routes (polylines)
  - Selected item highlight
- **Map Controls**:
  - Zoom in/out
  - Current location
  - Map style switcher
  - Fullscreen toggle

### 3. Overlay Components

#### Search Results Overlay
```typescript
interface SearchOverlayProps {
  results: SearchResults;
  isLoading: boolean;
  onSelectResult: (result: SearchResult) => void;
}
```
- Appears below search bar
- Semi-transparent background
- Categorized results (Places, Trips, Users)
- Keyboard navigation support

#### Place/Trip Details Panel
```typescript
interface DetailsPanelProps {
  item: Place | Trip;
  onClose: () => void;
  onEdit: () => void;
  onShare: () => void;
}
```
- Slides in from right (desktop) or bottom (mobile)
- Semi-transparent, doesn't fully cover map
- Key information display
- Action buttons (Edit, Share, Add to Trip)

#### Trip Planning Panel
```typescript
interface TripPlanningPanelProps {
  trip: Trip;
  onAddPlace: (place: Place) => void;
  onReorderPlaces: (places: Place[]) => void;
  onSave: () => void;
}
```
- Left sidebar (desktop) or bottom sheet (mobile)
- Draggable waypoint list
- Route optimization
- Trip details editing

## UI States

### 1. Default State
- Clean map view with header
- No overlays or panels
- User's current location centered (if permitted)

### 2. Search Active State
- Search overlay visible
- Map slightly dimmed
- Results update as user types

### 3. Item Selected State
- Details panel visible
- Selected item highlighted on map
- Map auto-pans to show item

### 4. Trip Planning State
- Trip planning panel visible
- Route displayed on map
- Waypoints marked and numbered

## Responsive Design

### Desktop (>1024px)
```
Header: Fixed height 64px
Map: calc(100vh - 64px)
Panels: Slide from sides, max-width 400px
```

### Tablet (768px - 1024px)
```
Header: Fixed height 56px
Map: calc(100vh - 56px)
Panels: Slide from bottom, max-height 50vh
```

### Mobile (<768px)
```
Header: Fixed height 48px, compressed layout
Map: calc(100vh - 48px)
Panels: Full-width bottom sheets
Search: Full-screen overlay
```

## Component Implementation

### Header Component
```tsx
// components/Header/Header.tsx
import React, { useState } from 'react';
import { Search, Plus, Bell, User } from 'lucide-react';
import SearchBar from './SearchBar';
import UserMenu from './UserMenu';
import NotificationBell from './NotificationBell';

interface HeaderProps {
  user: User | null;
  onSearch: (query: string) => void;
  onCreateTrip: () => void;
}

export const Header: React.FC<HeaderProps> = ({ user, onSearch, onCreateTrip }) => {
  return (
    <header className="fixed top-0 w-full h-16 bg-white border-b border-gray-200 z-50">
      <div className="flex items-center justify-between h-full px-4">
        {/* Logo */}
        <div className="flex items-center">
          <img src="/logo.svg" alt="newMap" className="h-8 w-auto" />
        </div>

        {/* Search Bar */}
        <div className="flex-1 max-w-2xl mx-4">
          <SearchBar onSearch={onSearch} />
        </div>

        {/* Action Buttons */}
        <div className="flex items-center space-x-4">
          <button
            onClick={onCreateTrip}
            className="flex items-center px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700"
          >
            <Plus className="w-4 h-4 mr-2" />
            New Trip
          </button>
          
          {user && <NotificationBell userId={user.id} />}
          
          <UserMenu user={user} />
        </div>
      </div>
    </header>
  );
};
```

### Map Container
```tsx
// components/Map/MapContainer.tsx
import React, { useRef, useEffect, useState } from 'react';
import mapboxgl from 'mapbox-gl';
import { useAppSelector } from '../../hooks/redux';

interface MapContainerProps {
  onPlaceClick: (place: Place) => void;
  onMapClick: (coordinates: [number, number]) => void;
}

export const MapContainer: React.FC<MapContainerProps> = ({ 
  onPlaceClick, 
  onMapClick 
}) => {
  const mapContainer = useRef<HTMLDivElement>(null);
  const map = useRef<mapboxgl.Map | null>(null);
  
  const places = useAppSelector(state => state.places.items);
  const selectedTrip = useAppSelector(state => state.trips.selected);

  useEffect(() => {
    if (!mapContainer.current) return;

    map.current = new mapboxgl.Map({
      container: mapContainer.current,
      style: 'mapbox://styles/mapbox/streets-v12',
      center: [-74.5, 40], // Default to NYC
      zoom: 9
    });

    // Add controls
    map.current.addControl(new mapboxgl.NavigationControl(), 'top-right');
    map.current.addControl(
      new mapboxgl.GeolocateControl({
        positionOptions: { enableHighAccuracy: true },
        trackUserLocation: true
      })
    );

    return () => map.current?.remove();
  }, []);

  return (
    <div 
      ref={mapContainer} 
      className="absolute top-16 left-0 right-0 bottom-0"
    />
  );
};
```

### Search Overlay
```tsx
// components/Search/SearchOverlay.tsx
import React from 'react';
import { Search, MapPin, Route, User } from 'lucide-react';

interface SearchOverlayProps {
  results: SearchResults;
  isVisible: boolean;
  onSelect: (result: SearchResult) => void;
  onClose: () => void;
}

export const SearchOverlay: React.FC<SearchOverlayProps> = ({
  results,
  isVisible,
  onSelect,
  onClose
}) => {
  if (!isVisible) return null;

  return (
    <>
      {/* Backdrop */}
      <div 
        className="fixed inset-0 bg-black bg-opacity-20 z-40"
        onClick={onClose}
      />
      
      {/* Results Panel */}
      <div className="absolute top-20 left-1/2 transform -translate-x-1/2 w-full max-w-2xl bg-white rounded-lg shadow-lg z-50 max-h-96 overflow-y-auto">
        {/* Places Section */}
        {results.places.length > 0 && (
          <div className="p-4 border-b">
            <h3 className="text-sm font-semibold text-gray-600 mb-2 flex items-center">
              <MapPin className="w-4 h-4 mr-2" />
              Places
            </h3>
            {results.places.map(place => (
              <div
                key={place.id}
                className="py-2 px-3 hover:bg-gray-50 cursor-pointer rounded"
                onClick={() => onSelect(place)}
              >
                <div className="font-medium">{place.name}</div>
                <div className="text-sm text-gray-600">{place.address}</div>
              </div>
            ))}
          </div>
        )}

        {/* Trips Section */}
        {results.trips.length > 0 && (
          <div className="p-4 border-b">
            <h3 className="text-sm font-semibold text-gray-600 mb-2 flex items-center">
              <Route className="w-4 h-4 mr-2" />
              Trips
            </h3>
            {results.trips.map(trip => (
              <div
                key={trip.id}
                className="py-2 px-3 hover:bg-gray-50 cursor-pointer rounded"
                onClick={() => onSelect(trip)}
              >
                <div className="font-medium">{trip.title}</div>
                <div className="text-sm text-gray-600">
                  {trip.places.length} places · {trip.collaborators.length} collaborators
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </>
  );
};
```

## State Management

### Redux Store Structure
```typescript
interface AppState {
  ui: {
    searchQuery: string;
    searchResults: SearchResults | null;
    isSearching: boolean;
    selectedItem: Place | Trip | null;
    activePanel: 'none' | 'search' | 'details' | 'trip-planning';
    mapView: {
      center: [number, number];
      zoom: number;
      style: string;
    };
  };
  trips: TripState;
  places: PlaceState;
  user: UserState;
}
```

### Key Actions
```typescript
// UI Actions
setSearchQuery(query: string)
setSearchResults(results: SearchResults)
selectItem(item: Place | Trip)
setActivePanel(panel: PanelType)
updateMapView(view: MapView)

// Trip Actions
createTrip(trip: CreateTripInput)
updateTrip(id: string, updates: UpdateTripInput)
addPlaceToTrip(tripId: string, place: Place)
reorderTripPlaces(tripId: string, placeIds: string[])
```

## Performance Optimizations

### 1. Map Performance
- Cluster markers at different zoom levels
- Lazy load place details
- Use vector tiles for better performance
- Debounce map move events

### 2. Search Performance
- Debounce search input (300ms)
- Cancel previous search requests
- Cache recent search results
- Use search indexes on backend

### 3. Component Performance
- Memoize expensive computations
- Virtual scrolling for long lists
- Lazy load images
- Code split overlay components

## Accessibility

### Keyboard Navigation
- Tab through header elements
- Arrow keys for search results
- Escape to close overlays
- Enter to select

### Screen Reader Support
- Proper ARIA labels
- Landmark regions
- Live regions for updates
- Focus management

### Color Contrast
- WCAG AA compliant
- High contrast mode support
- Color blind friendly markers

## Theme Configuration

```typescript
const theme = {
  colors: {
    primary: '#3B82F6',    // Blue
    secondary: '#10B981',  // Green
    accent: '#F59E0B',     // Amber
    background: '#FFFFFF',
    surface: '#F9FAFB',
    text: {
      primary: '#111827',
      secondary: '#6B7280',
    }
  },
  spacing: {
    headerHeight: {
      mobile: '48px',
      tablet: '56px',
      desktop: '64px',
    }
  },
  zIndex: {
    map: 1,
    header: 50,
    overlay: 40,
    panel: 45,
    modal: 60,
  }
};
```

## Mobile Considerations

### Touch Gestures
- Pinch to zoom map
- Swipe up for bottom sheets
- Pull to refresh lists
- Long press for context menu

### Offline Support
- Cache map tiles
- Store recent searches
- Queue actions when offline
- Sync when connection restored

This architecture provides a clean, map-focused interface that puts the visual exploration of places and trips at the center of the user experience while keeping essential controls easily accessible in the header.