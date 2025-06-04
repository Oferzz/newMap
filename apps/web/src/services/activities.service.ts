import { api } from './api';
import { Activity, ShareLink, ShareSettings } from '../types/activity.types';

class ActivitiesService {
  // Create a new activity
  async createActivity(activityData: any): Promise<Activity> {
    const response = await api.post<{ success: boolean; data: Activity }>('/activities', activityData);
    if (response.success) {
      return response.data!;
    }
    throw new Error('Failed to create activity');
  }

  // Get activity by ID
  async getActivity(id: string): Promise<Activity> {
    const response = await api.get<{ success: boolean; data: Activity }>(`/activities/${id}`);
    if (response.success) {
      return response.data!;
    }
    throw new Error('Failed to fetch activity');
  }

  // Update activity
  async updateActivity(id: string, updates: Partial<Activity>): Promise<Activity> {
    const response = await api.put<{ success: boolean; data: Activity }>(`/activities/${id}`, updates);
    if (response.success) {
      return response.data!;
    }
    throw new Error('Failed to update activity');
  }

  // Delete activity
  async deleteActivity(id: string): Promise<void> {
    await api.delete(`/activities/${id}`);
  }

  // List activities with filters
  async listActivities(params?: {
    page?: number;
    limit?: number;
    activityType?: string[];
    difficulty?: string[];
    privacy?: string;
  }): Promise<{ activities: Activity[]; total: number }> {
    // Build query string
    const queryParams = new URLSearchParams();
    if (params?.page) queryParams.append('page', params.page.toString());
    if (params?.limit) queryParams.append('limit', params.limit.toString());
    if (params?.privacy) queryParams.append('privacy', params.privacy);
    if (params?.activityType) {
      params.activityType.forEach(type => queryParams.append('activity_type', type));
    }
    if (params?.difficulty) {
      params.difficulty.forEach(diff => queryParams.append('difficulty', diff));
    }

    const queryString = queryParams.toString();
    const endpoint = queryString ? `/activities?${queryString}` : '/activities';
    
    const response = await api.get<{
      success: boolean;
      data: Activity[];
      meta: { total: number };
    }>(endpoint);
    
    if (response.success) {
      return {
        activities: response.data || [],
        total: response.meta?.total || 0,
      };
    }
    throw new Error('Failed to list activities');
  }

  // Like an activity
  async likeActivity(id: string): Promise<void> {
    await api.post(`/activities/${id}/like`);
  }

  // Unlike an activity
  async unlikeActivity(id: string): Promise<void> {
    await api.delete(`/activities/${id}/like`);
  }

  // Share functionality
  async generateShareLink(activityId: string, settings: ShareSettings): Promise<ShareLink> {
    const response = await api.post<{ success: boolean; data: ShareLink }>(
      `/activities/${activityId}/share`,
      settings
    );
    if (response.success) {
      return response.data!;
    }
    throw new Error('Failed to generate share link');
  }

  // Get all share links for an activity
  async getShareLinks(activityId: string): Promise<ShareLink[]> {
    const response = await api.get<{ success: boolean; data: ShareLink[] }>(
      `/activities/${activityId}/share`
    );
    if (response.success) {
      return response.data || [];
    }
    throw new Error('Failed to fetch share links');
  }

  // Revoke a share link
  async revokeShareLink(activityId: string, linkId: string): Promise<void> {
    await api.delete(`/activities/${activityId}/share/${linkId}`);
  }

  // Access shared activity
  async getSharedActivity(token: string, password?: string): Promise<Activity> {
    const queryString = password ? `?password=${encodeURIComponent(password)}` : '';
    const endpoint = `/activities/shared/${token}${queryString}`;
    
    const response = await api.get<{ success: boolean; data: Activity }>(
      endpoint,
      { skipAuth: true }
    );
    if (response.success) {
      return response.data!;
    }
    throw new Error('Failed to access shared activity');
  }
}

export const activitiesService = new ActivitiesService();