import React from 'react';
import { useAppSelector, useAppDispatch } from '../../hooks/redux';
import { clearRouteCreation, finishRouteCreation } from '../../store/slices/uiSlice';

export const RouteCreationOverlay: React.FC = () => {
  const dispatch = useAppDispatch();
  const routeMode = useAppSelector((state) => state.ui.routeCreationMode);

  if (!routeMode.isActive) return null;

  return (
    <div className="absolute top-20 left-1/2 transform -translate-x-1/2 z-40 bg-white rounded-lg shadow-xl p-4">
      <div className="flex items-center gap-4">
        <div className="flex items-center gap-2">
          <div className="w-3 h-3 bg-blue-500 rounded-full animate-pulse"></div>
          <span className="text-sm font-medium">
            Route Creation Mode - Click on map to add waypoints
          </span>
        </div>
        <div className="text-sm text-gray-500">
          {routeMode.waypoints.length} waypoint{routeMode.waypoints.length !== 1 ? 's' : ''}
        </div>
        <div className="flex gap-2">
          <button
            onClick={() => dispatch(finishRouteCreation())}
            disabled={routeMode.waypoints.length < 2}
            className="px-3 py-1 bg-blue-500 text-white rounded-md text-sm hover:bg-blue-600 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            Create Trip
          </button>
          <button
            onClick={() => dispatch(cancelRouteCreation())}
            className="px-3 py-1 border border-gray-300 rounded-md text-sm hover:bg-gray-50"
          >
            Cancel
          </button>
        </div>
      </div>
    </div>
  );
};