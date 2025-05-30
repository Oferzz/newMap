import { Collection, Place, Trip } from '../../types';
import { DataService } from './dataService.interface';
import { localStorageService } from './localStorage.service';

export class LocalDataService implements DataService {
  // Collections
  async getCollections(): Promise<Collection[]> {
    return Promise.resolve(localStorageService.getCollections());
  }

  async saveCollection(collection: Omit<Collection, 'id'>): Promise<Collection> {
    const newCollection: Collection = {
      ...collection,
      id: `col_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`,
      locations: [],
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString()
    };
    
    localStorageService.saveCollection(newCollection);
    return Promise.resolve(newCollection);
  }

  async updateCollection(id: string, updates: Partial<Collection>): Promise<Collection> {
    const collections = localStorageService.getCollections();
    const collection = collections.find(c => c.id === id);
    
    if (!collection) {
      throw new Error('Collection not found');
    }
    
    const updatedCollection: Collection = {
      ...collection,
      ...updates,
      id: collection.id, // Ensure ID doesn't change
      updated_at: new Date().toISOString()
    };
    
    localStorageService.saveCollection(updatedCollection);
    return Promise.resolve(updatedCollection);
  }

  async deleteCollection(id: string): Promise<void> {
    localStorageService.deleteCollection(id);
    return Promise.resolve();
  }

  async addLocationToCollection(collectionId: string, location: { latitude: number; longitude: number; name?: string }): Promise<void> {
    localStorageService.addLocationToCollection(collectionId, location);
    return Promise.resolve();
  }

  // Places
  async getPlaces(): Promise<Place[]> {
    return Promise.resolve(localStorageService.getPlaces());
  }

  async savePlace(place: Omit<Place, 'id'>): Promise<Place> {
    const newPlace: Place = {
      ...place,
      id: `place_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`,
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
      // Ensure required fields for the API format
      street_address: '',
      city: '',
      state: '',
      country: '',
      postal_code: '',
      tags: []
    } as Place;
    
    localStorageService.savePlace(newPlace);
    return Promise.resolve(newPlace);
  }

  async updatePlace(id: string, updates: Partial<Place>): Promise<Place> {
    const places = localStorageService.getPlaces();
    const place = places.find(p => p.id === id);
    
    if (!place) {
      throw new Error('Place not found');
    }
    
    const updatedPlace: Place = {
      ...place,
      ...updates,
      id: place.id,
      updated_at: new Date().toISOString()
    } as Place;
    
    localStorageService.savePlace(updatedPlace);
    return Promise.resolve(updatedPlace);
  }

  async deletePlace(id: string): Promise<void> {
    localStorageService.deletePlace(id);
    return Promise.resolve();
  }

  // Trips
  async getTrips(): Promise<Trip[]> {
    return Promise.resolve(localStorageService.getTrips());
  }

  async saveTrip(trip: Omit<Trip, 'id'>): Promise<Trip> {
    const newTrip: Trip = {
      ...trip,
      id: `trip_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`,
      collaborators: [],
      waypoints: [],
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString()
    } as Trip;
    
    localStorageService.saveTrip(newTrip);
    return Promise.resolve(newTrip);
  }

  async updateTrip(id: string, updates: Partial<Trip>): Promise<Trip> {
    const trips = localStorageService.getTrips();
    const trip = trips.find(t => t.id === id);
    
    if (!trip) {
      throw new Error('Trip not found');
    }
    
    const updatedTrip: Trip = {
      ...trip,
      ...updates,
      id: trip.id,
      updated_at: new Date().toISOString()
    } as Trip;
    
    localStorageService.saveTrip(updatedTrip);
    return Promise.resolve(updatedTrip);
  }

  async deleteTrip(id: string): Promise<void> {
    localStorageService.deleteTrip(id);
    return Promise.resolve();
  }

  // Temporary Markers
  async getTemporaryMarkers(): Promise<Array<{ id: string; coordinates: [number, number] }>> {
    return Promise.resolve(localStorageService.getTemporaryMarkers());
  }

  async saveTemporaryMarker(coordinates: [number, number]): Promise<{ id: string; coordinates: [number, number] }> {
    const marker = {
      id: `marker_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`,
      coordinates
    };
    
    localStorageService.saveTemporaryMarker(marker);
    return Promise.resolve(marker);
  }

  async removeTemporaryMarker(id: string): Promise<void> {
    localStorageService.removeTemporaryMarker(id);
    return Promise.resolve();
  }

  async clearTemporaryMarkers(): Promise<void> {
    localStorageService.clearTemporaryMarkers();
    return Promise.resolve();
  }
}