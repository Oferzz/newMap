import React, { useEffect, useState } from 'react';
import { useAppDispatch, useAppSelector } from '../../hooks/redux';
import { useNavigate } from 'react-router-dom';
import { 
  Plus, 
  Calendar, 
  MapPin, 
  Users, 
  Search,
  ArrowRight,
  Clock,
  Globe,
  Lock,
  Trash2,
  Edit
} from 'lucide-react';
import { getUserTripsThunk, deleteTripThunk } from '../../store/thunks/trips.thunks';
import { setActivePanel } from '../../store/slices/uiSlice';
import { Trip } from '../../store/slices/tripsSlice';
import { format, differenceInDays } from 'date-fns';

interface TripsPanelProps {
  isOpen: boolean;
  onClose: () => void;
}

export const TripsPanel: React.FC<TripsPanelProps> = ({ isOpen, onClose }) => {
  const dispatch = useAppDispatch();
  const navigate = useNavigate();
  const { items: trips, isLoading, error } = useAppSelector(state => state.trips);
  const { isAuthenticated } = useAppSelector(state => state.auth);
  
  const [naturalLanguageQuery, setNaturalLanguageQuery] = useState('');

  useEffect(() => {
    if (isOpen) {
      // Load trips when panel opens
      dispatch(getUserTripsThunk());
    }
  }, [isOpen, dispatch]);

  // Parse natural language query for filtering
  const parseNaturalQuery = (query: string) => {
    const lowerQuery = query.toLowerCase();
    let status: string | null = null;
    let sortBy = 'recent';
    let searchTerms = query;

    // Extract status filters
    if (lowerQuery.includes('planning') || lowerQuery.includes('planned')) {
      status = 'planning';
      searchTerms = searchTerms.replace(/\b(planning|planned)\b/gi, '');
    } else if (lowerQuery.includes('active') || lowerQuery.includes('ongoing')) {
      status = 'active';
      searchTerms = searchTerms.replace(/\b(active|ongoing)\b/gi, '');
    } else if (lowerQuery.includes('completed') || lowerQuery.includes('finished')) {
      status = 'completed';
      searchTerms = searchTerms.replace(/\b(completed|finished)\b/gi, '');
    }

    // Extract sort preferences
    if (lowerQuery.includes('alphabetical') || lowerQuery.includes('a-z') || lowerQuery.includes('name')) {
      sortBy = 'alphabetical';
      searchTerms = searchTerms.replace(/\b(alphabetical|a-z|by name)\b/gi, '');
    } else if (lowerQuery.includes('date') || lowerQuery.includes('when')) {
      sortBy = 'date';
      searchTerms = searchTerms.replace(/\b(by date|date|when)\b/gi, '');
    }

    return { status, sortBy, searchTerms: searchTerms.trim() };
  };

  const { status: filterStatus, sortBy, searchTerms } = parseNaturalQuery(naturalLanguageQuery);

  const filteredTrips = trips
    .filter(trip => {
      const matchesSearch = !searchTerms || 
                           trip.title.toLowerCase().includes(searchTerms.toLowerCase()) ||
                           trip.description.toLowerCase().includes(searchTerms.toLowerCase());
      const matchesStatus = !filterStatus || trip.status === filterStatus;
      return matchesSearch && matchesStatus;
    })
    .sort((a, b) => {
      switch (sortBy) {
        case 'alphabetical':
          return a.title.localeCompare(b.title);
        case 'date':
          return new Date(a.startDate).getTime() - new Date(b.startDate).getTime();
        case 'recent':
        default:
          return new Date(b.updatedAt).getTime() - new Date(a.updatedAt).getTime();
      }
    });

  const handleCreateTrip = () => {
    navigate('/trips/new');
    onClose();
  };

  const handleOpenTrip = (trip: Trip) => {
    navigate(`/trips/${trip.id}`);
    dispatch(setActivePanel('trip-planning'));
    onClose();
  };

  const handleDeleteTrip = async (tripId: string) => {
    if (window.confirm('Are you sure you want to delete this trip? This action cannot be undone.')) {
      try {
        await dispatch(deleteTripThunk(tripId)).unwrap();
      } catch (error) {
        // Error handled in thunk
      }
    }
  };

  const getTripIcon = (status: string) => {
    switch (status) {
      case 'active':
        return <MapPin className="w-4 h-4 text-green-600" />;
      case 'completed':
        return <Calendar className="w-4 h-4 text-blue-600" />;
      case 'planning':
      default:
        return <Clock className="w-4 h-4 text-orange-600" />;
    }
  };

  const getPrivacyIcon = (privacy: string) => {
    switch (privacy) {
      case 'public':
        return <Globe className="w-4 h-4 text-gray-500" />;
      case 'friends':
        return <Users className="w-4 h-4 text-gray-500" />;
      case 'private':
      default:
        return <Lock className="w-4 h-4 text-gray-500" />;
    }
  };

  const getTripDuration = (trip: Trip) => {
    const days = differenceInDays(new Date(trip.endDate), new Date(trip.startDate)) + 1;
    return `${days} day${days > 1 ? 's' : ''}`;
  };

  if (!isOpen) return null;

  return (
    <div className="absolute top-16 left-0 w-96 h-[calc(100vh-4rem)] bg-white shadow-2xl z-40 flex flex-col">
      {/* Header */}
      <div className="p-4 border-b bg-gradient-to-r from-indigo-500 to-purple-600 text-white">
        <div className="flex items-center justify-between mb-3">
          <h2 className="text-xl font-bold">My Trips</h2>
          <button
            onClick={onClose}
            className="p-1 hover:bg-white/20 rounded-lg transition-colors"
          >
            ×
          </button>
        </div>
        
        <button
          onClick={handleCreateTrip}
          className="w-full py-2 bg-white/20 hover:bg-white/30 rounded-lg font-medium transition-colors flex items-center justify-center gap-2"
        >
          <Plus className="w-5 h-5" />
          Create New Trip
        </button>
      </div>

      {/* Natural Language Search */}
      <div className="p-4 border-b space-y-2">
        <div className="relative">
          <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400 w-4 h-4" />
          <input
            type="text"
            placeholder="Try: 'active trips', 'planning trips sorted by date', 'completed hiking trips'..."
            value={naturalLanguageQuery}
            onChange={(e) => setNaturalLanguageQuery(e.target.value)}
            className="w-full pl-10 pr-4 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-indigo-500"
          />
        </div>
        
        {/* Help text */}
        <p className="text-xs text-gray-500 px-1">
          Use natural language to filter and sort your trips
        </p>
      </div>

      {/* Storage Mode Info */}
      <div className="px-4 py-2 bg-gray-50 border-b">
        <p className="text-xs text-gray-600">
          {isAuthenticated 
            ? '☁️ Synced to cloud' 
            : '💾 Saved locally (sign in to sync)'
          }
        </p>
      </div>

      {/* Trips List */}
      <div className="flex-1 overflow-y-auto">
        {isLoading ? (
          <div className="flex items-center justify-center h-32">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-indigo-600"></div>
          </div>
        ) : error ? (
          <div className="p-4 text-center text-red-600">
            <p>Failed to load trips</p>
            <button
              onClick={() => dispatch(getUserTripsThunk())}
              className="mt-2 text-indigo-600 hover:text-indigo-700"
            >
              Try again
            </button>
          </div>
        ) : filteredTrips.length === 0 ? (
          <div className="p-8 text-center text-gray-500">
            {trips.length === 0 ? (
              <>
                <MapPin className="w-12 h-12 mx-auto mb-4 text-gray-300" />
                <h3 className="font-semibold mb-2">No trips yet</h3>
                <p className="text-sm mb-4">Start planning your first adventure!</p>
                <button
                  onClick={handleCreateTrip}
                  className="px-4 py-2 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 transition-colors"
                >
                  Create Your First Trip
                </button>
              </>
            ) : (
              <>
                <Search className="w-8 h-8 mx-auto mb-2 text-gray-300" />
                <p>No trips match your search</p>
              </>
            )}
          </div>
        ) : (
          <div className="p-4 space-y-3">
            {filteredTrips.map((trip) => (
              <div
                key={trip.id}
                className="bg-white border rounded-lg p-4 hover:shadow-md transition-shadow cursor-pointer group"
                onClick={() => handleOpenTrip(trip)}
              >
                <div className="flex items-start justify-between">
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2 mb-1">
                      {getTripIcon(trip.status)}
                      <h3 className="font-semibold text-gray-900 truncate">
                        {trip.title}
                      </h3>
                      {getPrivacyIcon(trip.privacy)}
                    </div>
                    
                    <p className="text-sm text-gray-600 mb-2 line-clamp-2">
                      {trip.description}
                    </p>

                    <div className="flex items-center gap-4 text-xs text-gray-500">
                      <div className="flex items-center gap-1">
                        <Calendar className="w-3 h-3" />
                        <span>{format(new Date(trip.startDate), 'MMM d, yyyy')}</span>
                      </div>
                      <div className="flex items-center gap-1">
                        <Clock className="w-3 h-3" />
                        <span>{getTripDuration(trip)}</span>
                      </div>
                      {trip.waypoints.length > 0 && (
                        <div className="flex items-center gap-1">
                          <MapPin className="w-3 h-3" />
                          <span>{trip.waypoints.length} places</span>
                        </div>
                      )}
                    </div>

                    {trip.collaborators.length > 1 && (
                      <div className="flex items-center gap-1 mt-2">
                        <Users className="w-3 h-3 text-gray-400" />
                        <span className="text-xs text-gray-500">
                          {trip.collaborators.length} collaborators
                        </span>
                      </div>
                    )}
                  </div>

                  <div className="flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
                    <button
                      onClick={(e) => {
                        e.stopPropagation();
                        navigate(`/trips/${trip.id}/edit`);
                        onClose();
                      }}
                      className="p-1 hover:bg-gray-100 rounded transition-colors"
                      title="Edit trip"
                    >
                      <Edit className="w-4 h-4 text-gray-500" />
                    </button>
                    <button
                      onClick={(e) => {
                        e.stopPropagation();
                        handleDeleteTrip(trip.id);
                      }}
                      className="p-1 hover:bg-gray-100 rounded transition-colors"
                      title="Delete trip"
                    >
                      <Trash2 className="w-4 h-4 text-red-500" />
                    </button>
                    <ArrowRight className="w-4 h-4 text-gray-400" />
                  </div>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>

      {/* Footer Stats */}
      {trips.length > 0 && (
        <div className="p-4 border-t bg-gray-50 text-center">
          <p className="text-sm text-gray-600">
            {trips.length} trip{trips.length > 1 ? 's' : ''} total
          </p>
        </div>
      )}
    </div>
  );
};