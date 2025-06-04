import { createAsyncThunk } from '@reduxjs/toolkit';
import { RootState } from '../index';
import { Activity } from '../slices/activitiesSlice';
import { ActivityFormData } from '../../pages/ActivityCreationPage';
import { api } from '../../services/api';
import toast from 'react-hot-toast';

// API interfaces
interface CreateActivityInput {
  title: string;
  description: string;
  activityType: string;
  route?: {
    type: 'out-and-back' | 'loop' | 'point-to-point';
    waypoints: Array<{ lat: number; lng: number; elevation?: number }>;
    distance?: number;
    elevationGain?: number;
    elevationLoss?: number;
  };
  metadata: {
    difficulty: 'easy' | 'moderate' | 'hard' | 'expert';
    duration: number;
    distance: number;
    elevationGain: number;
    terrain: string[];
    waterFeatures: string[];
    gear: string[];
    seasons: string[];
    conditions: string[];
    tags: string[];
  };
  visibility: {
    privacy: 'public' | 'friends' | 'private';
    allowComments: boolean;
    allowDownloads: boolean;
    shareWithGroups: string[];
  };
}

interface GetActivitiesParams {
  page?: number;
  limit?: number;
  filters?: {
    activityType?: string[];
    difficulty?: string[];
    duration?: { min: number; max: number };
    distance?: { min: number; max: number };
    terrain?: string[];
    search?: string;
  };
  userId?: string; // For getting user-specific activities
}

interface ActivitiesResponse {
  activities: Activity[];
  pagination: {
    page: number;
    limit: number;
    total: number;
    hasMore: boolean;
  };
}

// Transform ActivityFormData to API format
const transformFormDataToAPI = (formData: ActivityFormData): CreateActivityInput => {
  return {
    title: formData.title,
    description: formData.description,
    activityType: formData.activityType,
    route: formData.route,
    metadata: formData.metadata,
    visibility: formData.visibility,
  };
};

// Create Activity
export const createActivityThunk = createAsyncThunk<
  Activity,
  ActivityFormData,
  { state: RootState }
>(
  'activities/create',
  async (formData, { dispatch, getState, rejectWithValue }) => {
    try {
      const state = getState();
      const isAuthenticated = state.auth.isAuthenticated;

      if (!isAuthenticated) {
        toast.error('Please log in to create activities');
        throw new Error('Authentication required');
      }

      const activityData = transformFormDataToAPI(formData);
      
      const response = await api.post('/activities', activityData);
      
      if (response.success) {
        toast.success('Activity created successfully!');
        return response.data;
      } else {
        throw new Error(response.error?.message || 'Failed to create activity');
      }
    } catch (error: any) {
      const message = error.response?.data?.error?.message || error.message || 'Failed to create activity';
      toast.error(message);
      return rejectWithValue(message);
    }
  }
);

// Get Activities (with filters and pagination)
export const getActivitiesThunk = createAsyncThunk<
  ActivitiesResponse,
  GetActivitiesParams,
  { state: RootState }
>(
  'activities/getAll',
  async (params = {}, { rejectWithValue }) => {
    try {
      const queryParams = new URLSearchParams();
      
      if (params.page) queryParams.append('page', params.page.toString());
      if (params.limit) queryParams.append('limit', params.limit.toString());
      if (params.userId) queryParams.append('userId', params.userId);
      
      // Add filters
      if (params.filters) {
        const { filters } = params;
        if (filters.search) queryParams.append('search', filters.search);
        if (filters.activityType?.length) {
          filters.activityType.forEach(type => queryParams.append('activityType', type));
        }
        if (filters.difficulty?.length) {
          filters.difficulty.forEach(diff => queryParams.append('difficulty', diff));
        }
        if (filters.terrain?.length) {
          filters.terrain.forEach(terrain => queryParams.append('terrain', terrain));
        }
        if (filters.duration) {
          queryParams.append('minDuration', filters.duration.min.toString());
          queryParams.append('maxDuration', filters.duration.max.toString());
        }
        if (filters.distance) {
          queryParams.append('minDistance', filters.distance.min.toString());
          queryParams.append('maxDistance', filters.distance.max.toString());
        }
      }

      const response = await api.get(`/activities?${queryParams.toString()}`);
      
      if (response.success) {
        return {
          activities: response.data.activities || [],
          pagination: response.data.pagination || {
            page: 1,
            limit: 20,
            total: 0,
            hasMore: false,
          },
        };
      } else {
        throw new Error(response.error?.message || 'Failed to fetch activities');
      }
    } catch (error: any) {
      const message = error.response?.data?.error?.message || error.message || 'Failed to fetch activities';
      return rejectWithValue(message);
    }
  }
);

// Get Single Activity
export const getActivityThunk = createAsyncThunk<
  Activity,
  string,
  { state: RootState }
>(
  'activities/getOne',
  async (activityId, { rejectWithValue }) => {
    try {
      const response = await api.get(`/activities/${activityId}`);
      
      if (response.success) {
        return response.data;
      } else {
        throw new Error(response.error?.message || 'Failed to fetch activity');
      }
    } catch (error: any) {
      const message = error.response?.data?.error?.message || error.message || 'Failed to fetch activity';
      return rejectWithValue(message);
    }
  }
);

// Update Activity
export const updateActivityThunk = createAsyncThunk<
  Activity,
  { id: string; updates: Partial<CreateActivityInput> },
  { state: RootState }
>(
  'activities/update',
  async ({ id, updates }, { rejectWithValue }) => {
    try {
      const response = await api.put(`/activities/${id}`, updates);
      
      if (response.success) {
        toast.success('Activity updated successfully!');
        return response.data;
      } else {
        throw new Error(response.error?.message || 'Failed to update activity');
      }
    } catch (error: any) {
      const message = error.response?.data?.error?.message || error.message || 'Failed to update activity';
      toast.error(message);
      return rejectWithValue(message);
    }
  }
);

// Delete Activity
export const deleteActivityThunk = createAsyncThunk<
  string,
  string,
  { state: RootState }
>(
  'activities/delete',
  async (activityId, { rejectWithValue }) => {
    try {
      const response = await api.delete(`/activities/${activityId}`);
      
      if (response.success) {
        toast.success('Activity deleted successfully');
        return activityId;
      } else {
        throw new Error(response.error?.message || 'Failed to delete activity');
      }
    } catch (error: any) {
      const message = error.response?.data?.error?.message || error.message || 'Failed to delete activity';
      toast.error(message);
      return rejectWithValue(message);
    }
  }
);

// Like Activity
export const likeActivityThunk = createAsyncThunk<
  { activityId: string; liked: boolean },
  string,
  { state: RootState }
>(
  'activities/like',
  async (activityId, { rejectWithValue }) => {
    try {
      const response = await api.post(`/activities/${activityId}/like`);
      
      if (response.success) {
        return { activityId, liked: true };
      } else {
        throw new Error(response.error?.message || 'Failed to like activity');
      }
    } catch (error: any) {
      const message = error.response?.data?.error?.message || error.message || 'Failed to like activity';
      return rejectWithValue(message);
    }
  }
);

// Unlike Activity
export const unlikeActivityThunk = createAsyncThunk<
  { activityId: string; liked: boolean },
  string,
  { state: RootState }
>(
  'activities/unlike',
  async (activityId, { rejectWithValue }) => {
    try {
      const response = await api.delete(`/activities/${activityId}/like`);
      
      if (response.success) {
        return { activityId, liked: false };
      } else {
        throw new Error(response.error?.message || 'Failed to unlike activity');
      }
    } catch (error: any) {
      const message = error.response?.data?.error?.message || error.message || 'Failed to unlike activity';
      return rejectWithValue(message);
    }
  }
);