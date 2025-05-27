import { createSlice, PayloadAction } from '@reduxjs/toolkit';

export interface Waypoint {
  id: string;
  day: number;
  placeId: string;
  place: {
    id: string;
    name: string;
    address: string;
    category: string;
    coordinates: { lat: number; lng: number };
  };
  arrivalTime: string;
  departureTime: string;
  notes: string;
}

export interface Trip {
  id: string;
  title: string;
  description: string;
  startDate: Date;
  endDate: Date;
  coverImage: string;
  status: 'planning' | 'active' | 'completed';
  privacy: 'public' | 'friends' | 'private';
  ownerID: string;
  collaborators: Array<{
    id: string;
    name: string;
    role: 'owner' | 'editor' | 'viewer';
    avatar: string | null;
  }>;
  waypoints: Waypoint[];
  tags: string[];
  createdAt: Date;
  updatedAt: Date;
}

interface TripsState {
  items: Trip[];
  currentTrip: Trip | null;
  isLoading: boolean;
  error: string | null;
  filters: {
    status?: string;
    privacy?: string;
    tags?: string[];
  };
}

const initialState: TripsState = {
  items: [],
  currentTrip: null,
  isLoading: false,
  error: null,
  filters: {},
};

const tripsSlice = createSlice({
  name: 'trips',
  initialState,
  reducers: {
    setTrips: (state, action: PayloadAction<Trip[]>) => {
      state.items = action.payload;
    },
    setCurrentTrip: (state, action: PayloadAction<Trip>) => {
      state.currentTrip = action.payload;
    },
    addTrip: (state, action: PayloadAction<Trip>) => {
      state.items.push(action.payload);
    },
    updateTrip: (state, action: PayloadAction<Trip>) => {
      const index = state.items.findIndex(t => t.id === action.payload.id);
      if (index !== -1) {
        state.items[index] = action.payload;
      }
      if (state.currentTrip?.id === action.payload.id) {
        state.currentTrip = action.payload;
      }
    },
    deleteTrip: (state, action: PayloadAction<string>) => {
      state.items = state.items.filter(t => t.id !== action.payload);
      if (state.currentTrip?.id === action.payload) {
        state.currentTrip = null;
      }
    },
    addWaypoint: (state, action: PayloadAction<{ tripId: string; waypoint: Waypoint }>) => {
      const trip = state.items.find(t => t.id === action.payload.tripId);
      if (trip) {
        trip.waypoints.push(action.payload.waypoint);
      }
      if (state.currentTrip?.id === action.payload.tripId) {
        state.currentTrip.waypoints.push(action.payload.waypoint);
      }
    },
    updateWaypoint: (state, action: PayloadAction<{ tripId: string; waypoint: Waypoint }>) => {
      const trip = state.items.find(t => t.id === action.payload.tripId);
      if (trip) {
        const index = trip.waypoints.findIndex(w => w.id === action.payload.waypoint.id);
        if (index !== -1) {
          trip.waypoints[index] = action.payload.waypoint;
        }
      }
    },
    removeWaypoint: (state, action: PayloadAction<{ tripId: string; waypointId: string }>) => {
      const trip = state.items.find(t => t.id === action.payload.tripId);
      if (trip) {
        trip.waypoints = trip.waypoints.filter(w => w.id !== action.payload.waypointId);
      }
    },
    reorderWaypoints: (state, action: PayloadAction<{ tripId: string; waypoints: Waypoint[] }>) => {
      const trip = state.items.find(t => t.id === action.payload.tripId);
      if (trip) {
        trip.waypoints = action.payload.waypoints;
      }
    },
    setFilters: (state, action: PayloadAction<TripsState['filters']>) => {
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
  setTrips,
  setCurrentTrip,
  addTrip,
  updateTrip,
  deleteTrip,
  addWaypoint,
  updateWaypoint,
  removeWaypoint,
  reorderWaypoints,
  setFilters,
  setLoading,
  setError,
} = tripsSlice.actions;

export default tripsSlice.reducer;