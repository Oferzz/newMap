import { createSlice, PayloadAction } from '@reduxjs/toolkit';
import { Trip } from './tripsSlice';
import { Place } from './placesSlice';
import { SearchResults } from '../../types';

interface MapViewState {
  center: [number, number];
  zoom: number;
  style?: string;
}

interface UIState {
  activePanel: 'none' | 'details' | 'trip-planning' | 'place-creation';
  selectedItem: Trip | Place | null;
  mapView: MapViewState;
  searchResults: SearchResults | null;
  isSearching: boolean;
  mapClickLocation: [number, number] | null;
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
} = uiSlice.actions;

export default uiSlice.reducer;