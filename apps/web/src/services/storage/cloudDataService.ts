import { Collection, Place, Trip } from '../../types';
import { Waypoint } from '../../store/slices/tripsSlice';
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
      privacy: collection.privacy as 'public' | 'friends' | 'private' || 'private'
    });
    return response;
  }

  async updateCollection(id: string, updates: Partial<Collection>): Promise<Collection> {
    const response = await collectionsService.updateCollection(id, {
      name: updates.name,
      description: updates.description || undefined,
      privacy: updates.privacy as 'public' | 'friends' | 'private' | undefined
    });
    return response;
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
    const response = await placesService.getUserPlaces({});
    return response.data || [];
  }

  async savePlace(place: Omit<Place, 'id'>): Promise<Place> {
    const response = await placesService.create(place as any);
    return response;
  }

  async updatePlace(id: string, updates: Partial<Place>): Promise<Place> {
    const response = await placesService.update(id, updates as any);
    return response;
  }

  async deletePlace(id: string): Promise<void> {
    await placesService.delete(id);
  }

  // Trips
  async getTrips(): Promise<Trip[]> {
    const response = await tripsService.getUserTrips();
    return response.data || [];
  }

  async saveTrip(trip: Omit<Trip, 'id'>): Promise<Trip> {
    const response = await tripsService.create(trip as any);
    return response;
  }

  async updateTrip(id: string, updates: Partial<Trip>): Promise<Trip> {
    const response = await tripsService.update(id, updates as any);
    return response;
  }

  async deleteTrip(id: string): Promise<void> {
    await tripsService.delete(id);
  }
  
  async getTrip(id: string): Promise<Trip | null> {
    try {
      const response = await tripsService.getById(id);
      return response as Trip;
    } catch (error) {
      return null;
    }
  }

  // Waypoints
  async addWaypoint(tripId: string, waypoint: Omit<Waypoint, 'id'>): Promise<Waypoint> {
    const response = await tripsService.addWaypoint(tripId, {
      placeId: waypoint.placeId,
      arrivalTime: waypoint.arrivalTime,
      departureTime: waypoint.departureTime,
      notes: waypoint.notes
    } as any);
    
    // Transform API response to match Waypoint type
    return {
      id: (response as any).id,
      placeId: (response as any).placeId,
      place: waypoint.place, // This should be populated by the service
      arrivalTime: (response as any).arrivalTime,
      departureTime: (response as any).departureTime,
      notes: (response as any).notes
    } as Waypoint;
  }

  async updateWaypoint(tripId: string, waypointId: string, updates: Partial<Waypoint>): Promise<Waypoint> {
    const response = await tripsService.updateWaypoint(tripId, waypointId, {
      arrivalTime: updates.arrivalTime,
      departureTime: updates.departureTime,
      notes: updates.notes
    } as any);
    
    return {
      id: (response as any).id,
      placeId: (response as any).placeId,
      place: updates.place || (response as any).place,
      arrivalTime: (response as any).arrivalTime,
      departureTime: (response as any).departureTime,
      notes: (response as any).notes
    } as Waypoint;
  }

  async removeWaypoint(tripId: string, waypointId: string): Promise<void> {
    await tripsService.removeWaypoint(tripId, waypointId);
  }

  async reorderWaypoints(tripId: string, waypointIds: string[]): Promise<void> {
    await tripsService.reorderWaypoints(tripId, waypointIds);
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