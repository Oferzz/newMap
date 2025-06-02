import React from 'react';
import { useAppSelector } from '../../hooks/redux';

export const StorageModeIndicator: React.FC = () => {
  const isAuthenticated = useAppSelector((state) => state.auth.isAuthenticated);
  const isLocalStorage = useAppSelector((state) => state.collections.isLocalStorage);
  
  // Don't show any indicator - removed per user request
  return null;
};