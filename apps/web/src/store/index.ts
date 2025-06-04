import { configureStore } from '@reduxjs/toolkit';
import authReducer from './slices/authSlice';
import tripsReducer from './slices/tripsSlice';
import placesReducer from './slices/placesSlice';
import uiReducer from './slices/uiSlice';
import searchReducer from './slices/searchSlice';
import collectionsReducer from './slices/collectionsSlice';
import activitiesReducer from './slices/activitiesSlice';

export const store = configureStore({
  reducer: {
    auth: authReducer,
    trips: tripsReducer,
    places: placesReducer,
    ui: uiReducer,
    search: searchReducer,
    collections: collectionsReducer,
    activities: activitiesReducer,
  },
});

export type RootState = ReturnType<typeof store.getState>;
export type AppDispatch = typeof store.dispatch;