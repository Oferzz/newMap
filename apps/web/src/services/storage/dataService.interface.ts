import { Collection, Place, Trip } from '../../types';
import { Waypoint } from '../../store/slices/tripsSlice';

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
  getTrip(id: string): Promise<Trip | null>;
  saveTrip(trip: Omit<Trip, 'id'>): Promise<Trip>;
  updateTrip(id: string, updates: Partial<Trip>): Promise<Trip>;
  deleteTrip(id: string): Promise<void>;
  
  // Waypoints
  addWaypoint(tripId: string, waypoint: Omit<Waypoint, 'id'>): Promise<Waypoint>;
  updateWaypoint(tripId: string, waypointId: string, updates: Partial<Waypoint>): Promise<Waypoint>;
  removeWaypoint(tripId: string, waypointId: string): Promise<void>;
  reorderWaypoints(tripId: string, waypointIds: string[]): Promise<void>;

  // Temporary Markers
  getTemporaryMarkers(): Promise<Array<{ id: string; coordinates: [number, number] }>>;
  saveTemporaryMarker(coordinates: [number, number]): Promise<{ id: string; coordinates: [number, number] }>;
  removeTemporaryMarker(id: string): Promise<void>;
  clearTemporaryMarkers(): Promise<void>;
}