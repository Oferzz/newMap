import { createSlice, PayloadAction } from '@reduxjs/toolkit';
import { Collection } from '../../types';

interface CollectionsState {
  items: Collection[];
  selectedCollection: Collection | null;
  isLoading: boolean;
  error: string | null;
  isLocalStorage: boolean; // Track if data is from local storage
}

const initialState: CollectionsState = {
  items: [],
  selectedCollection: null,
  isLoading: false,
  error: null,
  isLocalStorage: false,
};

const collectionsSlice = createSlice({
  name: 'collections',
  initialState,
  reducers: {
    setStorageMode: (state, action: PayloadAction<boolean>) => {
      state.isLocalStorage = action.payload;
    },
    selectCollection: (state, action: PayloadAction<string | null>) => {
      if (action.payload === null) {
        state.selectedCollection = null;
      } else {
        state.selectedCollection = state.items.find((c: Collection) => c.id === action.payload) || null;
      }
    },
    setCollections: (state, action: PayloadAction<Collection[]>) => {
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
  setStorageMode,
  selectCollection,
  setCollections,
  setLoading,
  setError,
} = collectionsSlice.actions;

export default collectionsSlice.reducer;