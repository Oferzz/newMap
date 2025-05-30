import { createSlice, PayloadAction } from '@reduxjs/toolkit';

export interface LocationCollection {
  id: string;
  name: string;
  description?: string;
  locations: Array<{
    id: string;
    coordinates: [number, number];
    name?: string;
    addedAt: number;
  }>;
  createdAt: number;
  updatedAt: number;
}

interface CollectionsState {
  items: LocationCollection[];
  selectedCollection: LocationCollection | null;
  isLoading: boolean;
  error: string | null;
}

const initialState: CollectionsState = {
  items: [],
  selectedCollection: null,
  isLoading: false,
  error: null,
};

const collectionsSlice = createSlice({
  name: 'collections',
  initialState,
  reducers: {
    createCollection: (state, action: PayloadAction<{ name: string; description?: string }>) => {
      const newCollection: LocationCollection = {
        id: Date.now().toString(),
        name: action.payload.name,
        description: action.payload.description,
        locations: [],
        createdAt: Date.now(),
        updatedAt: Date.now(),
      };
      state.items.push(newCollection);
    },
    addLocationToCollection: (state, action: PayloadAction<{ 
      collectionId: string; 
      coordinates: [number, number];
      name?: string;
    }>) => {
      const collection = state.items.find(c => c.id === action.payload.collectionId);
      if (collection) {
        collection.locations.push({
          id: Date.now().toString(),
          coordinates: action.payload.coordinates,
          name: action.payload.name,
          addedAt: Date.now(),
        });
        collection.updatedAt = Date.now();
      }
    },
    removeLocationFromCollection: (state, action: PayloadAction<{ 
      collectionId: string; 
      locationId: string;
    }>) => {
      const collection = state.items.find(c => c.id === action.payload.collectionId);
      if (collection) {
        collection.locations = collection.locations.filter(
          loc => loc.id !== action.payload.locationId
        );
        collection.updatedAt = Date.now();
      }
    },
    deleteCollection: (state, action: PayloadAction<string>) => {
      state.items = state.items.filter(c => c.id !== action.payload);
      if (state.selectedCollection?.id === action.payload) {
        state.selectedCollection = null;
      }
    },
    selectCollection: (state, action: PayloadAction<string | null>) => {
      if (action.payload === null) {
        state.selectedCollection = null;
      } else {
        state.selectedCollection = state.items.find(c => c.id === action.payload) || null;
      }
    },
    setCollections: (state, action: PayloadAction<LocationCollection[]>) => {
      state.items = action.payload;
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
  createCollection,
  addLocationToCollection,
  removeLocationFromCollection,
  deleteCollection,
  selectCollection,
  setCollections,
  setLoading,
  setError,
} = collectionsSlice.actions;

export default collectionsSlice.reducer;