import { createAsyncThunk } from '@reduxjs/toolkit';
import { getDataService } from '../../services/storage/dataServiceFactory';
import { Collection } from '../../types';
import {
  setCollections,
  selectCollection,
  setLoading,
  setError,
  setStorageMode,
} from '../slices/collectionsSlice';
import { addNotification } from '../slices/uiSlice';
import { RootState } from '../index';

export const createCollectionThunk = createAsyncThunk<
  Collection,
  { name: string; description?: string; privacy?: string },
  { state: RootState }
>(
  'collections/create',
  async (input, { dispatch, getState }) => {
    try {
      dispatch(setLoading(true));
      const dataService = getDataService();
      const isAuthenticated = getState().auth.isAuthenticated;
      
      const collection = await dataService.saveCollection({
        name: input.name,
        description: input.description || null,
        privacy: input.privacy || 'private',
        user_id: isAuthenticated ? getState().auth.user?.id || '' : 'guest',
        locations: [],
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString()
      });
      
      // Refresh collections list
      const collections = await dataService.getCollections();
      dispatch(setCollections(collections));
      dispatch(setStorageMode(!isAuthenticated));
      
      dispatch(addNotification({
        type: 'success',
        message: 'Collection created successfully!',
      }));
      
      return collection;
    } catch (error) {
      const message = error instanceof Error ? error.message : 'Failed to create collection';
      dispatch(setError(message));
      dispatch(addNotification({
        type: 'error',
        message,
      }));
      throw error;
    } finally {
      dispatch(setLoading(false));
    }
  }
);

export const getUserCollectionsThunk = createAsyncThunk<
  Collection[],
  void,
  { state: RootState }
>(
  'collections/getUserCollections',
  async (_, { dispatch, getState }) => {
    try {
      dispatch(setLoading(true));
      const dataService = getDataService();
      const isAuthenticated = getState().auth.isAuthenticated;
      
      const collections = await dataService.getCollections();
      dispatch(setCollections(collections));
      dispatch(setStorageMode(!isAuthenticated));
      
      return collections;
    } catch (error) {
      const message = error instanceof Error ? error.message : 'Failed to load collections';
      dispatch(setError(message));
      
      // If authenticated and failed, might be network issue, show error
      if (getState().auth.isAuthenticated) {
        dispatch(addNotification({
          type: 'error',
          message,
        }));
      }
      throw error;
    } finally {
      dispatch(setLoading(false));
    }
  }
);

export const deleteCollectionThunk = createAsyncThunk<
  string,
  string,
  { state: RootState }
>(
  'collections/delete',
  async (id, { dispatch, getState }) => {
    try {
      dispatch(setLoading(true));
      const dataService = getDataService();
      
      await dataService.deleteCollection(id);
      
      // Refresh collections list
      const collections = await dataService.getCollections();
      dispatch(setCollections(collections));
      
      dispatch(addNotification({
        type: 'success',
        message: 'Collection deleted successfully',
      }));
      
      return id;
    } catch (error) {
      const message = error instanceof Error ? error.message : 'Failed to delete collection';
      dispatch(setError(message));
      dispatch(addNotification({
        type: 'error',
        message,
      }));
      throw error;
    } finally {
      dispatch(setLoading(false));
    }
  }
);

export const addLocationToCollectionThunk = createAsyncThunk<
  void,
  { collectionId: string; location: { latitude: number; longitude: number; name?: string } },
  { state: RootState }
>(
  'collections/addLocation',
  async ({ collectionId, location }, { dispatch, getState }) => {
    try {
      const dataService = getDataService();
      
      await dataService.addLocationToCollection(collectionId, location);
      
      // Refresh collections to get updated data
      const collections = await dataService.getCollections();
      dispatch(setCollections(collections));
      
      dispatch(addNotification({
        type: 'success',
        message: 'Location added to collection!',
      }));
    } catch (error) {
      const message = error instanceof Error ? error.message : 'Failed to add location';
      dispatch(addNotification({
        type: 'error',
        message,
      }));
      throw error;
    }
  }
);

export const updateCollectionThunk = createAsyncThunk<
  Collection,
  { id: string; updates: Partial<Collection> },
  { state: RootState }
>(
  'collections/update',
  async ({ id, updates }, { dispatch }) => {
    try {
      dispatch(setLoading(true));
      const dataService = getDataService();
      
      const collection = await dataService.updateCollection(id, updates);
      
      // Refresh collections list
      const collections = await dataService.getCollections();
      dispatch(setCollections(collections));
      
      dispatch(addNotification({
        type: 'success',
        message: 'Collection updated successfully!',
      }));
      
      return collection;
    } catch (error) {
      const message = error instanceof Error ? error.message : 'Failed to update collection';
      dispatch(setError(message));
      dispatch(addNotification({
        type: 'error',
        message,
      }));
      throw error;
    } finally {
      dispatch(setLoading(false));
    }
  }
);

// For authenticated-only features
export const removeLocationFromCollectionThunk = createAsyncThunk<
  void,
  { collectionId: string; locationId: string },
  { state: RootState }
>(
  'collections/removeLocation',
  async ({ collectionId, locationId }, { dispatch }) => {
    try {
      const dataService = getDataService();
      
      // For local storage, we need to implement this differently
      // since we don't have individual location removal in the interface
      const collections = await dataService.getCollections();
      const collection = collections.find(c => c.id === collectionId);
      
      if (collection) {
        collection.locations = collection.locations.filter(loc => loc.id !== locationId);
        await dataService.updateCollection(collectionId, collection);
      }
      
      // Refresh collections to get updated data
      const updatedCollections = await dataService.getCollections();
      dispatch(setCollections(updatedCollections));
      
      dispatch(addNotification({
        type: 'success',
        message: 'Location removed from collection',
      }));
    } catch (error) {
      const message = error instanceof Error ? error.message : 'Failed to remove location';
      dispatch(addNotification({
        type: 'error',
        message,
      }));
      throw error;
    }
  }
);

export const addCollaboratorThunk = createAsyncThunk<
  void,
  { collectionId: string; userId: string; role: 'viewer' | 'editor' },
  { state: RootState }
>(
  'collections/addCollaborator',
  async ({ collectionId, userId, role }, { dispatch, getState }) => {
    const isAuthenticated = getState().auth.isAuthenticated;
    
    if (!isAuthenticated) {
      dispatch(addNotification({
        type: 'error',
        message: 'Please sign in to share collections with others',
      }));
      throw new Error('Authentication required');
    }
    
    try {
      // This would only work with cloud service
      const { collectionsService } = await import('../../services/collections.service');
      await collectionsService.addCollaborator(collectionId, userId, role);
      
      dispatch(addNotification({
        type: 'success',
        message: 'Collaborator added successfully!',
      }));
    } catch (error) {
      const message = error instanceof Error ? error.message : 'Failed to add collaborator';
      dispatch(addNotification({
        type: 'error',
        message,
      }));
      throw error;
    }
  }
);

export const removeCollaboratorThunk = createAsyncThunk<
  void,
  { collectionId: string; userId: string },
  { state: RootState }
>(
  'collections/removeCollaborator',
  async ({ collectionId, userId }, { dispatch, getState }) => {
    const isAuthenticated = getState().auth.isAuthenticated;
    
    if (!isAuthenticated) {
      throw new Error('Authentication required');
    }
    
    try {
      const { collectionsService } = await import('../../services/collections.service');
      await collectionsService.removeCollaborator(collectionId, userId);
      
      dispatch(addNotification({
        type: 'success',
        message: 'Collaborator removed',
      }));
    } catch (error) {
      const message = error instanceof Error ? error.message : 'Failed to remove collaborator';
      dispatch(addNotification({
        type: 'error',
        message,
      }));
      throw error;
    }
  }
);