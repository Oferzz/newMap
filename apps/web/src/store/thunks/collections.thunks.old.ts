import { createAsyncThunk } from '@reduxjs/toolkit';
import { 
  collectionsService, 
  CreateCollectionInput, 
  UpdateCollectionInput, 
  AddLocationInput,
  GetCollectionsParams 
} from '../../services/collections.service';
import {
  setCollections,
  selectCollection,
  setLoading,
  setError,
} from '../slices/collectionsSlice';
import { addNotification } from '../slices/uiSlice';

export const createCollectionThunk = createAsyncThunk(
  'collections/create',
  async (input: CreateCollectionInput, { dispatch }) => {
    try {
      dispatch(setLoading(true));
      const collection = await collectionsService.createCollection(input);
      // Refresh collections list
      const response = await collectionsService.getUserCollections();
      dispatch(setCollections(response.data));
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

export const getCollectionThunk = createAsyncThunk(
  'collections/getById',
  async (id: string, { dispatch }) => {
    try {
      dispatch(setLoading(true));
      const collection = await collectionsService.getCollection(id);
      dispatch(selectCollection(collection.id));
      return collection;
    } catch (error) {
      const message = error instanceof Error ? error.message : 'Failed to load collection';
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

export const getUserCollectionsThunk = createAsyncThunk(
  'collections/getUserCollections',
  async (params: GetCollectionsParams | undefined, { dispatch }) => {
    try {
      dispatch(setLoading(true));
      const response = await collectionsService.getUserCollections(params);
      dispatch(setCollections(response.data));
      return response;
    } catch (error) {
      const message = error instanceof Error ? error.message : 'Failed to load collections';
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

export const updateCollectionThunk = createAsyncThunk(
  'collections/update',
  async ({ id, input }: { id: string; input: UpdateCollectionInput }, { dispatch }) => {
    try {
      dispatch(setLoading(true));
      const collection = await collectionsService.updateCollection(id, input);
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

export const deleteCollectionThunk = createAsyncThunk(
  'collections/delete',
  async (id: string, { dispatch }) => {
    try {
      dispatch(setLoading(true));
      await collectionsService.deleteCollection(id);
      // Refresh collections list
      const response = await collectionsService.getUserCollections();
      dispatch(setCollections(response.data));
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

export const addLocationToCollectionThunk = createAsyncThunk(
  'collections/addLocation',
  async ({ collectionId, location }: { collectionId: string; location: AddLocationInput }, { dispatch }) => {
    try {
      const result = await collectionsService.addLocationToCollection(collectionId, location);
      dispatch(addNotification({
        type: 'success',
        message: 'Location added to collection!',
      }));
      
      // Refresh the collection to get updated locations
      dispatch(getCollectionThunk(collectionId));
      return result;
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

export const removeLocationFromCollectionThunk = createAsyncThunk(
  'collections/removeLocation',
  async ({ collectionId, locationId }: { collectionId: string; locationId: string }, { dispatch }) => {
    try {
      await collectionsService.removeLocationFromCollection(collectionId, locationId);
      dispatch(addNotification({
        type: 'success',
        message: 'Location removed from collection',
      }));
      
      // Refresh the collection to get updated locations
      dispatch(getCollectionThunk(collectionId));
      return { collectionId, locationId };
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

export const addCollaboratorThunk = createAsyncThunk(
  'collections/addCollaborator',
  async ({ collectionId, userId, role }: { collectionId: string; userId: string; role: 'viewer' | 'editor' }, { dispatch }) => {
    try {
      await collectionsService.addCollaborator(collectionId, userId, role);
      dispatch(addNotification({
        type: 'success',
        message: 'Collaborator added successfully!',
      }));
      return { collectionId, userId, role };
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

export const removeCollaboratorThunk = createAsyncThunk(
  'collections/removeCollaborator',
  async ({ collectionId, userId }: { collectionId: string; userId: string }, { dispatch }) => {
    try {
      await collectionsService.removeCollaborator(collectionId, userId);
      dispatch(addNotification({
        type: 'success',
        message: 'Collaborator removed',
      }));
      return { collectionId, userId };
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