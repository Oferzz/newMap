import React from 'react';
import { useAppSelector } from '../../hooks/redux';

export const StorageModeIndicator: React.FC = () => {
  const isAuthenticated = useAppSelector((state) => state.auth.isAuthenticated);
  const isLocalStorage = useAppSelector((state) => state.collections.isLocalStorage);
  
  // Don't show indicator if authenticated (cloud storage is default)
  if (isAuthenticated && !isLocalStorage) {
    return null;
  }
  
  return (
    <div className="fixed bottom-4 left-4 z-40">
      <div className={`
        px-4 py-2 rounded-full text-sm font-medium shadow-lg
        ${isAuthenticated 
          ? 'bg-green-100 text-green-800 border border-green-200' 
          : 'bg-yellow-100 text-yellow-800 border border-yellow-200'
        }
      `}>
        <div className="flex items-center gap-2">
          <span className="text-lg">
            {isAuthenticated ? '☁️' : '💾'}
          </span>
          <span>
            {isAuthenticated 
              ? 'Cloud sync active' 
              : 'Saved locally (sign in to sync)'
            }
          </span>
        </div>
      </div>
    </div>
  );
};