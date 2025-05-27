import { createAsyncThunk } from '@reduxjs/toolkit';
import { authService, LoginInput, RegisterInput, UpdateProfileInput } from '../../services/auth.service';
import { loginSuccess, loginFailure, updateUser } from '../slices/authSlice';
import toast from 'react-hot-toast';

export const loginThunk = createAsyncThunk(
  'auth/login',
  async (input: LoginInput, { dispatch }) => {
    try {
      const response = await authService.login(input);
      dispatch(loginSuccess({
        user: response.user,
        accessToken: response.access_token,
        refreshToken: response.refresh_token,
      }));
      toast.success('Welcome back!');
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
  async (input: RegisterInput) => {
    try {
      const user = await authService.register(input);
      toast.success('Account created successfully! Please login.');
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
        user: response.user,
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
      dispatch(updateUser(user));
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
      dispatch(updateUser(user));
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