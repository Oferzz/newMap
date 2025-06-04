import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAppDispatch, useAppSelector } from '../hooks/redux';
import { MapPin, Route, Settings, Share, Check, ChevronLeft, ChevronRight } from 'lucide-react';
import { MapView } from '../components/map/MapView';
import { ActivityRouteDrawing } from '../components/activities/ActivityRouteDrawing';
import { ActivityMetadataForm } from '../components/activities/ActivityMetadataForm';
import { ActivityVisibilitySettings } from '../components/activities/ActivityVisibilitySettings';
import { createActivityThunk } from '../store/thunks/activities.thunks';
import { clearRouteCreation, setActivePanel } from '../store/slices/uiSlice';

export interface ActivityFormData {
  title: string;
  description: string;
  activityType: string;
  route?: {
    type: 'out-and-back' | 'loop' | 'point-to-point';
    waypoints: Array<{ lat: number; lng: number; elevation?: number }>;
    distance?: number;
    elevationGain?: number;
    elevationLoss?: number;
  };
  metadata: {
    difficulty: 'easy' | 'moderate' | 'hard' | 'expert';
    duration: number; // in hours
    distance: number; // in kilometers
    elevationGain: number; // in meters
    terrain: string[];
    waterFeatures: string[];
    gear: string[];
    seasons: string[];
    conditions: string[];
    tags: string[];
  };
  visibility: {
    privacy: 'public' | 'friends' | 'private';
    allowComments: boolean;
    allowDownloads: boolean;
    shareWithGroups: string[];
  };
}

const STEPS = [
  { id: 'basic', name: 'Basic Info', icon: MapPin },
  { id: 'route', name: 'Route', icon: Route },
  { id: 'details', name: 'Details', icon: Settings },
  { id: 'sharing', name: 'Sharing', icon: Share },
] as const;

export const ActivityCreationPage: React.FC = () => {
  const navigate = useNavigate();
  const dispatch = useAppDispatch();
  const [currentStep, setCurrentStep] = useState(0);
  const [isCreating, setIsCreating] = useState(false);
  
  const [formData, setFormData] = useState<ActivityFormData>({
    title: '',
    description: '',
    activityType: 'hiking',
    metadata: {
      difficulty: 'moderate',
      duration: 4,
      distance: 10,
      elevationGain: 500,
      terrain: [],
      waterFeatures: [],
      gear: [],
      seasons: [],
      conditions: [],
      tags: [],
    },
    visibility: {
      privacy: 'public',
      allowComments: true,
      allowDownloads: true,
      shareWithGroups: [],
    },
  });

  const routeCreationMode = useAppSelector((state) => state.ui.routeCreationMode);

  useEffect(() => {
    // Clear any existing route creation state when component mounts
    dispatch(clearRouteCreation());
    dispatch(setActivePanel('none'));
    
    return () => {
      // Clean up on unmount
      dispatch(clearRouteCreation());
    };
  }, [dispatch]);

  const updateFormData = (section: keyof ActivityFormData, data: any) => {
    setFormData(prev => ({
      ...prev,
      [section]: typeof data === 'object' && !Array.isArray(data) && prev[section] && typeof prev[section] === 'object'
        ? { ...(prev[section] as object), ...data }
        : data
    }));
  };

  const handleNext = () => {
    if (currentStep < STEPS.length - 1) {
      setCurrentStep(prev => prev + 1);
    }
  };

  const handlePrevious = () => {
    if (currentStep > 0) {
      setCurrentStep(prev => prev - 1);
    }
  };

  const validateCurrentStep = (): boolean => {
    switch (currentStep) {
      case 0: // Basic Info
        return formData.title.trim().length >= 3 && formData.activityType !== '';
      case 1: // Route
        return routeCreationMode.waypoints.length >= 2;
      case 2: // Details
        return formData.metadata.duration > 0 && formData.metadata.distance > 0;
      case 3: // Sharing
        return true; // Always valid
      default:
        return false;
    }
  };

  const handleCreate = async () => {
    if (!validateCurrentStep()) return;

    setIsCreating(true);
    try {
      // Combine route creation data with form data
      const activityData = {
        ...formData,
        route: {
          ...formData.route,
          waypoints: routeCreationMode.waypoints.map(wp => ({
            lat: wp.coordinates[1],
            lng: wp.coordinates[0],
            elevation: wp.elevation,
          })),
          type: formData.route?.type || 'point-to-point',
        }
      };

      await dispatch(createActivityThunk(activityData)).unwrap();
      
      // Clear route creation state
      dispatch(clearRouteCreation());
      
      // Navigate to activity or activities list
      navigate('/activities');
    } catch (error) {
      console.error('Failed to create activity:', error);
      // Error handling is managed by the thunk and notifications
    } finally {
      setIsCreating(false);
    }
  };

  const renderStepContent = () => {
    switch (currentStep) {
      case 0:
        return (
          <div className="space-y-6">
            <div>
              <label className="block text-sm font-medium text-trail-700 mb-2">
                Activity Title *
              </label>
              <input
                type="text"
                value={formData.title}
                onChange={(e) => updateFormData('title', e.target.value)}
                placeholder="e.g., Mount Tamalpais Loop Trail"
                className="w-full px-4 py-3 border border-terrain-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-forest-500 focus:border-transparent"
                required
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-trail-700 mb-2">
                Description
              </label>
              <textarea
                value={formData.description}
                onChange={(e) => updateFormData('description', e.target.value)}
                placeholder="Describe your activity, highlights, and any important details..."
                rows={4}
                className="w-full px-4 py-3 border border-terrain-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-forest-500 focus:border-transparent resize-none"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-trail-700 mb-2">
                Activity Type *
              </label>
              <select
                value={formData.activityType}
                onChange={(e) => updateFormData('activityType', e.target.value)}
                className="w-full px-4 py-3 border border-terrain-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-forest-500 focus:border-transparent"
                required
              >
                <option value="hiking">Hiking</option>
                <option value="trail-running">Trail Running</option>
                <option value="biking">Mountain Biking</option>
                <option value="road-cycling">Road Cycling</option>
                <option value="backpacking">Backpacking</option>
                <option value="climbing">Rock Climbing</option>
                <option value="skiing">Skiing</option>
                <option value="snowboarding">Snowboarding</option>
                <option value="kayaking">Kayaking</option>
                <option value="canoeing">Canoeing</option>
                <option value="fishing">Fishing</option>
                <option value="camping">Camping</option>
                <option value="photography">Photography</option>
                <option value="wildlife-viewing">Wildlife Viewing</option>
                <option value="other">Other</option>
              </select>
            </div>
          </div>
        );

      case 1:
        return (
          <div className="space-y-4">
            <div className="text-center">
              <h3 className="text-lg font-semibold text-trail-800 mb-2">
                Draw Your Route
              </h3>
              <p className="text-trail-600 text-sm mb-4">
                Click on the map to add waypoints and create your route. 
                Use the controls to customize your route type.
              </p>
            </div>
            <ActivityRouteDrawing
              onRouteUpdate={(routeData) => updateFormData('route', routeData)}
            />
          </div>
        );

      case 2:
        return (
          <ActivityMetadataForm
            metadata={formData.metadata}
            activityType={formData.activityType}
            onUpdate={(metadata) => updateFormData('metadata', metadata)}
          />
        );

      case 3:
        return (
          <ActivityVisibilitySettings
            settings={formData.visibility}
            onUpdate={(visibility) => updateFormData('visibility', visibility)}
          />
        );

      default:
        return null;
    }
  };

  return (
    <div className="min-h-screen bg-terrain-50">
      {/* Header */}
      <div className="bg-white border-b border-terrain-200">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex items-center justify-between h-16">
            <div className="flex items-center">
              <button
                onClick={() => navigate(-1)}
                className="mr-4 p-2 text-trail-600 hover:text-trail-800 hover:bg-terrain-100 rounded-lg transition-colors"
              >
                <ChevronLeft className="w-5 h-5" />
              </button>
              <h1 className="text-xl font-semibold text-trail-800">
                Create New Activity
              </h1>
            </div>
            
            {/* Progress Steps */}
            <div className="hidden md:flex items-center space-x-4">
              {STEPS.map((step, index) => {
                const Icon = step.icon;
                const isActive = index === currentStep;
                const isCompleted = index < currentStep;
                
                return (
                  <div
                    key={step.id}
                    className={`flex items-center ${
                      index < STEPS.length - 1 ? 'space-x-4' : ''
                    }`}
                  >
                    <div className="flex items-center space-x-2">
                      <div
                        className={`w-8 h-8 rounded-full flex items-center justify-center transition-colors ${
                          isCompleted
                            ? 'bg-forest-600 text-white'
                            : isActive
                            ? 'bg-forest-100 text-forest-600 border-2 border-forest-600'
                            : 'bg-terrain-200 text-trail-400'
                        }`}
                      >
                        {isCompleted ? (
                          <Check className="w-4 h-4" />
                        ) : (
                          <Icon className="w-4 h-4" />
                        )}
                      </div>
                      <span
                        className={`text-sm font-medium ${
                          isActive || isCompleted
                            ? 'text-trail-800'
                            : 'text-trail-400'
                        }`}
                      >
                        {step.name}
                      </span>
                    </div>
                    {index < STEPS.length - 1 && (
                      <div
                        className={`w-8 h-0.5 ${
                          isCompleted ? 'bg-forest-600' : 'bg-terrain-200'
                        }`}
                      />
                    )}
                  </div>
                );
              })}
            </div>
          </div>
        </div>
      </div>

      {/* Main Content */}
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
          {/* Form Panel */}
          <div className="bg-white rounded-lg shadow-sm border border-terrain-200 p-6">
            <div className="mb-6">
              <h2 className="text-lg font-semibold text-trail-800 mb-1">
                {STEPS[currentStep].name}
              </h2>
              <div className="text-sm text-trail-600">
                Step {currentStep + 1} of {STEPS.length}
              </div>
            </div>

            {renderStepContent()}

            {/* Navigation Buttons */}
            <div className="flex justify-between items-center mt-8 pt-6 border-t border-terrain-200">
              <button
                onClick={handlePrevious}
                disabled={currentStep === 0}
                className="flex items-center space-x-2 px-4 py-2 text-trail-600 hover:text-trail-800 hover:bg-terrain-100 rounded-lg transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
              >
                <ChevronLeft className="w-4 h-4" />
                <span>Previous</span>
              </button>

              {currentStep === STEPS.length - 1 ? (
                <button
                  onClick={handleCreate}
                  disabled={!validateCurrentStep() || isCreating}
                  className="flex items-center space-x-2 px-6 py-3 bg-forest-600 text-white rounded-lg font-medium hover:bg-forest-700 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  {isCreating ? 'Creating...' : 'Create Activity'}
                </button>
              ) : (
                <button
                  onClick={handleNext}
                  disabled={!validateCurrentStep()}
                  className="flex items-center space-x-2 px-4 py-2 bg-forest-600 text-white rounded-lg font-medium hover:bg-forest-700 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  <span>Next</span>
                  <ChevronRight className="w-4 h-4" />
                </button>
              )}
            </div>
          </div>

          {/* Map Panel */}
          <div className="bg-white rounded-lg shadow-sm border border-terrain-200 overflow-hidden">
            <div className="h-96 lg:h-full relative">
              <MapView />
              {currentStep === 1 && (
                <div className="absolute top-4 left-4 right-4 z-10">
                  <div className="bg-white rounded-lg shadow-lg p-4">
                    <div className="text-sm text-trail-600 mb-2">
                      Route Statistics
                    </div>
                    <div className="grid grid-cols-2 gap-4 text-sm">
                      <div>
                        <span className="text-trail-500">Waypoints:</span>
                        <span className="ml-2 font-medium text-trail-800">
                          {routeCreationMode.waypoints.length}
                        </span>
                      </div>
                      <div>
                        <span className="text-trail-500">Distance:</span>
                        <span className="ml-2 font-medium text-trail-800">
                          {routeCreationMode.distance ? `${routeCreationMode.distance.toFixed(1)} km` : '--'}
                        </span>
                      </div>
                    </div>
                  </div>
                </div>
              )}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};