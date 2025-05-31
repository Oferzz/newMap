import { createSlice, PayloadAction } from '@reduxjs/toolkit';
import { Place } from '../../types';

interface PlacesState {
  items: Place[];
  selectedPlace: Place | null;
  nearbyPlaces: Place[];
  isLoading: boolean;
  error: string | null;
  filters: {
    category?: string[];
    tags?: string[];
    rating?: number;
  };
}

const initialState: PlacesState = {
  items: [],
  selectedPlace: null,
  nearbyPlaces: [],
  isLoading: false,
  error: null,
  filters: {},
};

const placesSlice = createSlice({
  name: 'places',
  initialState,
  reducers: {
    setPlaces: (state, action: PayloadAction<Place[]>) => {
      state.items = action.payload;
    },
    setSelectedPlace: (state, action: PayloadAction<Place | null>) => {
      state.selectedPlace = action.payload;
    },
    setNearbyPlaces: (state, action: PayloadAction<Place[]>) => {
      state.nearbyPlaces = action.payload;
    },
    addPlace: (state, action: PayloadAction<Place>) => {
      state.items.push(action.payload);
    },
    updatePlace: (state, action: PayloadAction<Place>) => {
      const index = state.items.findIndex(p => p.id === action.payload.id);
      if (index !== -1) {
        state.items[index] = action.payload;
      }
      if (state.selectedPlace?.id === action.payload.id) {
        state.selectedPlace = action.payload;
      }
    },
    deletePlace: (state, action: PayloadAction<string>) => {
      state.items = state.items.filter(p => p.id !== action.payload);
      if (state.selectedPlace?.id === action.payload) {
        state.selectedPlace = null;
      }
    },
    setFilters: (state, action: PayloadAction<PlacesState['filters']>) => {
      state.filters = action.payload;
    },
    setLoading: (state, action: PayloadAction<boolean>) => {
      state.isLoading = action.payload;
    },
    setError: (state, action: PayloadAction<string | null>) => {
      state.error = action.payload;
    },
  },
});

export const {
  setPlaces,
  setSelectedPlace,
  setNearbyPlaces,
  addPlace,
  updatePlace,
  deletePlace,
  setFilters,
  setLoading,
  setError,
} = placesSlice.actions;

export default placesSlice.reducer;