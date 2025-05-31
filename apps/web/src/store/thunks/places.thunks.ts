import { createAsyncThunk } from '@reduxjs/toolkit';
import { getDataService } from '../../services/storage/dataServiceFactory';
import { Place } from '../../types';
import {
  addPlace,
  setLoading,
  setError
} from '../slices/placesSlice';
import { addNotification } from '../slices/uiSlice';
import { RootState } from '../index';

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
      
      const placeData: any = {
        name: input.name,
        description: input.description || '',
        category: input.category,
        location: {
          type: 'Point',
          coordinates: [input.longitude, input.latitude]
        },
        address: input.address || '',
        street_address: input.address || '',
        city: '',
        state: '',
        country: '',
        postal_code: '',
        postalCode: '',
        website: input.website || null,
        phone: input.phone || null,
        notes: input.notes || null,
        tags: [],
        images: [],
        user_id: isAuthenticated ? getState().auth.user?.id || '' : 'guest',
        createdBy: isAuthenticated ? getState().auth.user?.id || '' : 'guest',
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString(),
        createdAt: new Date(),
        updatedAt: new Date()
      };
      
      const place = await dataService.savePlace(placeData);
      
      // Convert to Redux format
      const reduxPlace: Place = {
        ...place,
        id: place.id,
        name: place.name,
        description: place.description,
        category: place.category,
        location: (place as any).location || {
          type: 'Point',
          coordinates: [input.longitude, input.latitude]
        },
        address: place.address || input.address || '',
        postalCode: (place as any).postal_code || (place as any).postalCode || '',
        images: (place as any).images || [],
        createdBy: place.created_by || (place as any).createdBy || 'guest',
        createdAt: new Date(place.created_at || (place as any).createdAt || Date.now()),
        updatedAt: new Date(place.updated_at || (place as any).updatedAt || Date.now())
      } as any;
      
      dispatch(addPlace(reduxPlace));
      dispatch(addNotification({
        type: 'success',
        message: 'Place created successfully!'
      }));
      
      return reduxPlace;
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

// TODO: Implement with data service
// export const getPlaceByIdThunk = createAsyncThunk(
//   'places/getById',
//   async (id: string, { dispatch }) => {
//     try {
//       dispatch(setLoading(true));
//       const place = await placesService.getById(id);
//       dispatch(setSelectedPlace(place as any));
//       return place;
//     } catch (error) {
//       const message = error instanceof Error ? error.message : 'Failed to load place';
//       dispatch(setError(message));
//       toast.error(message);
//       throw error;
//     } finally {
//       dispatch(setLoading(false));
//     }
//   }
// );

// TODO: Implement with data service
// export const updatePlaceThunk = createAsyncThunk(
//   'places/update',
//   async ({ id, input }: { id: string; input: UpdatePlaceInput }, { dispatch }) => {
//     try {
//       dispatch(setLoading(true));
//       const place = await placesService.update(id, input);
//       dispatch(updatePlace(place as any));
//       toast.success('Place updated successfully!');
//       return place;
//     } catch (error) {
//       const message = error instanceof Error ? error.message : 'Failed to update place';
//       dispatch(setError(message));
//       toast.error(message);
//       throw error;
//     } finally {
//       dispatch(setLoading(false));
//     }
//   }
// );

// TODO: Implement with data service
// export const deletePlaceThunk = createAsyncThunk(
//   'places/delete',
//   async (id: string, { dispatch }) => {
//     try {
//       dispatch(setLoading(true));
//       await placesService.delete(id);
//       dispatch(deletePlaceAction(id));
//       toast.success('Place deleted successfully');
//       return id;
//     } catch (error) {
//       const message = error instanceof Error ? error.message : 'Failed to delete place';
//       dispatch(setError(message));
//       toast.error(message);
//       throw error;
//     } finally {
//       dispatch(setLoading(false));
//     }
//   }
// );

// TODO: Implement remaining thunks with data service