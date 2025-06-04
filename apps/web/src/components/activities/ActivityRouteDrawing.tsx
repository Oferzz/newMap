import React, { useState, useEffect } from 'react';
import { useAppDispatch, useAppSelector } from '../../hooks/redux';
import { 
  startRouteCreation, 
  removeRouteWaypoint,
  clearRouteCreation,
  finishRouteCreation,
  undoLastWaypoint
} from '../../store/slices/uiSlice';
import { MapPin, Trash2, Undo, RotateCcw, Route as RouteIcon } from 'lucide-react';

interface RouteData {
  type: 'out-and-back' | 'loop' | 'point-to-point';
  waypoints: Array<{ lat: number; lng: number; elevation?: number }>;
  distance?: number;
  elevationGain?: number;
  elevationLoss?: number;
}

interface ActivityRouteDrawingProps {
  onRouteUpdate?: (routeData: RouteData) => void;
}

export const ActivityRouteDrawing: React.FC<ActivityRouteDrawingProps> = ({ 
  onRouteUpdate 
}) => {
  const dispatch = useAppDispatch();
  const routeCreationMode = useAppSelector((state) => state.ui.routeCreationMode);
  const [routeType, setRouteType] = useState<RouteData['type']>('point-to-point');
  const [showAdvanced, setShowAdvanced] = useState(false);

  useEffect(() => {
    // Start route creation mode when component mounts
    if (!routeCreationMode.isActive) {
      dispatch(startRouteCreation({}));
    }
  }, [dispatch, routeCreationMode.isActive]);

  useEffect(() => {
    // Update parent component when route data changes
    if (onRouteUpdate && routeCreationMode.waypoints.length > 0) {
      const routeData: RouteData = {
        type: routeType,
        waypoints: routeCreationMode.waypoints.map(wp => ({
          lat: wp.coordinates[1],
          lng: wp.coordinates[0],
          elevation: wp.elevation,
        })),
        distance: routeCreationMode.distance,
        elevationGain: routeCreationMode.elevationGain,
        elevationLoss: routeCreationMode.elevationLoss,
      };
      onRouteUpdate(routeData);
    }
  }, [routeCreationMode, routeType, onRouteUpdate]);

  const handleStartRoute = () => {
    dispatch(startRouteCreation({}));
  };

  const handleClearRoute = () => {
    dispatch(clearRouteCreation());
  };

  const handleUndoLastPoint = () => {
    dispatch(undoLastWaypoint());
  };

  const handleFinishRoute = () => {
    if (routeCreationMode.waypoints.length >= 2) {
      dispatch(finishRouteCreation());
    }
  };

  const handleRemoveWaypoint = (index: number) => {
    dispatch(removeRouteWaypoint({ index }));
  };

  const calculateRouteStats = () => {
    const { waypoints } = routeCreationMode;
    if (waypoints.length < 2) return null;

    // Calculate total distance (simplified - in real implementation would use proper geodesic calculations)
    let totalDistance = 0;
    let elevationGain = 0;
    let elevationLoss = 0;

    for (let i = 1; i < waypoints.length; i++) {
      const prev = waypoints[i - 1];
      const curr = waypoints[i];
      
      // Simple distance calculation (should use proper geodesic formula)
      const latDiff = curr.coordinates[1] - prev.coordinates[1];
      const lngDiff = curr.coordinates[0] - prev.coordinates[0];
      const segmentDistance = Math.sqrt(latDiff * latDiff + lngDiff * lngDiff) * 111.32; // Rough conversion to km
      totalDistance += segmentDistance;

      // Elevation calculations
      if (prev.elevation && curr.elevation) {
        const elevDiff = curr.elevation - prev.elevation;
        if (elevDiff > 0) {
          elevationGain += elevDiff;
        } else {
          elevationLoss += Math.abs(elevDiff);
        }
      }
    }

    return {
      distance: totalDistance,
      elevationGain,
      elevationLoss,
    };
  };

  const routeStats = calculateRouteStats();

  return (
    <div className="space-y-4">
      {/* Route Type Selection */}
      <div>
        <label className="block text-sm font-medium text-trail-700 mb-3">
          Route Type
        </label>
        <div className="grid grid-cols-3 gap-2">
          {[
            { value: 'point-to-point', label: 'Point to Point' },
            { value: 'out-and-back', label: 'Out & Back' },
            { value: 'loop', label: 'Loop' },
          ].map((option) => (
            <button
              key={option.value}
              onClick={() => setRouteType(option.value as RouteData['type'])}
              className={`px-3 py-2 text-sm rounded-lg border transition-colors ${
                routeType === option.value
                  ? 'bg-forest-100 border-forest-300 text-forest-800'
                  : 'bg-white border-terrain-300 text-trail-700 hover:bg-terrain-50'
              }`}
            >
              {option.label}
            </button>
          ))}
        </div>
      </div>

      {/* Route Controls */}
      <div className="flex flex-wrap gap-2">
        {!routeCreationMode.isActive ? (
          <button
            onClick={handleStartRoute}
            className="flex items-center space-x-2 px-3 py-2 bg-forest-600 text-white rounded-lg text-sm font-medium hover:bg-forest-700 transition-colors"
          >
            <RouteIcon className="w-4 h-4" />
            <span>Start Drawing</span>
          </button>
        ) : (
          <>
            <button
              onClick={handleUndoLastPoint}
              disabled={routeCreationMode.waypoints.length === 0}
              className="flex items-center space-x-1 px-3 py-2 bg-terrain-100 text-trail-700 rounded-lg text-sm hover:bg-terrain-200 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
            >
              <Undo className="w-4 h-4" />
              <span>Undo</span>
            </button>
            
            <button
              onClick={handleClearRoute}
              className="flex items-center space-x-1 px-3 py-2 bg-red-100 text-red-700 rounded-lg text-sm hover:bg-red-200 transition-colors"
            >
              <RotateCcw className="w-4 h-4" />
              <span>Clear</span>
            </button>

            {routeCreationMode.waypoints.length >= 2 && (
              <button
                onClick={handleFinishRoute}
                className="flex items-center space-x-1 px-3 py-2 bg-green-600 text-white rounded-lg text-sm font-medium hover:bg-green-700 transition-colors"
              >
                <span>Finish Route</span>
              </button>
            )}
          </>
        )}
      </div>

      {/* Instructions */}
      <div className="bg-blue-50 border border-blue-200 rounded-lg p-3">
        <div className="text-sm text-blue-800">
          {!routeCreationMode.isActive ? (
            <p>Click "Start Drawing" and then click on the map to create waypoints for your route.</p>
          ) : routeCreationMode.waypoints.length === 0 ? (
            <p>Click on the map to add your first waypoint.</p>
          ) : routeCreationMode.waypoints.length === 1 ? (
            <p>Add at least one more waypoint to create a route.</p>
          ) : (
            <p>Continue adding waypoints or click "Finish Route" when done.</p>
          )}
        </div>
      </div>

      {/* Waypoints List */}
      {routeCreationMode.waypoints.length > 0 && (
        <div>
          <div className="flex items-center justify-between mb-3">
            <h4 className="text-sm font-medium text-trail-700">
              Waypoints ({routeCreationMode.waypoints.length})
            </h4>
            <button
              onClick={() => setShowAdvanced(!showAdvanced)}
              className="text-xs text-forest-600 hover:text-forest-800"
            >
              {showAdvanced ? 'Hide Details' : 'Show Details'}
            </button>
          </div>
          
          <div className="space-y-2 max-h-40 overflow-y-auto">
            {routeCreationMode.waypoints.map((waypoint, index) => (
              <div
                key={index}
                className="flex items-center justify-between p-2 bg-terrain-50 rounded border border-terrain-200"
              >
                <div className="flex items-center space-x-2">
                  <MapPin className="w-4 h-4 text-forest-600" />
                  <div className="text-sm">
                    <div className="text-trail-800 font-medium">
                      Point {index + 1}
                    </div>
                    {showAdvanced && (
                      <div className="text-xs text-trail-600">
                        {waypoint.coordinates[1].toFixed(4)}, {waypoint.coordinates[0].toFixed(4)}
                        {waypoint.elevation && ` • ${waypoint.elevation.toFixed(0)}m`}
                      </div>
                    )}
                  </div>
                </div>
                
                <button
                  onClick={() => handleRemoveWaypoint(index)}
                  className="p-1 text-red-600 hover:text-red-800 hover:bg-red-100 rounded transition-colors"
                  title="Remove waypoint"
                >
                  <Trash2 className="w-3 h-3" />
                </button>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Route Statistics */}
      {routeStats && (
        <div className="bg-forest-50 border border-forest-200 rounded-lg p-4">
          <h4 className="text-sm font-medium text-forest-800 mb-3">Route Statistics</h4>
          <div className="grid grid-cols-2 gap-4 text-sm">
            <div>
              <span className="text-forest-600">Distance:</span>
              <div className="font-medium text-forest-800">
                {routeStats.distance.toFixed(2)} km
              </div>
            </div>
            
            <div>
              <span className="text-forest-600">Waypoints:</span>
              <div className="font-medium text-forest-800">
                {routeCreationMode.waypoints.length}
              </div>
            </div>
            
            {routeStats.elevationGain > 0 && (
              <div>
                <span className="text-forest-600">Elevation Gain:</span>
                <div className="font-medium text-forest-800">
                  {routeStats.elevationGain.toFixed(0)} m
                </div>
              </div>
            )}
            
            {routeStats.elevationLoss > 0 && (
              <div>
                <span className="text-forest-600">Elevation Loss:</span>
                <div className="font-medium text-forest-800">
                  {routeStats.elevationLoss.toFixed(0)} m
                </div>
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  );
};