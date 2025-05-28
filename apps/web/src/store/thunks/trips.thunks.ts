import { createAsyncThunk } from '@reduxjs/toolkit';
import { tripsService, CreateTripInput, UpdateTripInput, AddWaypointInput, UpdateWaypointInput } from '../../services/trips.service';
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
import toast from 'react-hot-toast';

export const createTripThunk = createAsyncThunk(
  'trips/create',
  async (input: CreateTripInput, { dispatch }) => {
    try {
      dispatch(setLoading(true));
      const trip = await tripsService.create(input);
      dispatch(addTrip(trip as any)); // Type conversion for now
      toast.success('Trip created successfully!');
      return trip;
    } catch (error) {
      const message = error instanceof Error ? error.message : 'Failed to create trip';
      dispatch(setError(message));
      toast.error(message);
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

export const getUserTripsThunk = createAsyncThunk(
  'trips/getUserTrips',
  async (params: { page?: number; limit?: number; status?: string; privacy?: string } | undefined, { dispatch }) => {
    try {
      dispatch(setLoading(true));
      const response = await tripsService.getUserTrips(params);
      dispatch(setTrips(response.data as any));
      return response;
    } catch (error) {
      const message = error instanceof Error ? error.message : 'Failed to load trips';
      dispatch(setError(message));
      toast.error(message);
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
export const addWaypointThunk = createAsyncThunk(
  'trips/addWaypoint',
  async ({ tripId, input }: { tripId: string; input: AddWaypointInput }, { dispatch }) => {
    try {
      const waypoint = await tripsService.addWaypoint(tripId, input);
      dispatch(addWaypoint({ tripId, waypoint: waypoint as any }));
      toast.success('Place added to trip!');
      return waypoint;
    } catch (error) {
      const message = error instanceof Error ? error.message : 'Failed to add place';
      toast.error(message);
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

export const reorderWaypointsThunk = createAsyncThunk(
  'trips/reorderWaypoints',
  async ({ tripId, waypoints }: { tripId: string; waypoints: any[] }, { dispatch }) => {
    try {
      const waypointIds = waypoints.map(w => w.id);
      await tripsService.reorderWaypoints(tripId, waypointIds);
      dispatch(reorderWaypoints({ tripId, waypoints }));
      toast.success('Itinerary reordered!');
      return { tripId, waypoints };
    } catch (error) {
      const message = error instanceof Error ? error.message : 'Failed to reorder itinerary';
      toast.error(message);
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