import { store } from '../store';
import { logout } from '../store/slices/authSlice';

const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080/api/v1';

interface ApiOptions extends RequestInit {
  skipAuth?: boolean;
}

class ApiService {
  private async request<T>(endpoint: string, options: ApiOptions = {}): Promise<T> {
    const { skipAuth = false, ...fetchOptions } = options;
    
    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
      ...options.headers as Record<string, string>,
    };

    // Add auth token if available and not skipped
    if (!skipAuth) {
      const state = store.getState();
      const token = state.auth.accessToken;
      if (token) {
        headers['Authorization'] = `Bearer ${token}`;
      }
    }

    const url = `${API_BASE_URL}${endpoint}`;
    
    try {
      const response = await fetch(url, {
        ...fetchOptions,
        headers,
      });

      // Handle 401 Unauthorized
      if (response.status === 401 && !skipAuth) {
        store.dispatch(logout());
        window.location.href = '/login';
        throw new Error('Session expired. Please login again.');
      }

      const data = await response.json();

      if (!response.ok) {
        throw new Error(data.error || `API Error: ${response.status}`);
      }

      return data;
    } catch (error) {
      if (error instanceof Error) {
        throw error;
      }
      throw new Error('Network error occurred');
    }
  }

  // GET request
  async get<T>(endpoint: string, options?: ApiOptions): Promise<T> {
    return this.request<T>(endpoint, {
      ...options,
      method: 'GET',
    });
  }

  // POST request
  async post<T>(endpoint: string, body?: any, options?: ApiOptions): Promise<T> {
    return this.request<T>(endpoint, {
      ...options,
      method: 'POST',
      body: body ? JSON.stringify(body) : undefined,
    });
  }

  // PUT request
  async put<T>(endpoint: string, body?: any, options?: ApiOptions): Promise<T> {
    return this.request<T>(endpoint, {
      ...options,
      method: 'PUT',
      body: body ? JSON.stringify(body) : undefined,
    });
  }

  // PATCH request
  async patch<T>(endpoint: string, body?: any, options?: ApiOptions): Promise<T> {
    return this.request<T>(endpoint, {
      ...options,
      method: 'PATCH',
      body: body ? JSON.stringify(body) : undefined,
    });
  }

  // DELETE request
  async delete<T>(endpoint: string, options?: ApiOptions): Promise<T> {
    return this.request<T>(endpoint, {
      ...options,
      method: 'DELETE',
    });
  }

  // File upload
  async upload<T>(endpoint: string, formData: FormData, options?: ApiOptions): Promise<T> {
    const { headers = {}, ...restOptions } = options || {};
    
    // Don't set Content-Type for FormData, let browser set it
    const { 'Content-Type': _, ...restHeaders } = headers as Record<string, string>;
    
    return this.request<T>(endpoint, {
      ...restOptions,
      method: 'POST',
      headers: restHeaders,
      body: formData,
    });
  }
}

// Export singleton instance
export const api = new ApiService();

// Export types for API responses
export interface ApiResponse<T> {
  success: boolean;
  data: T;
  error?: string;
  meta?: {
    page?: number;
    limit?: number;
    total?: number;
    hasMore?: boolean;
  };
}

export interface PaginatedResponse<T> extends ApiResponse<T[]> {
  meta: {
    page: number;
    limit: number;
    total: number;
    hasMore: boolean;
  };
}