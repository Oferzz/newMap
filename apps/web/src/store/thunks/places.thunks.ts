import { createAsyncThunk } from '@reduxjs/toolkit';
import { getDataService } from '../../services/storage/dataServiceFactory';
import { Place } from '../../types';
import {
  setPlaces,
  setSelectedPlace,
  setNearbyPlaces,
  addPlace,
  updatePlace,
  deletePlace as deletePlaceAction,
  setLoading,
  setError
} from '../slices/placesSlice';
import { addNotification } from '../slices/uiSlice';
import { RootState } from '../index';
import toast from 'react-hot-toast';

interface CreatePlaceInput {
  name: string;
  description?: string;
  category: string;
  latitude: number;
  longitude: number;
  address?: string;
  website?: string;
  phone?: string;
  notes?: string;
}

export const createPlaceThunk = createAsyncThunk<
  Place,
  CreatePlaceInput,
  { state: RootState }
>(
  'places/create',
  async (input, { dispatch, getState }) => {
    try {
      dispatch(setLoading(true));
      const dataService = getDataService();
      const isAuthenticated = getState().auth.isAuthenticated;
      
      const place = await dataService.savePlace({
        name: input.name,
        description: input.description || '',
        category: input.category,
        latitude: input.latitude,
        longitude: input.longitude,
        address: input.address || '',
        website: input.website || null,
        phone: input.phone || null,
        notes: input.notes || null,
        user_id: isAuthenticated ? getState().auth.user?.id || '' : 'guest',
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString()
      });
      
      dispatch(addPlace(place));
      dispatch(addNotification({
        type: 'success',
        message: 'Place created successfully!'
      }));
      
      return place;
    } catch (error) {
      const message = error instanceof Error ? error.message : 'Failed to create place';
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

export const getPlaceByIdThunk = createAsyncThunk(
  'places/getById',
  async (id: string, { dispatch }) => {
    try {
      dispatch(setLoading(true));
      const place = await placesService.getById(id);
      dispatch(setSelectedPlace(place as any));
      return place;
    } catch (error) {
      const message = error instanceof Error ? error.message : 'Failed to load place';
      dispatch(setError(message));
      toast.error(message);
      throw error;
    } finally {
      dispatch(setLoading(false));
    }
  }
);

export const updatePlaceThunk = createAsyncThunk(
  'places/update',
  async ({ id, input }: { id: string; input: UpdatePlaceInput }, { dispatch }) => {
    try {
      dispatch(setLoading(true));
      const place = await placesService.update(id, input);
      dispatch(updatePlace(place as any));
      toast.success('Place updated successfully!');
      return place;
    } catch (error) {
      const message = error instanceof Error ? error.message : 'Failed to update place';
      dispatch(setError(message));
      toast.error(message);
      throw error;
    } finally {
      dispatch(setLoading(false));
    }
  }
);

export const deletePlaceThunk = createAsyncThunk(
  'places/delete',
  async (id: string, { dispatch }) => {
    try {
      dispatch(setLoading(true));
      await placesService.delete(id);
      dispatch(deletePlaceAction(id));
      toast.success('Place deleted successfully');
      return id;
    } catch (error) {
      const message = error instanceof Error ? error.message : 'Failed to delete place';
      dispatch(setError(message));
      toast.error(message);
      throw error;
    } finally {
      dispatch(setLoading(false));
    }
  }
);

export const getUserPlacesThunk = createAsyncThunk(
  'places/getUserPlaces',
  async (params: { page?: number; limit?: number } | undefined, { dispatch }) => {
    try {
      dispatch(setLoading(true));
      const response = await placesService.getUserPlaces(params);
      dispatch(setPlaces(response.data as any));
      return response;
    } catch (error) {
      const message = error instanceof Error ? error.message : 'Failed to load places';
      dispatch(setError(message));
      toast.error(message);
      throw error;
    } finally {
      dispatch(setLoading(false));
    }
  }
);

export const searchPlacesThunk = createAsyncThunk(
  'places/search',
  async (input: SearchPlacesInput, { dispatch }) => {
    try {
      dispatch(setLoading(true));
      const response = await placesService.search(input);
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

export const getNearbyPlacesThunk = createAsyncThunk(
  'places/getNearby',
  async (input: NearbyPlacesInput, { dispatch }) => {
    try {
      dispatch(setLoading(true));
      const places = await placesService.getNearby(input);
      dispatch(setNearbyPlaces(places as any));
      return places;
    } catch (error) {
      const message = error instanceof Error ? error.message : 'Failed to load nearby places';
      dispatch(setError(message));
      toast.error(message);
      throw error;
    } finally {
      dispatch(setLoading(false));
    }
  }
);

export const getChildPlacesThunk = createAsyncThunk(
  'places/getChildren',
  async (parentId: string, { dispatch }) => {
    try {
      dispatch(setLoading(true));
      const places = await placesService.getChildPlaces(parentId);
      return places;
    } catch (error) {
      const message = error instanceof Error ? error.message : 'Failed to load child places';
      dispatch(setError(message));
      toast.error(message);
      throw error;
    } finally {
      dispatch(setLoading(false));
    }
  }
);

// Media thunks
export const uploadMediaThunk = createAsyncThunk(
  'places/uploadMedia',
  async ({ placeId, file, caption }: { placeId: string; file: File; caption?: string }, { dispatch }) => {
    try {
      const media = await placesService.uploadMedia(placeId, file, caption);
      toast.success('Photo uploaded successfully!');
      // Refresh place to get updated media
      dispatch(getPlaceByIdThunk(placeId));
      return media;
    } catch (error) {
      const message = error instanceof Error ? error.message : 'Failed to upload photo';
      toast.error(message);
      throw error;
    }
  }
);

export const removeMediaThunk = createAsyncThunk(
  'places/removeMedia',
  async ({ placeId, mediaId }: { placeId: string; mediaId: string }, { dispatch }) => {
    try {
      await placesService.removeMedia(placeId, mediaId);
      toast.success('Photo removed');
      // Refresh place to get updated media
      dispatch(getPlaceByIdThunk(placeId));
      return { placeId, mediaId };
    } catch (error) {
      const message = error instanceof Error ? error.message : 'Failed to remove photo';
      toast.error(message);
      throw error;
    }
  }
);