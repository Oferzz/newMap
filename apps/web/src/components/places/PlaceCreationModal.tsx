import React, { useState, useEffect } from 'react';
import { useAppDispatch, useAppSelector } from '../../hooks/redux';
import { X, MapPin, Save, Loader2 } from 'lucide-react';
import { createPlaceThunk } from '../../store/thunks/places.thunks';
import { addLocationToCollectionThunk } from '../../store/thunks/collections.thunks';
import { addWaypointThunk } from '../../store/thunks/trips.thunks';
import { removeTemporaryMarker } from '../../store/slices/uiSlice';

interface PlaceCreationModalProps {
  isOpen: boolean;
  onClose: () => void;
  coordinates?: [number, number] | null;
  context?: 'general' | 'collection' | 'trip';
  contextId?: string; // collection or trip ID
  tripDay?: number; // for trip context
}

export const PlaceCreationModal: React.FC<PlaceCreationModalProps> = ({
  isOpen,
  onClose,
  coordinates,
  context = 'general',
  contextId,
  tripDay = 1
}) => {
  const dispatch = useAppDispatch();
  const { temporaryMarkers } = useAppSelector(state => state.ui);
  const { items: collections } = useAppSelector(state => state.collections);
  const { currentTrip } = useAppSelector(state => state.trips);
  const { isAuthenticated } = useAppSelector(state => state.auth);
  
  const [formData, setFormData] = useState({
    name: '',
    description: '',
    category: 'restaurant',
    address: '',
    website: '',
    phone: '',
    notes: ''
  });
  
  const [selectedCollection, setSelectedCollection] = useState<string>('');
  const [addToTrip, setAddToTrip] = useState(context === 'trip');
  const [selectedTrip, setSelectedTrip] = useState<string>(contextId || '');
  const [waypointTime, setWaypointTime] = useState({
    arrivalTime: '10:00',
    departureTime: '12:00',
    day: tripDay
  });
  
  const [isLoading, setIsLoading] = useState(false);

  // Get coordinates from temporary markers if not provided
  const placeCoordinates = coordinates || (temporaryMarkers.length > 0 ? temporaryMarkers[0].coordinates : null);

  useEffect(() => {
    if (context === 'collection' && contextId) {
      setSelectedCollection(contextId);
    }
    if (context === 'trip' && contextId) {
      setSelectedTrip(contextId);
      setAddToTrip(true);
    }
  }, [context, contextId]);

  // Auto-populate address from coordinates
  useEffect(() => {
    if (placeCoordinates && !formData.address) {
      // You could implement reverse geocoding here
      setFormData(prev => ({
        ...prev,
        address: `${placeCoordinates[1].toFixed(4)}, ${placeCoordinates[0].toFixed(4)}`
      }));
    }
  }, [placeCoordinates, formData.address]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    if (!placeCoordinates) {
      return;
    }

    setIsLoading(true);
    
    try {
      // Create the place
      const place = await dispatch(createPlaceThunk({
        name: formData.name,
        description: formData.description,
        category: formData.category,
        latitude: placeCoordinates[1],
        longitude: placeCoordinates[0],
        address: formData.address,
        website: formData.website || undefined,
        phone: formData.phone || undefined,
        notes: formData.notes || undefined
      })).unwrap();

      // Add to collection if selected
      if (selectedCollection && collections.find(c => c.id === selectedCollection)) {
        await dispatch(addLocationToCollectionThunk({
          collectionId: selectedCollection,
          location: {
            latitude: placeCoordinates[1],
            longitude: placeCoordinates[0],
            name: formData.name
          }
        }));
      }

      // Add to trip if selected
      if (addToTrip && selectedTrip) {
        await dispatch(addWaypointThunk({
          tripId: selectedTrip,
          waypoint: {
            day: waypointTime.day,
            placeId: place.id,
            place: {
              id: place.id,
              name: place.name,
              address: place.address || '',
              category: place.category,
              coordinates: { lat: placeCoordinates[1], lng: placeCoordinates[0] }
            },
            arrivalTime: waypointTime.arrivalTime,
            departureTime: waypointTime.departureTime,
            notes: formData.notes
          }
        }));
      }

      // Clean up temporary markers
      if (temporaryMarkers.length > 0) {
        dispatch(removeTemporaryMarker(temporaryMarkers[0].id));
      }

      onClose();
      resetForm();
    } catch (error) {
      console.error('Failed to create place:', error);
    } finally {
      setIsLoading(false);
    }
  };

  const resetForm = () => {
    setFormData({
      name: '',
      description: '',
      category: 'restaurant',
      address: '',
      website: '',
      phone: '',
      notes: ''
    });
    setSelectedCollection('');
    setSelectedTrip(contextId || '');
    setAddToTrip(context === 'trip');
    setWaypointTime({
      arrivalTime: '10:00',
      departureTime: '12:00',
      day: tripDay
    });
  };

  const handleClose = () => {
    // Clean up temporary markers
    if (temporaryMarkers.length > 0) {
      dispatch(removeTemporaryMarker(temporaryMarkers[0].id));
    }
    onClose();
    resetForm();
  };

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
      <div className="bg-white rounded-lg max-w-md w-full max-h-[90vh] overflow-y-auto">
        <div className="flex items-center justify-between p-4 border-b">
          <div className="flex items-center gap-2">
            <MapPin className="w-5 h-5 text-indigo-600" />
            <h2 className="text-lg font-semibold">Add New Place</h2>
          </div>
          <button
            onClick={handleClose}
            className="p-1 hover:bg-gray-100 rounded-lg transition-colors"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        <form onSubmit={handleSubmit} className="p-4 space-y-4">
          {/* Coordinates display */}
          {placeCoordinates && (
            <div className="bg-gray-50 p-3 rounded-lg">
              <p className="text-sm text-gray-600">
                <strong>Location:</strong> {placeCoordinates[1].toFixed(4)}, {placeCoordinates[0].toFixed(4)}
              </p>
            </div>
          )}

          {/* Basic info */}
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Name *
            </label>
            <input
              type="text"
              required
              value={formData.name}
              onChange={(e) => setFormData(prev => ({ ...prev, name: e.target.value }))}
              className="w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-indigo-500"
              placeholder="Restaurant name, landmark, etc."
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Category
            </label>
            <select
              value={formData.category}
              onChange={(e) => setFormData(prev => ({ ...prev, category: e.target.value }))}
              className="w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-indigo-500"
            >
              <option value="restaurant">Restaurant</option>
              <option value="attraction">Attraction</option>
              <option value="hotel">Hotel</option>
              <option value="museum">Museum</option>
              <option value="park">Park</option>
              <option value="shopping">Shopping</option>
              <option value="entertainment">Entertainment</option>
              <option value="transport">Transport</option>
              <option value="other">Other</option>
            </select>
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Address
            </label>
            <input
              type="text"
              value={formData.address}
              onChange={(e) => setFormData(prev => ({ ...prev, address: e.target.value }))}
              className="w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-indigo-500"
              placeholder="Street address or description"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Description
            </label>
            <textarea
              value={formData.description}
              onChange={(e) => setFormData(prev => ({ ...prev, description: e.target.value }))}
              className="w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-indigo-500"
              rows={3}
              placeholder="What makes this place special?"
            />
          </div>

          {/* Optional fields */}
          <div className="grid grid-cols-1 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Website
              </label>
              <input
                type="url"
                value={formData.website}
                onChange={(e) => setFormData(prev => ({ ...prev, website: e.target.value }))}
                className="w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-indigo-500"
                placeholder="https://..."
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Phone
              </label>
              <input
                type="tel"
                value={formData.phone}
                onChange={(e) => setFormData(prev => ({ ...prev, phone: e.target.value }))}
                className="w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-indigo-500"
                placeholder="+1 (555) 123-4567"
              />
            </div>
          </div>

          {/* Collection selection */}
          {isAuthenticated && collections.length > 0 && (
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Add to Collection (optional)
              </label>
              <select
                value={selectedCollection}
                onChange={(e) => setSelectedCollection(e.target.value)}
                className="w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-indigo-500"
              >
                <option value="">Select a collection...</option>
                {collections.map(collection => (
                  <option key={collection.id} value={collection.id}>
                    {collection.name}
                  </option>
                ))}
              </select>
            </div>
          )}

          {/* Trip selection */}
          {(context !== 'trip' || !contextId) && (
            <div>
              <label className="flex items-center gap-2">
                <input
                  type="checkbox"
                  checked={addToTrip}
                  onChange={(e) => setAddToTrip(e.target.checked)}
                  className="rounded border-gray-300 text-indigo-600 focus:ring-indigo-500"
                />
                <span className="text-sm font-medium text-gray-700">Add to trip itinerary</span>
              </label>
            </div>
          )}

          {addToTrip && (
            <div className="space-y-3 bg-gray-50 p-3 rounded-lg">
              {!contextId && (
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">
                    Trip
                  </label>
                  <select
                    value={selectedTrip}
                    onChange={(e) => setSelectedTrip(e.target.value)}
                    className="w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-indigo-500"
                  >
                    <option value="">Select a trip...</option>
                    {currentTrip && (
                      <option value={currentTrip.id}>{currentTrip.title}</option>
                    )}
                  </select>
                </div>
              )}
              
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Day
                </label>
                <input
                  type="number"
                  min="1"
                  value={waypointTime.day}
                  onChange={(e) => setWaypointTime(prev => ({ ...prev, day: parseInt(e.target.value) }))}
                  className="w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-indigo-500"
                />
              </div>
              
              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">
                    Arrival
                  </label>
                  <input
                    type="time"
                    value={waypointTime.arrivalTime}
                    onChange={(e) => setWaypointTime(prev => ({ ...prev, arrivalTime: e.target.value }))}
                    className="w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-indigo-500"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">
                    Departure
                  </label>
                  <input
                    type="time"
                    value={waypointTime.departureTime}
                    onChange={(e) => setWaypointTime(prev => ({ ...prev, departureTime: e.target.value }))}
                    className="w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-indigo-500"
                  />
                </div>
              </div>
            </div>
          )}

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Notes
            </label>
            <textarea
              value={formData.notes}
              onChange={(e) => setFormData(prev => ({ ...prev, notes: e.target.value }))}
              className="w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-indigo-500"
              rows={2}
              placeholder="Any additional notes or reminders"
            />
          </div>

          {/* Actions */}
          <div className="flex gap-3 pt-4 border-t">
            <button
              type="button"
              onClick={handleClose}
              className="flex-1 py-2 px-4 border border-gray-300 rounded-lg text-gray-700 hover:bg-gray-50 transition-colors"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={isLoading || !formData.name || !placeCoordinates}
              className="flex-1 py-2 px-4 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors flex items-center justify-center gap-2"
            >
              {isLoading ? (
                <>
                  <Loader2 className="w-4 h-4 animate-spin" />
                  Creating...
                </>
              ) : (
                <>
                  <Save className="w-4 h-4" />
                  Create Place
                </>
              )}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};