import { api, ApiResponse, PaginatedResponse } from './api';
import { Collection } from '../types';

export interface CreateCollectionInput {
  name: string;
  description?: string;
  privacy?: 'public' | 'private' | 'friends';
}

export interface UpdateCollectionInput {
  name?: string;
  description?: string;
  privacy?: 'public' | 'private' | 'friends';
}

export interface AddLocationInput {
  name?: string;
  latitude: number;
  longitude: number;
}

export interface GetCollectionsParams {
  page?: number;
  limit?: number;
}

class CollectionsService {
  async createCollection(input: CreateCollectionInput): Promise<Collection> {
    const response = await api.post<ApiResponse<Collection>>('/collections', {
      ...input,
      privacy: input.privacy || 'private',
    });
    return response.data;
  }

  async getCollection(id: string): Promise<Collection> {
    const response = await api.get<ApiResponse<Collection>>(`/collections/${id}`);
    return response.data;
  }

  async getUserCollections(params?: GetCollectionsParams): Promise<PaginatedResponse<Collection>> {
    const queryParams = new URLSearchParams();
    if (params?.page) queryParams.append('page', params.page.toString());
    if (params?.limit) queryParams.append('limit', params.limit.toString());

    const query = queryParams.toString();
    return api.get<PaginatedResponse<Collection>>(`/collections${query ? `?${query}` : ''}`);
  }

  async updateCollection(id: string, input: UpdateCollectionInput): Promise<Collection> {
    const response = await api.put<ApiResponse<Collection>>(`/collections/${id}`, input);
    return response.data;
  }

  async deleteCollection(id: string): Promise<void> {
    await api.delete(`/collections/${id}`);
  }

  async addLocationToCollection(collectionId: string, input: AddLocationInput): Promise<any> {
    const response = await api.post<ApiResponse<any>>(`/collections/${collectionId}/locations`, input);
    return response.data;
  }

  async removeLocationFromCollection(collectionId: string, locationId: string): Promise<void> {
    await api.delete(`/collections/${collectionId}/locations/${locationId}`);
  }

  async addCollaborator(collectionId: string, userId: string, role: 'viewer' | 'editor'): Promise<void> {
    await api.post(`/collections/${collectionId}/collaborators`, {
      user_id: userId,
      role,
    });
  }

  async removeCollaborator(collectionId: string, userId: string): Promise<void> {
    await api.delete(`/collections/${collectionId}/collaborators/${userId}`);
  }
}

export const collectionsService = new CollectionsService();