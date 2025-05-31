import React, { useState, useEffect } from 'react';
import { useAppDispatch, useAppSelector } from '../../hooks/redux';
import { createPlaceThunk } from '../../store/thunks/places.thunks';
import { clearTemporaryMarkers, addNotification } from '../../store/slices/uiSlice';

interface PlaceCreationPanelProps {
  isOpen: boolean;
  onClose: () => void;
}

export const PlaceCreationPanel: React.FC<PlaceCreationPanelProps> = ({ isOpen, onClose }) => {
  const dispatch = useAppDispatch();
  const contextMenuCoordinates = useAppSelector((state) => state.ui.contextMenuState.coordinates);
  const temporaryMarkers = useAppSelector((state) => state.ui.temporaryMarkers);
  
  const [formData, setFormData] = useState({
    name: '',
    description: '',
    category: 'poi',
    tags: '',
    address: '',
    city: '',
    state: '',
    country: '',
    postalCode: '',
  });

  const [isSubmitting, setIsSubmitting] = useState(false);

  // Get coordinates from either context menu or the latest temporary marker
  const coordinates = contextMenuCoordinates || 
    (temporaryMarkers.length > 0 ? temporaryMarkers[temporaryMarkers.length - 1].coordinates : null);

  useEffect(() => {
    if (isOpen && coordinates) {
      // Could potentially reverse geocode here to get address
      setFormData(prev => ({
        ...prev,
        // Reset form when panel opens
        name: '',
        description: '',
      }));
    }
  }, [isOpen, coordinates]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    if (!coordinates) {
      dispatch(addNotification({
        type: 'error',
        message: 'No location selected. Please click on the map first.',
      }));
      return;
    }

    setIsSubmitting(true);

    try {
      await dispatch(createPlaceThunk({
        name: formData.name,
        description: formData.description,
        latitude: coordinates[1],
        longitude: coordinates[0],
        category: formData.category,
        address: formData.address,
      }) as any).unwrap();

      dispatch(addNotification({
        type: 'success',
        message: 'Place created successfully!',
      }));

      // Clear temporary markers after successful creation
      dispatch(clearTemporaryMarkers());
      
      // Close panel
      onClose();
    } catch (error) {
      dispatch(addNotification({
        type: 'error',
        message: 'Failed to create place. Please try again.',
      }));
    } finally {
      setIsSubmitting(false);
    }
  };

  if (!isOpen) return null;

  return (
    <div className="absolute left-4 top-20 bottom-4 w-96 bg-white rounded-lg shadow-xl z-30 flex flex-col">
      <div className="p-6 border-b">
        <div className="flex items-center justify-between mb-2">
          <h2 className="text-2xl font-bold">Create New Place</h2>
          <button
            onClick={onClose}
            className="p-2 hover:bg-gray-100 rounded-full"
          >
            ✕
          </button>
        </div>
        {coordinates && (
          <p className="text-sm text-gray-500">
            Location: {coordinates[0].toFixed(6)}, {coordinates[1].toFixed(6)}
          </p>
        )}
      </div>

      <form onSubmit={handleSubmit} className="flex-1 overflow-y-auto p-6">
        <div className="space-y-4">
          <div>
            <label className="block text-sm font-medium mb-1">Name *</label>
            <input
              type="text"
              required
              value={formData.name}
              onChange={(e) => setFormData({ ...formData, name: e.target.value })}
              className="w-full px-3 py-2 border rounded-lg focus:ring-2 focus:ring-blue-500"
              placeholder="Enter place name"
            />
          </div>

          <div>
            <label className="block text-sm font-medium mb-1">Description</label>
            <textarea
              value={formData.description}
              onChange={(e) => setFormData({ ...formData, description: e.target.value })}
              className="w-full px-3 py-2 border rounded-lg focus:ring-2 focus:ring-blue-500"
              rows={3}
              placeholder="Enter description"
            />
          </div>

          <div>
            <label className="block text-sm font-medium mb-1">Category</label>
            <select
              value={formData.category}
              onChange={(e) => setFormData({ ...formData, category: e.target.value })}
              className="w-full px-3 py-2 border rounded-lg focus:ring-2 focus:ring-blue-500"
            >
              <option value="poi">Point of Interest</option>
              <option value="area">Area</option>
              <option value="region">Region</option>
            </select>
          </div>

          <div>
            <label className="block text-sm font-medium mb-1">Tags</label>
            <input
              type="text"
              value={formData.tags}
              onChange={(e) => setFormData({ ...formData, tags: e.target.value })}
              className="w-full px-3 py-2 border rounded-lg focus:ring-2 focus:ring-blue-500"
              placeholder="restaurant, scenic, family-friendly (comma separated)"
            />
          </div>

          <div className="border-t pt-4">
            <h3 className="font-medium mb-3">Address Information</h3>
            
            <div className="space-y-3">
              <div>
                <label className="block text-sm font-medium mb-1">Street Address</label>
                <input
                  type="text"
                  value={formData.address}
                  onChange={(e) => setFormData({ ...formData, address: e.target.value })}
                  className="w-full px-3 py-2 border rounded-lg focus:ring-2 focus:ring-blue-500"
                  placeholder="123 Main St"
                />
              </div>

              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="block text-sm font-medium mb-1">City</label>
                  <input
                    type="text"
                    value={formData.city}
                    onChange={(e) => setFormData({ ...formData, city: e.target.value })}
                    className="w-full px-3 py-2 border rounded-lg focus:ring-2 focus:ring-blue-500"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium mb-1">State</label>
                  <input
                    type="text"
                    value={formData.state}
                    onChange={(e) => setFormData({ ...formData, state: e.target.value })}
                    className="w-full px-3 py-2 border rounded-lg focus:ring-2 focus:ring-blue-500"
                  />
                </div>
              </div>

              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="block text-sm font-medium mb-1">Country</label>
                  <input
                    type="text"
                    value={formData.country}
                    onChange={(e) => setFormData({ ...formData, country: e.target.value })}
                    className="w-full px-3 py-2 border rounded-lg focus:ring-2 focus:ring-blue-500"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium mb-1">Postal Code</label>
                  <input
                    type="text"
                    value={formData.postalCode}
                    onChange={(e) => setFormData({ ...formData, postalCode: e.target.value })}
                    className="w-full px-3 py-2 border rounded-lg focus:ring-2 focus:ring-blue-500"
                  />
                </div>
              </div>
            </div>
          </div>
        </div>
      </form>

      <div className="p-6 border-t">
        <div className="flex gap-3">
          <button
            type="submit"
            onClick={handleSubmit}
            disabled={isSubmitting || !formData.name}
            className="flex-1 bg-blue-600 text-white py-2 px-4 rounded-lg hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {isSubmitting ? 'Creating...' : 'Create Place'}
          </button>
          <button
            type="button"
            onClick={onClose}
            className="px-4 py-2 border rounded-lg hover:bg-gray-50"
          >
            Cancel
          </button>
        </div>
      </div>
    </div>
  );
};