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
      <div className="absolute top-full mt-1 left-0 right-0 z-50">
        <div 
          ref={overlayRef}
          className="bg-terrain-100 rounded-lg shadow-xl p-8 text-center border border-terrain-300 animate-slideDown"
          style={{ backgroundColor: '#faf8f5' }}
        >
          <p className="text-trail-500">No results found</p>
        </div>
      </div>
    );
  }

  return (
    <div className="absolute top-full mt-1 left-0 right-0 z-50">
      <div 
        ref={overlayRef}
        className="bg-terrain-100 rounded-lg shadow-xl max-h-[calc(100vh-10rem)] overflow-hidden border border-terrain-300 animate-slideDown"
        style={{ backgroundColor: '#faf8f5' }}
      >
          <div className="overflow-y-auto max-h-[calc(100vh-6rem)]">
            {/* Places Section */}
            {results.places.length > 0 && (
              <div className="border-b border-terrain-300">
                <div className="p-4 bg-terrain-100">
                  <h3 className="text-sm font-semibold text-trail-700 flex items-center">
                    <MapPin className="w-4 h-4 mr-2" />
                    Places ({results.places.length})
                  </h3>
                </div>
                <div className="divide-y divide-terrain-200">
                  {results.places.map((place) => (
                    <button
                      key={place.id}
                      className="w-full px-4 py-3 hover:bg-terrain-100 transition-colors text-left"
                      onClick={() => onSelect({ ...place, type: 'place' as const })}
                    >
                      <h4 className="font-medium text-trail-800">
                        {place.name}
                      </h4>
                    </button>
                  ))}
                </div>
              </div>
            )}

            {/* Trips Section */}
            {results.trips.length > 0 && (
              <div className="border-b border-terrain-300">
                <div className="p-4 bg-terrain-100">
                  <h3 className="text-sm font-semibold text-trail-700 flex items-center">
                    <Route className="w-4 h-4 mr-2" />
                    Trips ({results.trips.length})
                  </h3>
                </div>
                <div className="divide-y divide-terrain-200">
                  {results.trips.map((trip) => (
                    <button
                      key={trip.id}
                      className="w-full px-4 py-3 hover:bg-terrain-100 transition-colors text-left"
                      onClick={() => onSelect({ ...trip, name: trip.title || (trip as any).name || '', type: 'trip' as const })}
                    >
                      <div className="flex items-start justify-between">
                        <div className="flex-1">
                          <h4 className="font-medium text-trail-800">
                            {trip.title || (trip as any).name}
                          </h4>
                          {trip.description && (
                            <p className="text-sm text-trail-600 mt-1 line-clamp-2">
                              {trip.description}
                            </p>
                          )}
                          <div className="flex items-center gap-4 mt-2 text-sm text-trail-500">
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
                              {trip.created_at ? formatDistanceToNow(trip.created_at) : 'Unknown'}
                            </span>
                          </div>
                        </div>
                        <div className="ml-4">
                          <span className={`
                            px-2 py-1 text-xs rounded-full
                            ${trip.privacy === 'public' 
                              ? 'bg-forest-100 text-forest-700' 
                              : trip.privacy === 'private'
                              ? 'bg-terrain-200 text-trail-700'
                              : 'bg-water-100 text-water-700'
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
                <div className="p-4 bg-terrain-100">
                  <h3 className="text-sm font-semibold text-trail-700 flex items-center">
                    <User className="w-4 h-4 mr-2" />
                    Users ({results.users.length})
                  </h3>
                </div>
                <div className="divide-y divide-terrain-200">
                  {results.users.map((user) => (
                    <button
                      key={user.id}
                      className="w-full px-4 py-3 hover:bg-terrain-100 transition-colors text-left"
                      onClick={() => onSelect(user)}
                    >
                      <div className="flex items-center">
                        <img
                          src={user.avatar_url || '/default-avatar.png'}
                          alt={user.display_name}
                          className="w-10 h-10 rounded-full mr-3"
                        />
                        <div className="flex-1">
                          <h4 className="font-medium text-trail-800">
                            {user.display_name}
                          </h4>
                          <p className="text-sm text-trail-600">
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
  );
};