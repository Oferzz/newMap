import { api, ApiResponse } from './api';

export interface LoginInput {
  email: string;
  password: string;
}

export interface RegisterInput {
  email: string;
  username: string;
  password: string;
  display_name: string;
}

export interface User {
  id: string;
  email: string;
  username: string;
  display_name: string;
  avatar_url?: string;
  bio?: string;
  location?: string;
  created_at: string;
  updated_at: string;
}

export interface AuthResponse {
  user: User;
  access_token: string;
  refresh_token: string;
  expires_in: number;
}

export interface ChangePasswordInput {
  current_password: string;
  new_password: string;
}

export interface UpdateProfileInput {
  display_name?: string;
  bio?: string;
  location?: string;
  avatar_url?: string;
}

class AuthService {
  async login(input: LoginInput): Promise<AuthResponse> {
    const response = await api.post<ApiResponse<AuthResponse>>('/auth/login', input, {
      skipAuth: true,
    });
    return response.data;
  }

  async register(input: RegisterInput): Promise<User> {
    const response = await api.post<ApiResponse<User>>('/auth/register', input, {
      skipAuth: true,
    });
    return response.data;
  }

  async logout(): Promise<void> {
    await api.post('/auth/logout');
  }

  async refreshToken(refreshToken: string): Promise<AuthResponse> {
    const response = await api.post<ApiResponse<AuthResponse>>('/auth/refresh', {
      refresh_token: refreshToken,
    }, {
      skipAuth: true,
    });
    return response.data;
  }

  async getProfile(): Promise<User> {
    const response = await api.get<ApiResponse<User>>('/users/me');
    return response.data;
  }

  async updateProfile(input: UpdateProfileInput): Promise<User> {
    const response = await api.put<ApiResponse<User>>('/users/me', input);
    return response.data;
  }

  async changePassword(input: ChangePasswordInput): Promise<void> {
    await api.put('/users/me/password', input);
  }

  async resetPasswordRequest(email: string): Promise<void> {
    await api.post('/auth/reset-password', { email }, { skipAuth: true });
  }

  async resetPassword(token: string, newPassword: string): Promise<void> {
    await api.post('/auth/reset-password/confirm', {
      token,
      new_password: newPassword,
    }, {
      skipAuth: true,
    });
  }

  async verifyEmail(token: string): Promise<void> {
    await api.post('/auth/verify-email', { token }, { skipAuth: true });
  }

  async resendVerificationEmail(): Promise<void> {
    await api.post('/auth/resend-verification');
  }
}

export const authService = new AuthService();