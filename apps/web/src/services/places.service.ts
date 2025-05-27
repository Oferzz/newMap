import { api, ApiResponse, PaginatedResponse } from './api';

export interface CreatePlaceInput {
  name: string;
  description?: string;
  type: 'poi' | 'area' | 'region';
  parent_id?: string;
  location?: {
    latitude: number;
    longitude: number;
  };
  bounds?: {
    coordinates: number[][][];
  };
  street_address?: string;
  city?: string;
  state?: string;
  country?: string;
  postal_code?: string;
  category?: string[];
  tags?: string[];
  opening_hours?: any;
  contact_info?: {
    phone?: string;
    email?: string;
    website?: string;
  };
  amenities?: string[];
  privacy?: 'public' | 'friends' | 'private';
}

export interface UpdatePlaceInput {
  name?: string;
  description?: string;
  type?: 'poi' | 'area' | 'region';
  location?: {
    latitude: number;
    longitude: number;
  };
  bounds?: {
    coordinates: number[][][];
  };
  street_address?: string;
  city?: string;
  state?: string;
  country?: string;
  postal_code?: string;
  category?: string[];
  tags?: string[];
  opening_hours?: any;
  contact_info?: {
    phone?: string;
    email?: string;
    website?: string;
  };
  amenities?: string[];
  privacy?: 'public' | 'friends' | 'private';
  status?: 'active' | 'pending' | 'archived';
}

export interface Place {
  id: string;
  name: string;
  description: string;
  type: string;
  parent_id?: string;
  location?: {
    type: string;
    coordinates: [number, number];
  };
  bounds?: any;
  street_address: string;
  city: string;
  state: string;
  country: string;
  postal_code: string;
  created_by: string;
  category: string[];
  tags: string[];
  opening_hours?: any;
  contact_info?: any;
  amenities: string[];
  average_rating?: number;
  rating_count: number;
  privacy: string;
  status: string;
  created_at: string;
  updated_at: string;
  media?: any[];
  collaborators?: any[];
}

export interface SearchPlacesInput {
  q?: string;
  type?: string;
  category?: string[];
  tags?: string[];
  city?: string;
  country?: string;
  lat?: number;
  lng?: number;
  radius?: number;
  limit?: number;
  offset?: number;
}

export interface NearbyPlacesInput {
  lat: number;
  lng: number;
  radius: number;
  type?: string;
  category?: string[];
  tags?: string[];
  limit?: number;
  offset?: number;
}

class PlacesService {
  async create(input: CreatePlaceInput): Promise<Place> {
    const response = await api.post<ApiResponse<Place>>('/places', input);
    return response.data;
  }

  async getById(id: string): Promise<Place> {
    const response = await api.get<ApiResponse<Place>>(`/places/${id}`);
    return response.data;
  }

  async update(id: string, input: UpdatePlaceInput): Promise<Place> {
    const response = await api.put<ApiResponse<Place>>(`/places/${id}`, input);
    return response.data;
  }

  async delete(id: string): Promise<void> {
    await api.delete(`/places/${id}`);
  }

  async getUserPlaces(params?: {
    page?: number;
    limit?: number;
  }): Promise<PaginatedResponse<Place>> {
    const queryParams = new URLSearchParams();
    if (params?.page) queryParams.append('page', params.page.toString());
    if (params?.limit) queryParams.append('limit', params.limit.toString());

    const query = queryParams.toString();
    return api.get<PaginatedResponse<Place>>(`/places/my${query ? `?${query}` : ''}`);
  }

  async getChildPlaces(parentId: string): Promise<Place[]> {
    const response = await api.get<ApiResponse<Place[]>>(`/places/${parentId}/children`);
    return response.data;
  }

  async search(input: SearchPlacesInput): Promise<PaginatedResponse<Place>> {
    const queryParams = new URLSearchParams();
    if (input.q) queryParams.append('q', input.q);
    if (input.type) queryParams.append('type', input.type);
    if (input.category?.length) {
      input.category.forEach(cat => queryParams.append('category', cat));
    }
    if (input.tags?.length) {
      input.tags.forEach(tag => queryParams.append('tags', tag));
    }
    if (input.city) queryParams.append('city', input.city);
    if (input.country) queryParams.append('country', input.country);
    if (input.lat !== undefined) queryParams.append('lat', input.lat.toString());
    if (input.lng !== undefined) queryParams.append('lng', input.lng.toString());
    if (input.radius !== undefined) queryParams.append('radius', input.radius.toString());
    if (input.limit !== undefined) queryParams.append('limit', input.limit.toString());
    if (input.offset !== undefined) queryParams.append('offset', input.offset.toString());

    return api.get<PaginatedResponse<Place>>(`/places/search?${queryParams.toString()}`);
  }

  async getNearby(input: NearbyPlacesInput): Promise<Place[]> {
    const queryParams = new URLSearchParams();
    queryParams.append('lat', input.lat.toString());
    queryParams.append('lng', input.lng.toString());
    queryParams.append('radius', input.radius.toString());
    if (input.type) queryParams.append('type', input.type);
    if (input.category?.length) {
      input.category.forEach(cat => queryParams.append('category', cat));
    }
    if (input.tags?.length) {
      input.tags.forEach(tag => queryParams.append('tags', tag));
    }
    if (input.limit !== undefined) queryParams.append('limit', input.limit.toString());
    if (input.offset !== undefined) queryParams.append('offset', input.offset.toString());

    const response = await api.get<ApiResponse<Place[]>>(`/places/nearby?${queryParams.toString()}`);
    return response.data;
  }

  // Collaborator management
  async addCollaborator(placeId: string, userId: string, role: string): Promise<void> {
    await api.post(`/places/${placeId}/collaborators`, { user_id: userId, role });
  }

  async removeCollaborator(placeId: string, userId: string): Promise<void> {
    await api.delete(`/places/${placeId}/collaborators/${userId}`);
  }

  async updateCollaboratorRole(placeId: string, userId: string, role: string): Promise<void> {
    await api.patch(`/places/${placeId}/collaborators/${userId}`, { role });
  }

  // Media management
  async uploadMedia(placeId: string, file: File, caption?: string): Promise<any> {
    const formData = new FormData();
    formData.append('file', file);
    if (caption) formData.append('caption', caption);

    const response = await api.upload<ApiResponse<any>>(`/places/${placeId}/media`, formData);
    return response.data;
  }

  async removeMedia(placeId: string, mediaId: string): Promise<void> {
    await api.delete(`/places/${placeId}/media/${mediaId}`);
  }

  async reorderMedia(placeId: string, mediaIds: string[]): Promise<void> {
    await api.put(`/places/${placeId}/media/reorder`, { media_ids: mediaIds });
  }
}

export const placesService = new PlacesService();