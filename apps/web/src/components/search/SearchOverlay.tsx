import React, { useRef, useEffect } from 'react';
import { MapPin, Route, User, Clock, Users } from 'lucide-react';
import { SearchResult, SearchResults } from '../../types';
import { formatDistanceToNow } from '../../utils/date';

interface SearchOverlayProps {
  results: SearchResults;
  onSelect: (result: SearchResult) => void;
  onClose: () => void;
}

export const SearchOverlay: React.FC<SearchOverlayProps> = ({
  results,
  onSelect,
  onClose,
}) => {
  const overlayRef = useRef<HTMLDivElement>(null);

  // Handle click outside
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (overlayRef.current && !overlayRef.current.contains(event.target as Node)) {
        onClose();
      }
    };

    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, [onClose]);

  // Handle keyboard navigation
  useEffect(() => {
    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key === 'Escape') {
        onClose();
      }
    };

    document.addEventListener('keydown', handleKeyDown);
    return () => document.removeEventListener('keydown', handleKeyDown);
  }, [onClose]);

  const hasResults = 
    results.places.length > 0 || 
    results.trips.length > 0 || 
    results.users.length > 0;

  if (!hasResults) {
    return (
      <div className="absolute top-20 left-1/2 transform -translate-x-1/2 w-full max-w-2xl">
        <div 
          ref={overlayRef}
          className="bg-white rounded-lg shadow-lg p-8 text-center"
        >
          <p className="text-gray-500">No results found</p>
        </div>
      </div>
    );
  }

  return (
    <>
      {/* Semi-transparent backdrop */}
      <div 
        className="fixed inset-0 bg-black bg-opacity-10 z-40"
        onClick={onClose}
      />
      
      {/* Results Panel */}
      <div className="absolute top-20 left-1/2 transform -translate-x-1/2 w-full max-w-2xl z-50">
        <div 
          ref={overlayRef}
          className="bg-white rounded-lg shadow-xl max-h-[calc(100vh-6rem)] overflow-hidden"
        >
          <div className="overflow-y-auto max-h-[calc(100vh-6rem)]">
            {/* Places Section */}
            {results.places.length > 0 && (
              <div className="border-b border-gray-200">
                <div className="p-4 bg-gray-50">
                  <h3 className="text-sm font-semibold text-gray-700 flex items-center">
                    <MapPin className="w-4 h-4 mr-2" />
                    Places ({results.places.length})
                  </h3>
                </div>
                <div className="divide-y divide-gray-100">
                  {results.places.map((place) => (
                    <button
                      key={place.id}
                      className="w-full px-4 py-3 hover:bg-gray-50 transition-colors text-left"
                      onClick={() => onSelect(place)}
                    >
                      <div className="flex items-start justify-between">
                        <div className="flex-1">
                          <h4 className="font-medium text-gray-900">
                            {place.name}
                          </h4>
                          <p className="text-sm text-gray-600 mt-1">
                            {place.street_address && `${place.street_address}, `}
                            {place.city}
                            {place.country && `, ${place.country}`}
                          </p>
                          {place.category && place.category.length > 0 && (
                            <div className="flex flex-wrap gap-1 mt-2">
                              {place.category.slice(0, 3).map((cat) => (
                                <span
                                  key={cat}
                                  className="px-2 py-1 text-xs bg-gray-100 text-gray-600 rounded"
                                >
                                  {cat}
                                </span>
                              ))}
                            </div>
                          )}
                        </div>
                        {place.average_rating && (
                          <div className="ml-4 text-sm text-gray-500">
                            ⭐ {place.average_rating.toFixed(1)}
                          </div>
                        )}
                      </div>
                    </button>
                  ))}
                </div>
              </div>
            )}

            {/* Trips Section */}
            {results.trips.length > 0 && (
              <div className="border-b border-gray-200">
                <div className="p-4 bg-gray-50">
                  <h3 className="text-sm font-semibold text-gray-700 flex items-center">
                    <Route className="w-4 h-4 mr-2" />
                    Trips ({results.trips.length})
                  </h3>
                </div>
                <div className="divide-y divide-gray-100">
                  {results.trips.map((trip) => (
                    <button
                      key={trip.id}
                      className="w-full px-4 py-3 hover:bg-gray-50 transition-colors text-left"
                      onClick={() => onSelect(trip)}
                    >
                      <div className="flex items-start justify-between">
                        <div className="flex-1">
                          <h4 className="font-medium text-gray-900">
                            {trip.title}
                          </h4>
                          {trip.description && (
                            <p className="text-sm text-gray-600 mt-1 line-clamp-2">
                              {trip.description}
                            </p>
                          )}
                          <div className="flex items-center gap-4 mt-2 text-sm text-gray-500">
                            <span className="flex items-center">
                              <MapPin className="w-3 h-3 mr-1" />
                              {trip.waypoints?.length || 0} places
                            </span>
                            <span className="flex items-center">
                              <Users className="w-3 h-3 mr-1" />
                              {trip.collaborators?.length || 0} collaborators
                            </span>
                            <span className="flex items-center">
                              <Clock className="w-3 h-3 mr-1" />
                              {formatDistanceToNow(trip.created_at)}
                            </span>
                          </div>
                        </div>
                        <div className="ml-4">
                          <span className={`
                            px-2 py-1 text-xs rounded-full
                            ${trip.privacy === 'public' 
                              ? 'bg-green-100 text-green-700' 
                              : trip.privacy === 'private'
                              ? 'bg-gray-100 text-gray-700'
                              : 'bg-blue-100 text-blue-700'
                            }
                          `}>
                            {trip.privacy}
                          </span>
                        </div>
                      </div>
                    </button>
                  ))}
                </div>
              </div>
            )}

            {/* Users Section */}
            {results.users.length > 0 && (
              <div>
                <div className="p-4 bg-gray-50">
                  <h3 className="text-sm font-semibold text-gray-700 flex items-center">
                    <User className="w-4 h-4 mr-2" />
                    Users ({results.users.length})
                  </h3>
                </div>
                <div className="divide-y divide-gray-100">
                  {results.users.map((user) => (
                    <button
                      key={user.id}
                      className="w-full px-4 py-3 hover:bg-gray-50 transition-colors text-left"
                      onClick={() => onSelect(user)}
                    >
                      <div className="flex items-center">
                        <img
                          src={user.avatar_url || '/default-avatar.png'}
                          alt={user.display_name}
                          className="w-10 h-10 rounded-full mr-3"
                        />
                        <div className="flex-1">
                          <h4 className="font-medium text-gray-900">
                            {user.display_name}
                          </h4>
                          <p className="text-sm text-gray-600">
                            @{user.username}
                          </p>
                        </div>
                      </div>
                    </button>
                  ))}
                </div>
              </div>
            )}
          </div>
        </div>
      </div>
    </>
  );
};