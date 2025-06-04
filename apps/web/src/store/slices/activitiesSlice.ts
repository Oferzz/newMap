import { createSlice, PayloadAction } from '@reduxjs/toolkit';
import { createActivityThunk, getActivitiesThunk, getActivityThunk, updateActivityThunk, deleteActivityThunk } from '../thunks/activities.thunks';

export interface Activity {
  id: string;
  title: string;
  description: string;
  activityType: string;
  route?: {
    type: 'out-and-back' | 'loop' | 'point-to-point';
    waypoints: Array<{ lat: number; lng: number; elevation?: number }>;
    distance?: number;
    elevationGain?: number;
    elevationLoss?: number;
    geoJSON?: any;
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
  stats: {
    views: number;
    likes: number;
    comments: number;
    downloads: number;
  };
  creator: {
    id: string;
    name: string;
    avatar?: string;
  };
  createdAt: string;
  updatedAt: string;
  coverImage?: string;
  shareLink?: string;
}

export interface ActivitiesState {
  activities: Activity[];
  currentActivity: Activity | null;
  userActivities: Activity[];
  featuredActivities: Activity[];
  isLoading: boolean;
  isCreating: boolean;
  isUpdating: boolean;
  isDeleting: boolean;
  error: string | null;
  filters: {
    activityType: string[];
    difficulty: string[];
    duration: { min: number; max: number };
    distance: { min: number; max: number };
    terrain: string[];
    search: string;
  };
  pagination: {
    page: number;
    limit: number;
    total: number;
    hasMore: boolean;
  };
}

const initialState: ActivitiesState = {
  activities: [],
  currentActivity: null,
  userActivities: [],
  featuredActivities: [],
  isLoading: false,
  isCreating: false,
  isUpdating: false,
  isDeleting: false,
  error: null,
  filters: {
    activityType: [],
    difficulty: [],
    duration: { min: 0, max: 24 },
    distance: { min: 0, max: 100 },
    terrain: [],
    search: '',
  },
  pagination: {
    page: 1,
    limit: 20,
    total: 0,
    hasMore: false,
  },
};

const activitiesSlice = createSlice({
  name: 'activities',
  initialState,
  reducers: {
    clearError: (state) => {
      state.error = null;
    },
    
    setCurrentActivity: (state, action: PayloadAction<Activity | null>) => {
      state.currentActivity = action.payload;
    },
    
    updateFilters: (state, action: PayloadAction<Partial<ActivitiesState['filters']>>) => {
      state.filters = { ...state.filters, ...action.payload };
      // Reset pagination when filters change
      state.pagination.page = 1;
    },
    
    clearFilters: (state) => {
      state.filters = initialState.filters;
      state.pagination.page = 1;
    },
    
    updateActivityStats: (state, action: PayloadAction<{ id: string; stats: Partial<Activity['stats']> }>) => {
      const { id, stats } = action.payload;
      
      // Update in activities array
      const activityIndex = state.activities.findIndex(a => a.id === id);
      if (activityIndex !== -1) {
        state.activities[activityIndex].stats = { 
          ...state.activities[activityIndex].stats, 
          ...stats 
        };
      }
      
      // Update current activity if it matches
      if (state.currentActivity?.id === id) {
        state.currentActivity.stats = { 
          ...state.currentActivity.stats, 
          ...stats 
        };
      }
      
      // Update in user activities
      const userActivityIndex = state.userActivities.findIndex(a => a.id === id);
      if (userActivityIndex !== -1) {
        state.userActivities[userActivityIndex].stats = { 
          ...state.userActivities[userActivityIndex].stats, 
          ...stats 
        };
      }
    },
    
    likeActivity: (state, action: PayloadAction<string>) => {
      const activityId = action.payload;
      // This would typically also track if the current user has liked it
      // For now, just increment the count
      const activity = state.activities.find(a => a.id === activityId);
      if (activity) {
        activity.stats.likes += 1;
      }
      
      if (state.currentActivity?.id === activityId) {
        state.currentActivity.stats.likes += 1;
      }
    },
    
    unlikeActivity: (state, action: PayloadAction<string>) => {
      const activityId = action.payload;
      const activity = state.activities.find(a => a.id === activityId);
      if (activity && activity.stats.likes > 0) {
        activity.stats.likes -= 1;
      }
      
      if (state.currentActivity?.id === activityId && state.currentActivity.stats.likes > 0) {
        state.currentActivity.stats.likes -= 1;
      }
    },
    
    incrementViews: (state, action: PayloadAction<string>) => {
      const activityId = action.payload;
      const activity = state.activities.find(a => a.id === activityId);
      if (activity) {
        activity.stats.views += 1;
      }
      
      if (state.currentActivity?.id === activityId) {
        state.currentActivity.stats.views += 1;
      }
    },
  },
  
  extraReducers: (builder) => {
    // Get Activities
    builder
      .addCase(getActivitiesThunk.pending, (state) => {
        state.isLoading = true;
        state.error = null;
      })
      .addCase(getActivitiesThunk.fulfilled, (state, action) => {
        state.isLoading = false;
        const { activities, pagination } = action.payload;
        
        if (pagination.page === 1) {
          state.activities = activities;
        } else {
          // Append for pagination
          state.activities.push(...activities);
        }
        
        state.pagination = pagination;
      })
      .addCase(getActivitiesThunk.rejected, (state, action) => {
        state.isLoading = false;
        state.error = action.error.message || 'Failed to fetch activities';
      });

    // Get Single Activity
    builder
      .addCase(getActivityThunk.pending, (state) => {
        state.isLoading = true;
        state.error = null;
      })
      .addCase(getActivityThunk.fulfilled, (state, action) => {
        state.isLoading = false;
        state.currentActivity = action.payload;
      })
      .addCase(getActivityThunk.rejected, (state, action) => {
        state.isLoading = false;
        state.error = action.error.message || 'Failed to fetch activity';
      });

    // Create Activity
    builder
      .addCase(createActivityThunk.pending, (state) => {
        state.isCreating = true;
        state.error = null;
      })
      .addCase(createActivityThunk.fulfilled, (state, action) => {
        state.isCreating = false;
        const newActivity = action.payload;
        state.activities.unshift(newActivity);
        state.userActivities.unshift(newActivity);
        state.currentActivity = newActivity;
      })
      .addCase(createActivityThunk.rejected, (state, action) => {
        state.isCreating = false;
        state.error = action.error.message || 'Failed to create activity';
      });

    // Update Activity
    builder
      .addCase(updateActivityThunk.pending, (state) => {
        state.isUpdating = true;
        state.error = null;
      })
      .addCase(updateActivityThunk.fulfilled, (state, action) => {
        state.isUpdating = false;
        const updatedActivity = action.payload;
        
        // Update in activities array
        const index = state.activities.findIndex(a => a.id === updatedActivity.id);
        if (index !== -1) {
          state.activities[index] = updatedActivity;
        }
        
        // Update current activity
        if (state.currentActivity?.id === updatedActivity.id) {
          state.currentActivity = updatedActivity;
        }
        
        // Update in user activities
        const userIndex = state.userActivities.findIndex(a => a.id === updatedActivity.id);
        if (userIndex !== -1) {
          state.userActivities[userIndex] = updatedActivity;
        }
      })
      .addCase(updateActivityThunk.rejected, (state, action) => {
        state.isUpdating = false;
        state.error = action.error.message || 'Failed to update activity';
      });

    // Delete Activity
    builder
      .addCase(deleteActivityThunk.pending, (state) => {
        state.isDeleting = true;
        state.error = null;
      })
      .addCase(deleteActivityThunk.fulfilled, (state, action) => {
        state.isDeleting = false;
        const deletedId = action.payload;
        
        // Remove from activities array
        state.activities = state.activities.filter(a => a.id !== deletedId);
        
        // Remove from user activities
        state.userActivities = state.userActivities.filter(a => a.id !== deletedId);
        
        // Clear current activity if it was deleted
        if (state.currentActivity?.id === deletedId) {
          state.currentActivity = null;
        }
      })
      .addCase(deleteActivityThunk.rejected, (state, action) => {
        state.isDeleting = false;
        state.error = action.error.message || 'Failed to delete activity';
      });
  },
});

export const {
  clearError,
  setCurrentActivity,
  updateFilters,
  clearFilters,
  updateActivityStats,
  likeActivity,
  unlikeActivity,
  incrementViews,
} = activitiesSlice.actions;

export default activitiesSlice.reducer;