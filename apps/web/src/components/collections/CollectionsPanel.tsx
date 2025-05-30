import React, { useState, useEffect } from 'react';
import { useAppSelector, useAppDispatch } from '../../hooks/redux';
import { 
  selectCollection,
} from '../../store/slices/collectionsSlice';
import { 
  createCollectionThunk,
  getUserCollectionsThunk,
  deleteCollectionThunk,
  addLocationToCollectionThunk
} from '../../store/thunks/collections.thunks';

interface CollectionsPanelProps {
  isOpen: boolean;
  onClose: () => void;
  locationToAdd?: [number, number];
}

export const CollectionsPanel: React.FC<CollectionsPanelProps> = ({ 
  isOpen, 
  onClose,
  locationToAdd 
}) => {
  const dispatch = useAppDispatch();
  const collections = useAppSelector((state) => state.collections.items);
  const isLoading = useAppSelector((state) => state.collections.isLoading);
  const [isCreating, setIsCreating] = useState(false);
  const [newCollectionName, setNewCollectionName] = useState('');
  const [newCollectionDescription, setNewCollectionDescription] = useState('');

  // Load collections when panel opens
  useEffect(() => {
    if (isOpen) {
      dispatch(getUserCollectionsThunk());
    }
  }, [isOpen, dispatch]);

  const handleCreateCollection = async () => {
    if (!newCollectionName.trim()) return;

    try {
      await dispatch(createCollectionThunk({
        name: newCollectionName,
        description: newCollectionDescription || undefined,
      })).unwrap();

      setNewCollectionName('');
      setNewCollectionDescription('');
      setIsCreating(false);
      
      // Refresh collections list
      dispatch(getUserCollectionsThunk());
    } catch (error) {
      // Error is handled in the thunk
    }
  };

  const handleAddToCollection = async (collectionId: string) => {
    if (!locationToAdd) return;

    try {
      await dispatch(addLocationToCollectionThunk({
        collectionId,
        location: {
          latitude: locationToAdd[1],
          longitude: locationToAdd[0],
        },
      })).unwrap();

      onClose();
    } catch (error) {
      // Error is handled in the thunk
    }
  };

  const handleDeleteCollection = async (collectionId: string) => {
    if (window.confirm('Are you sure you want to delete this collection?')) {
      try {
        await dispatch(deleteCollectionThunk(collectionId)).unwrap();
      } catch (error) {
        // Error is handled in the thunk
      }
    }
  };

  if (!isOpen) return null;

  return (
    <div className="absolute left-4 top-20 bottom-4 w-96 bg-white rounded-lg shadow-xl z-30 flex flex-col">
      <div className="p-6 border-b">
        <div className="flex items-center justify-between">
          <h2 className="text-2xl font-bold">Collections</h2>
          <button
            onClick={onClose}
            className="p-2 hover:bg-gray-100 rounded-full"
          >
            ✕
          </button>
        </div>
        {locationToAdd && (
          <p className="text-sm text-gray-500 mt-2">
            Select a collection to add location
          </p>
        )}
      </div>

      <div className="flex-1 overflow-y-auto p-6">
        {/* Create new collection */}
        {!isCreating ? (
          <button
            onClick={() => setIsCreating(true)}
            className="w-full p-4 border-2 border-dashed border-gray-300 rounded-lg hover:border-blue-500 hover:bg-blue-50 transition-colors"
          >
            <div className="flex items-center justify-center gap-2">
              <span className="text-2xl">+</span>
              <span>Create New Collection</span>
            </div>
          </button>
        ) : (
          <div className="p-4 border rounded-lg mb-4">
            <input
              type="text"
              value={newCollectionName}
              onChange={(e) => setNewCollectionName(e.target.value)}
              placeholder="Collection name"
              className="w-full px-3 py-2 border rounded-md mb-2"
              autoFocus
            />
            <textarea
              value={newCollectionDescription}
              onChange={(e) => setNewCollectionDescription(e.target.value)}
              placeholder="Description (optional)"
              className="w-full px-3 py-2 border rounded-md mb-2"
              rows={2}
            />
            <div className="flex gap-2">
              <button
                onClick={handleCreateCollection}
                disabled={!newCollectionName.trim() || isLoading}
                className="px-4 py-2 bg-blue-500 text-white rounded-md hover:bg-blue-600 disabled:opacity-50"
              >
                Create
              </button>
              <button
                onClick={() => {
                  setIsCreating(false);
                  setNewCollectionName('');
                  setNewCollectionDescription('');
                }}
                className="px-4 py-2 border rounded-md hover:bg-gray-50"
              >
                Cancel
              </button>
            </div>
          </div>
        )}

        {/* Collections list */}
        <div className="space-y-3 mt-4">
          {collections.map((collection) => (
            <div
              key={collection.id}
              className="p-4 border rounded-lg hover:shadow-md transition-shadow"
            >
              <div className="flex items-start justify-between">
                <div className="flex-1">
                  <h3 className="font-semibold">{collection.name}</h3>
                  {collection.description && (
                    <p className="text-sm text-gray-600 mt-1">
                      {collection.description}
                    </p>
                  )}
                  <p className="text-xs text-gray-500 mt-2">
                    {collection.locations.length} location{collection.locations.length !== 1 ? 's' : ''}
                  </p>
                </div>
                <div className="flex gap-2">
                  {locationToAdd && (
                    <button
                      onClick={() => handleAddToCollection(collection.id)}
                      className="p-2 text-blue-500 hover:bg-blue-50 rounded"
                      title="Add location to this collection"
                    >
                      +
                    </button>
                  )}
                  <button
                    onClick={() => dispatch(selectCollection(collection.id))}
                    className="p-2 text-gray-500 hover:bg-gray-50 rounded"
                    title="View collection"
                  >
                    👁
                  </button>
                  <button
                    onClick={() => handleDeleteCollection(collection.id)}
                    className="p-2 text-red-500 hover:bg-red-50 rounded"
                    title="Delete collection"
                  >
                    🗑
                  </button>
                </div>
              </div>
            </div>
          ))}
        </div>

        {collections.length === 0 && !isCreating && (
          <p className="text-center text-gray-500 mt-8">
            No collections yet. Create your first collection!
          </p>
        )}
      </div>
    </div>
  );
};