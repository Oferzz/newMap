import React, { useEffect, useRef } from 'react';
import { X, MapPin, Clock, Users, Share2, Edit, Star, Navigation } from 'lucide-react';
import { useAppSelector, useAppDispatch } from '../../hooks/redux';
import { Place, Trip } from '../../types';
import { formatDate } from '../../utils/date';

interface DetailsPanelProps {
  item: Place | Trip;
  onClose: () => void;
}

export const DetailsPanel: React.FC<DetailsPanelProps> = ({ item, onClose }) => {
  const panelRef = useRef<HTMLDivElement>(null);
  const dispatch = useAppDispatch();
  const user = useAppSelector((state) => state.auth.user);
  
  const isPlace = 'location' in item;
  const isOwner = user?.id === (isPlace ? item.created_by : item.owner_id);

  // Handle escape key
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape') {
        onClose();
      }
    };

    document.addEventListener('keydown', handleKeyDown);
    return () => document.removeEventListener('keydown', handleKeyDown);
  }, [onClose]);

  const handleEdit = () => {
    // Dispatch edit action
    dispatch({ 
      type: 'ui/setActivePanel', 
      payload: isPlace ? 'place-edit' : 'trip-edit' 
    });
  };

  const handleShare = () => {
    // Copy link to clipboard
    const url = `${window.location.origin}/${isPlace ? 'places' : 'trips'}/${item.id}`;
    navigator.clipboard.writeText(url);
    
    // Show toast notification
    dispatch({ 
      type: 'ui/showToast', 
      payload: { message: 'Link copied to clipboard!', type: 'success' } 
    });
  };

  const handleGetDirections = () => {
    if (isPlace && item.location) {
      const [lng, lat] = item.location.coordinates;
      window.open(
        `https://www.google.com/maps/dir/?api=1&destination=${lat},${lng}`,
        '_blank'
      );
    }
  };

  return (
    <div 
      ref={panelRef}
      className="details-panel panel-slide-right open"
    >
      <div className="h-full flex flex-col">
        {/* Header */}
        <div className="flex items-center justify-between p-4 border-b">
          <h2 className="text-lg font-semibold">
            {isPlace ? 'Place Details' : 'Trip Details'}
          </h2>
          <button
            onClick={onClose}
            className="p-2 hover:bg-gray-100 rounded-lg"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        {/* Mobile Handle */}
        <div className="panel-handle md:hidden" />

        {/* Content */}
        <div className="flex-1 overflow-y-auto panel-content">
          {/* Cover Image */}
          {(!isPlace && item.cover_image) && (
            <img 
              src={item.cover_image} 
              alt={item.title}
              className="w-full h-48 object-cover"
            />
          )}
          
          {/* Main Info */}
          <div className="p-4">
            <h1 className="text-2xl font-bold mb-2">
              {isPlace ? item.name : item.title}
            </h1>
            
            {/* Place Info */}
            {isPlace && (
              <>
                {item.average_rating && (
                  <div className="flex items-center mb-2">
                    <Star className="w-4 h-4 text-yellow-500 mr-1" />
                    <span className="font-medium">{item.average_rating.toFixed(1)}</span>
                    <span className="text-gray-600 ml-1">
                      ({item.rating_count} reviews)
                    </span>
                  </div>
                )}
                
                <p className="text-gray-600 flex items-start mb-4">
                  <MapPin className="w-4 h-4 mr-2 mt-1 flex-shrink-0" />
                  <span>
                    {item.street_address && `${item.street_address}, `}
                    {item.city}
                    {item.state && `, ${item.state}`}
                    {(item as any).country && `, ${(item as any).country}`}
                  </span>
                </p>

                {item.category && item.category.length > 0 && (
                  <div className="flex flex-wrap gap-2 mb-4">
                    {item.category.map((cat: string) => (
                      <span
                        key={cat}
                        className="px-3 py-1 bg-gray-100 text-gray-700 rounded-full text-sm"
                      >
                        {cat}
                      </span>
                    ))}
                  </div>
                )}
              </>
            )}

            {/* Trip Info */}
            {!isPlace && (
              <>
                <div className="flex items-center gap-4 text-sm text-gray-600 mb-4">
                  <span className="flex items-center">
                    <MapPin className="w-4 h-4 mr-1" />
                    {item.waypoints?.length || 0} places
                  </span>
                  <span className="flex items-center">
                    <Users className="w-4 h-4 mr-1" />
                    {(item as any).collaborators?.length || 0} collaborators
                  </span>
                  <span className="flex items-center">
                    <Clock className="w-4 h-4 mr-1" />
                    {formatDate(item.created_at)}
                  </span>
                </div>

                {item.start_date && item.end_date && (
                  <p className="text-gray-600 mb-4">
                    {formatDate(item.start_date)} - {formatDate(item.end_date)}
                  </p>
                )}
              </>
            )}

            {/* Description */}
            {item.description && (
              <div className="mb-6">
                <h3 className="font-semibold mb-2">Description</h3>
                <p className="text-gray-700 whitespace-pre-wrap">
                  {item.description}
                </p>
              </div>
            )}

            {/* Opening Hours (Place only) */}
            {isPlace && item.opening_hours && (
              <div className="mb-6">
                <h3 className="font-semibold mb-2">Opening Hours</h3>
                <div className="space-y-1 text-sm">
                  {Object.entries(item.opening_hours).map(([day, hours]) => (
                    <div key={day} className="flex justify-between">
                      <span className="capitalize text-gray-600">{day}</span>
                      <span>
                        {typeof hours === 'string' ? hours : 'Closed'}
                      </span>
                    </div>
                  ))}
                </div>
              </div>
            )}

            {/* Trip Waypoints */}
            {!isPlace && item.waypoints && item.waypoints.length > 0 && (
              <div className="mb-6">
                <h3 className="font-semibold mb-2">Itinerary</h3>
                <div className="space-y-2">
                  {item.waypoints.map((waypoint: any, index: number) => (
                    <div key={waypoint.id} className="flex items-start">
                      <div className="flex-shrink-0 w-6 h-6 bg-blue-500 text-white rounded-full flex items-center justify-center text-xs font-medium">
                        {index + 1}
                      </div>
                      <div className="ml-3 flex-1">
                        <p className="font-medium">{waypoint.place?.name}</p>
                        {waypoint.notes && (
                          <p className="text-sm text-gray-600 mt-1">
                            {waypoint.notes}
                          </p>
                        )}
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            )}
          </div>
        </div>

        {/* Actions */}
        <div className="p-4 border-t space-y-2">
          {isPlace && (
            <button
              onClick={handleGetDirections}
              className="w-full flex items-center justify-center px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700"
            >
              <Navigation className="w-4 h-4 mr-2" />
              Get Directions
            </button>
          )}
          
          <div className="flex gap-2">
            {isOwner && (
              <button
                onClick={handleEdit}
                className="flex-1 flex items-center justify-center px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50"
              >
                <Edit className="w-4 h-4 mr-2" />
                Edit
              </button>
            )}
            
            <button
              onClick={handleShare}
              className="flex-1 flex items-center justify-center px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50"
            >
              <Share2 className="w-4 h-4 mr-2" />
              Share
            </button>
          </div>
        </div>
      </div>
    </div>
  );
};