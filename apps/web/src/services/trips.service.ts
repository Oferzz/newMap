import { api, ApiResponse, PaginatedResponse } from './api';

export interface CreateTripInput {
  title: string;
  description: string;
  start_date?: string;
  end_date?: string;
  privacy?: 'public' | 'friends' | 'private';
  tags?: string[];
  cover_image?: string;
  timezone?: string;
}

export interface UpdateTripInput {
  title?: string;
  description?: string;
  start_date?: string;
  end_date?: string;
  privacy?: 'public' | 'friends' | 'private';
  status?: 'planning' | 'active' | 'completed' | 'cancelled';
  tags?: string[];
  cover_image?: string;
  timezone?: string;
}

export interface Trip {
  id: string;
  title: string;
  description: string;
  owner_id: string;
  cover_image: string;
  privacy: string;
  status: string;
  start_date?: string;
  end_date?: string;
  timezone: string;
  tags: string[];
  view_count: number;
  share_count: number;
  suggestion_count: number;
  created_at: string;
  updated_at: string;
  collaborators?: Collaborator[];
  waypoints?: Waypoint[];
}

export interface Collaborator {
  id: string;
  trip_id: string;
  user_id: string;
  role: string;
  can_edit: boolean;
  can_delete: boolean;
  can_invite: boolean;
  can_moderate_suggestions: boolean;
  invited_at: string;
  joined_at?: string;
  username?: string;
  display_name?: string;
  avatar_url?: string;
}

export interface Waypoint {
  id: string;
  trip_id: string;
  place_id: string;
  order_position: number;
  arrival_time?: string;
  departure_time?: string;
  notes: string;
  created_at: string;
  updated_at: string;
  place?: any; // Place details
}

export interface AddCollaboratorInput {
  user_id: string;
  role: 'admin' | 'editor' | 'viewer';
  can_edit?: boolean;
  can_delete?: boolean;
  can_invite?: boolean;
  can_moderate_suggestions?: boolean;
}

export interface AddWaypointInput {
  place_id: string;
  order_position: number;
  arrival_time?: string;
  departure_time?: string;
  notes?: string;
}

export interface UpdateWaypointInput {
  order_position?: number;
  arrival_time?: string;
  departure_time?: string;
  notes?: string;
}

export interface TripStats {
  total_places: number;
  total_waypoints: number;
  total_collaborators: number;
  total_suggestions: number;
  total_views: number;
  total_shares: number;
}

class TripsService {
  async create(input: CreateTripInput): Promise<Trip> {
    const response = await api.post<ApiResponse<Trip>>('/trips', input);
    return response.data;
  }

  async getById(id: string): Promise<Trip> {
    const response = await api.get<ApiResponse<Trip>>(`/trips/${id}`);
    return response.data;
  }

  async update(id: string, input: UpdateTripInput): Promise<Trip> {
    const response = await api.put<ApiResponse<Trip>>(`/trips/${id}`, input);
    return response.data;
  }

  async delete(id: string): Promise<void> {
    await api.delete(`/trips/${id}`);
  }

  async getUserTrips(params?: {
    page?: number;
    limit?: number;
    status?: string;
    privacy?: string;
  }): Promise<PaginatedResponse<Trip>> {
    const queryParams = new URLSearchParams();
    if (params?.page) queryParams.append('page', params.page.toString());
    if (params?.limit) queryParams.append('limit', params.limit.toString());
    if (params?.status) queryParams.append('status', params.status);
    if (params?.privacy) queryParams.append('privacy', params.privacy);

    const query = queryParams.toString();
    return api.get<PaginatedResponse<Trip>>(`/trips${query ? `?${query}` : ''}`);
  }

  async getSharedTrips(params?: {
    page?: number;
    limit?: number;
  }): Promise<PaginatedResponse<Trip>> {
    const queryParams = new URLSearchParams();
    if (params?.page) queryParams.append('page', params.page.toString());
    if (params?.limit) queryParams.append('limit', params.limit.toString());

    const query = queryParams.toString();
    return api.get<PaginatedResponse<Trip>>(`/trips/shared${query ? `?${query}` : ''}`);
  }

  async search(query: string, params?: {
    page?: number;
    limit?: number;
  }): Promise<PaginatedResponse<Trip>> {
    const queryParams = new URLSearchParams();
    queryParams.append('q', query);
    if (params?.page) queryParams.append('page', params.page.toString());
    if (params?.limit) queryParams.append('limit', params.limit.toString());

    return api.get<PaginatedResponse<Trip>>(`/trips/search?${queryParams.toString()}`);
  }

  // Collaborator management
  async addCollaborator(tripId: string, input: AddCollaboratorInput): Promise<void> {
    await api.post(`/trips/${tripId}/collaborators`, input);
  }

  async removeCollaborator(tripId: string, userId: string): Promise<void> {
    await api.delete(`/trips/${tripId}/collaborators/${userId}`);
  }

  async updateCollaboratorRole(tripId: string, userId: string, role: string): Promise<void> {
    await api.patch(`/trips/${tripId}/collaborators/${userId}`, { role });
  }

  // Waypoint management
  async addWaypoint(tripId: string, input: AddWaypointInput): Promise<Waypoint> {
    const response = await api.post<ApiResponse<Waypoint>>(`/trips/${tripId}/waypoints`, input);
    return response.data;
  }

  async updateWaypoint(tripId: string, waypointId: string, input: UpdateWaypointInput): Promise<Waypoint> {
    const response = await api.patch<ApiResponse<Waypoint>>(
      `/trips/${tripId}/waypoints/${waypointId}`,
      input
    );
    return response.data;
  }

  async removeWaypoint(tripId: string, waypointId: string): Promise<void> {
    await api.delete(`/trips/${tripId}/waypoints/${waypointId}`);
  }

  async reorderWaypoints(tripId: string, waypointIds: string[]): Promise<void> {
    await api.put(`/trips/${tripId}/waypoints/reorder`, { waypoint_ids: waypointIds });
  }

  // Additional features
  async getStats(tripId: string): Promise<TripStats> {
    const response = await api.get<ApiResponse<TripStats>>(`/trips/${tripId}/stats`);
    return response.data;
  }

  async export(tripId: string, format: 'json' | 'pdf' | 'ics'): Promise<Blob> {
    const response = await fetch(`${import.meta.env.VITE_API_URL || ''}/api/v1/trips/${tripId}/export?format=${format}`, {
      headers: {
        'Authorization': `Bearer ${localStorage.getItem('accessToken')}`,
      },
    });

    if (!response.ok) {
      throw new Error('Export failed');
    }

    return response.blob();
  }

  async clone(tripId: string): Promise<Trip> {
    const response = await api.post<ApiResponse<Trip>>(`/trips/${tripId}/clone`);
    return response.data;
  }
}

export const tripsService = new TripsService();