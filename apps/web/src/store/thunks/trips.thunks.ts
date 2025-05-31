import { createAsyncThunk } from '@reduxjs/toolkit';
import { getDataService } from '../../services/storage/dataServiceFactory';
import { Trip, Waypoint } from '../slices/tripsSlice';
import { 
  setTrips, 
  setCurrentTrip, 
  addTrip, 
  updateTrip, 
  deleteTrip as deleteTripAction,
  addWaypoint,
  updateWaypoint,
  removeWaypoint,
  reorderWaypoints,
  setLoading,
  setError 
} from '../slices/tripsSlice';
import { addNotification } from '../slices/uiSlice';
import { RootState } from '../index';
import toast from 'react-hot-toast';
import { tripsService, UpdateTripInput, UpdateWaypointInput } from '../../services/trips.service';

export const createTripThunk = createAsyncThunk<
  Trip,
  { title: string; description?: string; startDate: Date; endDate: Date; privacy?: string },
  { state: RootState }
>(
  'trips/create',
  async (input, { dispatch, getState }) => {
    try {
      dispatch(setLoading(true));
      const dataService = getDataService();
      const isAuthenticated = getState().auth.isAuthenticated;
      
      const tripData: any = {
        title: input.title,
        description: input.description || '',
        startDate: input.startDate,
        endDate: input.endDate,
        start_date: input.startDate.toISOString(),
        end_date: input.endDate.toISOString(),
        coverImage: '',
        cover_image: '',
        status: 'planning',
        privacy: input.privacy || 'private',
        ownerID: isAuthenticated ? getState().auth.user?.id || '' : 'guest',
        owner_id: isAuthenticated ? getState().auth.user?.id || '' : 'guest',
        collaborators: [],
        waypoints: [],
        tags: [],
        createdAt: new Date(),
        updatedAt: new Date(),
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString()
      };
      
      const trip = await dataService.saveTrip(tripData);
      
      // Convert to Redux format
      const reduxTrip: Trip = {
        ...trip,
        id: trip.id,
        title: trip.title,
        description: trip.description,
        startDate: new Date(trip.start_date || trip.startDate),
        endDate: new Date(trip.end_date || trip.endDate),
        coverImage: trip.cover_image || trip.coverImage || '',
        status: trip.status || 'planning',
        privacy: trip.privacy || 'private',
        ownerID: trip.owner_id || trip.ownerID || 'guest',
        collaborators: trip.collaborators || [],
        waypoints: trip.waypoints || [],
        tags: trip.tags || [],
        createdAt: new Date(trip.created_at || trip.createdAt),
        updatedAt: new Date(trip.updated_at || trip.updatedAt)
      };
      
      dispatch(addTrip(reduxTrip));
      dispatch(addNotification({
        type: 'success',
        message: 'Trip created successfully!'
      }));
      
      return reduxTrip;
    } catch (error) {
      const message = error instanceof Error ? error.message : 'Failed to create trip';
      dispatch(setError(message));
      dispatch(addNotification({
        type: 'error',
        message
      }));
      throw error;
    } finally {
      dispatch(setLoading(false));
    }
  }
);

export const getTripByIdThunk = createAsyncThunk(
  'trips/getById',
  async (id: string, { dispatch }) => {
    try {
      dispatch(setLoading(true));
      const trip = await tripsService.getById(id);
      dispatch(setCurrentTrip(trip as any));
      return trip;
    } catch (error) {
      const message = error instanceof Error ? error.message : 'Failed to load trip';
      dispatch(setError(message));
      toast.error(message);
      throw error;
    } finally {
      dispatch(setLoading(false));
    }
  }
);

export const updateTripThunk = createAsyncThunk(
  'trips/update',
  async ({ id, input }: { id: string; input: UpdateTripInput }, { dispatch }) => {
    try {
      dispatch(setLoading(true));
      const trip = await tripsService.update(id, input);
      dispatch(updateTrip(trip as any));
      toast.success('Trip updated successfully!');
      return trip;
    } catch (error) {
      const message = error instanceof Error ? error.message : 'Failed to update trip';
      dispatch(setError(message));
      toast.error(message);
      throw error;
    } finally {
      dispatch(setLoading(false));
    }
  }
);

export const deleteTripThunk = createAsyncThunk(
  'trips/delete',
  async (id: string, { dispatch }) => {
    try {
      dispatch(setLoading(true));
      await tripsService.delete(id);
      dispatch(deleteTripAction(id));
      toast.success('Trip deleted successfully');
      return id;
    } catch (error) {
      const message = error instanceof Error ? error.message : 'Failed to delete trip';
      dispatch(setError(message));
      toast.error(message);
      throw error;
    } finally {
      dispatch(setLoading(false));
    }
  }
);

export const getUserTripsThunk = createAsyncThunk<
  Trip[],
  { page?: number; limit?: number; status?: string; privacy?: string } | undefined,
  { state: RootState }
>(
  'trips/getUserTrips',
  async (_params, { dispatch, getState }) => {
    try {
      dispatch(setLoading(true));
      const dataService = getDataService();
      const isAuthenticated = getState().auth.isAuthenticated;
      
      const trips = await dataService.getTrips();
      
      // Convert to Redux format
      const reduxTrips: Trip[] = trips.map((trip: any) => ({
        id: trip.id,
        title: trip.title,
        description: trip.description,
        startDate: new Date(trip.start_date || trip.startDate || trip.created_at),
        endDate: new Date(trip.end_date || trip.endDate || trip.created_at),
        coverImage: trip.cover_image || trip.coverImage || '',
        status: trip.status || 'planning',
        privacy: trip.privacy || 'private',
        ownerID: trip.owner_id || trip.ownerID || 'guest',
        collaborators: trip.collaborators || [],
        waypoints: trip.waypoints || [],
        tags: trip.tags || [],
        createdAt: new Date(trip.created_at || trip.createdAt || Date.now()),
        updatedAt: new Date(trip.updated_at || trip.updatedAt || Date.now())
      }));
      
      dispatch(setTrips(reduxTrips));
      
      // Set storage mode indicator in trips state
      if (!isAuthenticated) {
        dispatch(addNotification({
          type: 'info',
          message: 'Trips are saved locally. Sign in to sync across devices.'
        }));
      }
      
      return reduxTrips;
    } catch (error) {
      const message = error instanceof Error ? error.message : 'Failed to load trips';
      dispatch(setError(message));
      dispatch(addNotification({
        type: 'error',
        message
      }));
      throw error;
    } finally {
      dispatch(setLoading(false));
    }
  }
);

export const searchTripsThunk = createAsyncThunk(
  'trips/search',
  async ({ query, params }: { query: string; params?: { page?: number; limit?: number } }, { dispatch }) => {
    try {
      dispatch(setLoading(true));
      const response = await tripsService.search(query, params);
      return response;
    } catch (error) {
      const message = error instanceof Error ? error.message : 'Search failed';
      dispatch(setError(message));
      throw error;
    } finally {
      dispatch(setLoading(false));
    }
  }
);

// Waypoint thunks
export const addWaypointThunk = createAsyncThunk<
  Waypoint,
  { tripId: string; waypoint: Omit<Waypoint, 'id'> },
  { state: RootState }
>(
  'trips/addWaypoint',
  async ({ tripId, waypoint }, { dispatch }) => {
    try {
      const dataService = getDataService();
      const newWaypoint = await dataService.addWaypoint(tripId, waypoint);
      
      dispatch(addWaypoint({ tripId, waypoint: newWaypoint }));
      dispatch(addNotification({
        type: 'success',
        message: 'Place added to trip!'
      }));
      
      return newWaypoint;
    } catch (error) {
      const message = error instanceof Error ? error.message : 'Failed to add place';
      dispatch(addNotification({
        type: 'error',
        message
      }));
      throw error;
    }
  }
);

export const updateWaypointThunk = createAsyncThunk(
  'trips/updateWaypoint',
  async ({ tripId, waypointId, input }: { tripId: string; waypointId: string; input: UpdateWaypointInput }, { dispatch }) => {
    try {
      const waypoint = await tripsService.updateWaypoint(tripId, waypointId, input);
      dispatch(updateWaypoint({ tripId, waypoint: waypoint as any }));
      toast.success('Waypoint updated!');
      return waypoint;
    } catch (error) {
      const message = error instanceof Error ? error.message : 'Failed to update waypoint';
      toast.error(message);
      throw error;
    }
  }
);

export const removeWaypointThunk = createAsyncThunk(
  'trips/removeWaypoint',
  async ({ tripId, waypointId }: { tripId: string; waypointId: string }, { dispatch }) => {
    try {
      await tripsService.removeWaypoint(tripId, waypointId);
      dispatch(removeWaypoint({ tripId, waypointId }));
      toast.success('Place removed from trip');
      return { tripId, waypointId };
    } catch (error) {
      const message = error instanceof Error ? error.message : 'Failed to remove place';
      toast.error(message);
      throw error;
    }
  }
);

export const reorderWaypointsThunk = createAsyncThunk<
  { tripId: string; waypoints: Waypoint[] },
  { tripId: string; waypoints: Waypoint[] },
  { state: RootState }
>(
  'trips/reorderWaypoints',
  async ({ tripId, waypoints }, { dispatch }) => {
    try {
      const dataService = getDataService();
      const waypointIds = waypoints.map(w => w.id);
      
      await dataService.reorderWaypoints(tripId, waypointIds);
      dispatch(reorderWaypoints({ tripId, waypoints }));
      dispatch(addNotification({
        type: 'success',
        message: 'Itinerary reordered!'
      }));
      
      return { tripId, waypoints };
    } catch (error) {
      const message = error instanceof Error ? error.message : 'Failed to reorder itinerary';
      dispatch(addNotification({
        type: 'error',
        message
      }));
      throw error;
    }
  }
);

// Collaborator thunks
export const inviteCollaboratorThunk = createAsyncThunk(
  'trips/inviteCollaborator',
  async ({ tripId, input }: { tripId: string; input: any }, { dispatch }) => {
    try {
      await tripsService.addCollaborator(tripId, input);
      toast.success('Collaborator invited!');
      // Refresh trip to get updated collaborators
      dispatch(getTripByIdThunk(tripId));
    } catch (error) {
      const message = error instanceof Error ? error.message : 'Failed to invite collaborator';
      toast.error(message);
      throw error;
    }
  }
);

export const removeCollaboratorThunk = createAsyncThunk(
  'trips/removeCollaborator',
  async ({ tripId, userId }: { tripId: string; userId: string }, { dispatch }) => {
    try {
      await tripsService.removeCollaborator(tripId, userId);
      toast.success('Collaborator removed');
      // Refresh trip to get updated collaborators
      dispatch(getTripByIdThunk(tripId));
    } catch (error) {
      const message = error instanceof Error ? error.message : 'Failed to remove collaborator';
      toast.error(message);
      throw error;
    }
  }
);