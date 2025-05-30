import { Collection, Place, Trip } from '../../types';
import { DataService } from './dataService.interface';
import { collectionsService } from '../collections.service';
import { placesService } from '../places.service';
import { tripsService } from '../trips.service';

export class CloudDataService implements DataService {
  // Collections
  async getCollections(): Promise<Collection[]> {
    const response = await collectionsService.getUserCollections();
    return response.data || [];
  }

  async saveCollection(collection: Omit<Collection, 'id'>): Promise<Collection> {
    const response = await collectionsService.createCollection({
      name: collection.name,
      description: collection.description || undefined,
      privacy: collection.privacy || 'private'
    });
    return response.data;
  }

  async updateCollection(id: string, updates: Partial<Collection>): Promise<Collection> {
    const response = await collectionsService.updateCollection(id, {
      name: updates.name,
      description: updates.description,
      privacy: updates.privacy
    });
    return response.data;
  }

  async deleteCollection(id: string): Promise<void> {
    await collectionsService.deleteCollection(id);
  }

  async addLocationToCollection(collectionId: string, location: { latitude: number; longitude: number; name?: string }): Promise<void> {
    await collectionsService.addLocationToCollection(collectionId, {
      latitude: location.latitude,
      longitude: location.longitude,
      name: location.name
    });
  }

  // Places
  async getPlaces(): Promise<Place[]> {
    const response = await placesService.getPlaces({});
    return response.data || [];
  }

  async savePlace(place: Omit<Place, 'id'>): Promise<Place> {
    const response = await placesService.createPlace(place);
    return response.data;
  }

  async updatePlace(id: string, updates: Partial<Place>): Promise<Place> {
    const response = await placesService.updatePlace(id, updates);
    return response.data;
  }

  async deletePlace(id: string): Promise<void> {
    await placesService.deletePlace(id);
  }

  // Trips
  async getTrips(): Promise<Trip[]> {
    const response = await tripsService.getTrips();
    return response.data || [];
  }

  async saveTrip(trip: Omit<Trip, 'id'>): Promise<Trip> {
    const response = await tripsService.createTrip(trip);
    return response.data;
  }

  async updateTrip(id: string, updates: Partial<Trip>): Promise<Trip> {
    const response = await tripsService.updateTrip(id, updates);
    return response.data;
  }

  async deleteTrip(id: string): Promise<void> {
    await tripsService.deleteTrip(id);
  }

  // Temporary Markers - Store in memory only for cloud service
  private temporaryMarkers: Array<{ id: string; coordinates: [number, number] }> = [];

  async getTemporaryMarkers(): Promise<Array<{ id: string; coordinates: [number, number] }>> {
    return Promise.resolve(this.temporaryMarkers);
  }

  async saveTemporaryMarker(coordinates: [number, number]): Promise<{ id: string; coordinates: [number, number] }> {
    const marker = {
      id: `marker_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`,
      coordinates
    };
    
    this.temporaryMarkers.push(marker);
    return Promise.resolve(marker);
  }

  async removeTemporaryMarker(id: string): Promise<void> {
    this.temporaryMarkers = this.temporaryMarkers.filter(m => m.id !== id);
    return Promise.resolve();
  }

  async clearTemporaryMarkers(): Promise<void> {
    this.temporaryMarkers = [];
    return Promise.resolve();
  }
}