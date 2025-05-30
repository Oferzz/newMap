import { Collection, Place, Trip } from '../../types';

export interface DataService {
  // Collections
  getCollections(): Promise<Collection[]>;
  saveCollection(collection: Omit<Collection, 'id'>): Promise<Collection>;
  updateCollection(id: string, updates: Partial<Collection>): Promise<Collection>;
  deleteCollection(id: string): Promise<void>;
  addLocationToCollection(collectionId: string, location: { latitude: number; longitude: number; name?: string }): Promise<void>;

  // Places
  getPlaces(): Promise<Place[]>;
  savePlace(place: Omit<Place, 'id'>): Promise<Place>;
  updatePlace(id: string, updates: Partial<Place>): Promise<Place>;
  deletePlace(id: string): Promise<void>;

  // Trips
  getTrips(): Promise<Trip[]>;
  saveTrip(trip: Omit<Trip, 'id'>): Promise<Trip>;
  updateTrip(id: string, updates: Partial<Trip>): Promise<Trip>;
  deleteTrip(id: string): Promise<void>;

  // Temporary Markers
  getTemporaryMarkers(): Promise<Array<{ id: string; coordinates: [number, number] }>>;
  saveTemporaryMarker(coordinates: [number, number]): Promise<{ id: string; coordinates: [number, number] }>;
  removeTemporaryMarker(id: string): Promise<void>;
  clearTemporaryMarkers(): Promise<void>;
}