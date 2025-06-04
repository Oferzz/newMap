import { createAsyncThunk } from '@reduxjs/toolkit';
import { authService, LoginInput, RegisterInput, UpdateProfileInput } from '../../services/auth.service';
import { loginSuccess, loginFailure, updateUser } from '../slices/authSlice';
import { dataMigrationService } from '../../services/storage/dataMigration.service';
import toast from 'react-hot-toast';

export const loginThunk = createAsyncThunk(
  'auth/login',
  async (input: LoginInput, { dispatch }) => {
    try {
      const response = await authService.login(input);
      dispatch(loginSuccess({
        user: {
          id: response.user.id,
          email: response.user.email,
          username: response.user.username,
          displayName: response.user.display_name,
          avatarUrl: response.user.avatar_url,
          bio: response.user.bio,
          location: response.user.location,
          created_at: response.user.created_at,
          updated_at: response.user.updated_at,
        },
        accessToken: response.access_token,
        refreshToken: response.refresh_token,
      }));
      toast.success('Welcome back!');
      
      // Check for local data and offer to migrate
      const shouldMigrate = await dataMigrationService.promptForMigration();
      if (shouldMigrate) {
        await dataMigrationService.migrateLocalDataToCloud();
      }
      
      return response;
    } catch (error) {
      const message = error instanceof Error ? error.message : 'Login failed';
      dispatch(loginFailure(message));
      toast.error(message);
      throw error;
    }
  }
);

export const registerThunk = createAsyncThunk(
  'auth/register',
  async (input: RegisterInput, { dispatch }) => {
    try {
      const user = await authService.register(input);
      toast.success('Account created successfully!');
      
      // Auto-login after registration
      const loginResponse = await authService.login({
        email: input.email,
        password: input.password,
      });
      
      dispatch(loginSuccess({
        user: {
          id: loginResponse.user.id,
          email: loginResponse.user.email,
          username: loginResponse.user.username,
          displayName: loginResponse.user.display_name,
          avatarUrl: loginResponse.user.avatar_url,
        },
        accessToken: loginResponse.access_token,
        refreshToken: loginResponse.refresh_token,
      }));
      
      // Check for local data and offer to migrate
      const shouldMigrate = await dataMigrationService.promptForMigration();
      if (shouldMigrate) {
        await dataMigrationService.migrateLocalDataToCloud();
      }
      
      return user;
    } catch (error) {
      const message = error instanceof Error ? error.message : 'Registration failed';
      toast.error(message);
      throw error;
    }
  }
);

export const refreshTokenThunk = createAsyncThunk(
  'auth/refreshToken',
  async (refreshToken: string, { dispatch }) => {
    try {
      const response = await authService.refreshToken(refreshToken);
      dispatch(loginSuccess({
        user: {
          id: response.user.id,
          email: response.user.email,
          username: response.user.username,
          displayName: response.user.display_name,
          avatarUrl: response.user.avatar_url,
          bio: response.user.bio,
          location: response.user.location,
          created_at: response.user.created_at,
          updated_at: response.user.updated_at,
        },
        accessToken: response.access_token,
        refreshToken: response.refresh_token,
      }));
      return response;
    } catch (error) {
      throw error;
    }
  }
);

export const getProfileThunk = createAsyncThunk(
  'auth/getProfile',
  async (_, { dispatch }) => {
    try {
      const user = await authService.getProfile();
      dispatch(updateUser({
        id: user.id,
        email: user.email,
        username: user.username,
        displayName: user.display_name,
        avatarUrl: user.avatar_url,
        bio: user.bio,
        location: user.location,
        created_at: user.created_at,
        updated_at: user.updated_at,
      }));
      return user;
    } catch (error) {
      throw error;
    }
  }
);

export const updateProfileThunk = createAsyncThunk(
  'auth/updateProfile',
  async (input: UpdateProfileInput, { dispatch }) => {
    try {
      const user = await authService.updateProfile(input);
      dispatch(updateUser({
        id: user.id,
        email: user.email,
        username: user.username,
        displayName: user.display_name,
        avatarUrl: user.avatar_url,
        bio: user.bio,
        location: user.location,
        created_at: user.created_at,
        updated_at: user.updated_at,
      }));
      toast.success('Profile updated successfully');
      return user;
    } catch (error) {
      const message = error instanceof Error ? error.message : 'Failed to update profile';
      toast.error(message);
      throw error;
    }
  }
);

export const changePasswordThunk = createAsyncThunk(
  'auth/changePassword',
  async (input: { current_password: string; new_password: string }) => {
    try {
      await authService.changePassword(input);
      toast.success('Password changed successfully');
    } catch (error) {
      const message = error instanceof Error ? error.message : 'Failed to change password';
      toast.error(message);
      throw error;
    }
  }
);

export const resetPasswordRequestThunk = createAsyncThunk(
  'auth/resetPasswordRequest',
  async (email: string) => {
    try {
      await authService.resetPasswordRequest(email);
      toast.success('Password reset email sent. Please check your inbox.');
    } catch (error) {
      const message = error instanceof Error ? error.message : 'Failed to send reset email';
      toast.error(message);
      throw error;
    }
  }
);

export const verifyEmailThunk = createAsyncThunk(
  'auth/verifyEmail',
  async (token: string) => {
    try {
      await authService.verifyEmail(token);
      toast.success('Email verified successfully!');
    } catch (error) {
      const message = error instanceof Error ? error.message : 'Failed to verify email';
      toast.error(message);
      throw error;
    }
  }
);

export const initializeAuthThunk = createAsyncThunk(
  'auth/initialize',
  async (_, { dispatch }) => {
    const accessToken = localStorage.getItem('accessToken');
    const refreshToken = localStorage.getItem('refreshToken');
    
    if (!accessToken || !refreshToken) {
      return null;
    }
    
    try {
      // Try to get profile with current token
      const user = await authService.getProfile();
      dispatch(loginSuccess({
        user: {
          id: user.id,
          email: user.email,
          username: user.username,
          displayName: user.display_name,
          avatarUrl: user.avatar_url,
          bio: user.bio,
          location: user.location,
          created_at: user.created_at,
          updated_at: user.updated_at,
        },
        accessToken,
        refreshToken,
      }));
      return user;
    } catch (error) {
      // If token is expired, try to refresh
      if (refreshToken) {
        try {
          const response = await authService.refreshToken(refreshToken);
          dispatch(loginSuccess({
            user: {
              id: response.user.id,
              email: response.user.email,
              username: response.user.username,
              displayName: response.user.display_name,
              avatarUrl: response.user.avatar_url,
            },
            accessToken: response.access_token,
            refreshToken: response.refresh_token,
          }));
          return response.user;
        } catch (refreshError) {
          // Clear invalid tokens
          localStorage.removeItem('accessToken');
          localStorage.removeItem('refreshToken');
          return null;
        }
      }
      return null;
    }
  }
);