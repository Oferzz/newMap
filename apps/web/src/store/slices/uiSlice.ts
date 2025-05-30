import { createSlice, PayloadAction } from '@reduxjs/toolkit';
import { Trip } from './tripsSlice';
import { Place } from './placesSlice';
import { SearchResults } from '../../types';

interface MapViewState {
  center: [number, number];
  zoom: number;
  style?: string;
}

interface TemporaryMarker {
  id: string;
  coordinates: [number, number];
  timestamp: number;
}

interface UIState {
  activePanel: 'none' | 'details' | 'trip-planning' | 'place-creation' | 'collections';
  selectedItem: Trip | Place | null;
  mapView: MapViewState;
  searchResults: SearchResults | null;
  isSearching: boolean;
  mapClickLocation: [number, number] | null;
  temporaryMarkers: TemporaryMarker[];
  contextMenuState: {
    isOpen: boolean;
    coordinates: [number, number] | null;
    position: { x: number; y: number } | null;
  };
  routeCreationMode: {
    isActive: boolean;
    waypoints: Array<{ id: string; coordinates: [number, number] }>;
  };
  collectionsMode: {
    isAddingLocation: boolean;
    locationToAdd: [number, number] | null;
  };
  isMobileMenuOpen: boolean;
  isLoading: boolean;
  notifications: Array<{
    id: string;
    type: 'success' | 'error' | 'info' | 'warning';
    message: string;
    timestamp: number;
  }>;
}

const initialState: UIState = {
  activePanel: 'none',
  selectedItem: null,
  mapView: {
    center: [-74.5, 40],
    zoom: 9,
    style: 'mapbox://styles/mapbox/outdoors-v12',
  },
  searchResults: null,
  isSearching: false,
  mapClickLocation: null,
  temporaryMarkers: [],
  contextMenuState: {
    isOpen: false,
    coordinates: null,
    position: null,
  },
  routeCreationMode: {
    isActive: false,
    waypoints: [],
  },
  collectionsMode: {
    isAddingLocation: false,
    locationToAdd: null,
  },
  isMobileMenuOpen: false,
  isLoading: false,
  notifications: [],
};

const uiSlice = createSlice({
  name: 'ui',
  initialState,
  reducers: {
    setActivePanel: (state, action: PayloadAction<UIState['activePanel']>) => {
      state.activePanel = action.payload;
    },
    selectItem: (state, action: PayloadAction<Trip | Place | null>) => {
      state.selectedItem = action.payload;
    },
    clearSelectedItem: (state) => {
      state.selectedItem = null;
    },
    updateMapView: (state, action: PayloadAction<Partial<MapViewState>>) => {
      state.mapView = { ...state.mapView, ...action.payload };
    },
    setSearchResults: (state, action: PayloadAction<SearchResults>) => {
      state.searchResults = action.payload;
      state.isSearching = true;
    },
    setIsSearching: (state, action: PayloadAction<boolean>) => {
      state.isSearching = action.payload;
    },
    clearSearch: (state) => {
      state.searchResults = null;
      state.isSearching = false;
    },
    setMapClickLocation: (state, action: PayloadAction<{ coordinates: [number, number] }>) => {
      state.mapClickLocation = action.payload.coordinates;
    },
    clearMapClickLocation: (state) => {
      state.mapClickLocation = null;
    },
    toggleMobileMenu: (state) => {
      state.isMobileMenuOpen = !state.isMobileMenuOpen;
    },
    setLoading: (state, action: PayloadAction<boolean>) => {
      state.isLoading = action.payload;
    },
    addNotification: (state, action: PayloadAction<{
      type: 'success' | 'error' | 'info' | 'warning';
      message: string;
    }>) => {
      state.notifications.push({
        id: Date.now().toString(),
        type: action.payload.type,
        message: action.payload.message,
        timestamp: Date.now(),
      });
    },
    removeNotification: (state, action: PayloadAction<string>) => {
      state.notifications = state.notifications.filter(n => n.id !== action.payload);
    },
    addTemporaryMarker: (state, action: PayloadAction<{ coordinates: [number, number] }>) => {
      state.temporaryMarkers.push({
        id: Date.now().toString(),
        coordinates: action.payload.coordinates,
        timestamp: Date.now(),
      });
    },
    removeTemporaryMarker: (state, action: PayloadAction<string>) => {
      state.temporaryMarkers = state.temporaryMarkers.filter(m => m.id !== action.payload);
    },
    clearTemporaryMarkers: (state) => {
      state.temporaryMarkers = [];
    },
    openContextMenu: (state, action: PayloadAction<{ coordinates: [number, number]; position: { x: number; y: number } }>) => {
      state.contextMenuState = {
        isOpen: true,
        coordinates: action.payload.coordinates,
        position: action.payload.position,
      };
    },
    closeContextMenu: (state) => {
      state.contextMenuState = {
        isOpen: false,
        coordinates: null,
        position: null,
      };
    },
    startRouteCreation: (state, action: PayloadAction<{ coordinates: [number, number] }>) => {
      state.routeCreationMode = {
        isActive: true,
        waypoints: [{
          id: Date.now().toString(),
          coordinates: action.payload.coordinates,
        }],
      };
    },
    addRouteWaypoint: (state, action: PayloadAction<{ coordinates: [number, number] }>) => {
      if (state.routeCreationMode.isActive) {
        state.routeCreationMode.waypoints.push({
          id: Date.now().toString(),
          coordinates: action.payload.coordinates,
        });
      }
    },
    cancelRouteCreation: (state) => {
      state.routeCreationMode = {
        isActive: false,
        waypoints: [],
      };
    },
    finishRouteCreation: (state) => {
      state.routeCreationMode = {
        isActive: false,
        waypoints: [],
      };
      state.activePanel = 'trip-planning';
    },
    startAddToCollection: (state, action: PayloadAction<{ coordinates: [number, number] }>) => {
      state.collectionsMode = {
        isAddingLocation: true,
        locationToAdd: action.payload.coordinates,
      };
      state.activePanel = 'collections';
    },
    cancelAddToCollection: (state) => {
      state.collectionsMode = {
        isAddingLocation: false,
        locationToAdd: null,
      };
    },
  },
});

export const {
  setActivePanel,
  selectItem,
  clearSelectedItem,
  updateMapView,
  setSearchResults,
  setIsSearching,
  clearSearch,
  setMapClickLocation,
  clearMapClickLocation,
  toggleMobileMenu,
  setLoading,
  addNotification,
  removeNotification,
  addTemporaryMarker,
  removeTemporaryMarker,
  clearTemporaryMarkers,
  openContextMenu,
  closeContextMenu,
  startRouteCreation,
  addRouteWaypoint,
  cancelRouteCreation,
  finishRouteCreation,
  startAddToCollection,
  cancelAddToCollection,
} = uiSlice.actions;

export default uiSlice.reducer;