import { Collection, Place, Trip } from '../../types';
import { Waypoint } from '../../store/slices/tripsSlice';
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
      // Ensure required fields have defaults
      type: (place as any).type || 'poi',
      tags: (place as any).tags || [],
      street_address: (place as any).street_address || '',
      city: (place as any).city || '',
      state: (place as any).state || '',
      country: (place as any).country || '',
      postal_code: (place as any).postal_code || ''
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
    const newTrip: any = {
      ...trip,
      id: `trip_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`,
      collaborators: trip.collaborators || [],
      waypoints: trip.waypoints || [],
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString()
    };
    
    localStorageService.saveTrip(newTrip);
    return Promise.resolve(newTrip as Trip);
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
  
  async getTrip(id: string): Promise<Trip | null> {
    const trips = localStorageService.getTrips();
    const trip = trips.find(t => t.id === id);
    return Promise.resolve(trip || null);
  }

  // Waypoints
  async addWaypoint(tripId: string, waypoint: Omit<Waypoint, 'id'>): Promise<Waypoint> {
    const trips = localStorageService.getTrips();
    const trip = trips.find(t => t.id === tripId);
    
    if (!trip) {
      throw new Error('Trip not found');
    }
    
    const newWaypoint: Waypoint = {
      ...waypoint,
      id: `waypoint_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`
    };
    
    // Add waypoint to trip
    const updatedTrip = {
      ...trip,
      waypoints: [...(trip.waypoints || []), newWaypoint],
      updatedAt: new Date()
    };
    
    localStorageService.saveTrip(updatedTrip);
    return Promise.resolve(newWaypoint);
  }

  async updateWaypoint(tripId: string, waypointId: string, updates: Partial<Waypoint>): Promise<Waypoint> {
    const trips = localStorageService.getTrips();
    const trip = trips.find(t => t.id === tripId);
    
    if (!trip) {
      throw new Error('Trip not found');
    }
    
    const waypointIndex = trip.waypoints?.findIndex(w => w.id === waypointId) ?? -1;
    if (waypointIndex === -1) {
      throw new Error('Waypoint not found');
    }
    
    const updatedWaypoint: Waypoint = {
      ...trip.waypoints![waypointIndex],
      ...updates,
      id: waypointId
    };
    
    const updatedTrip = {
      ...trip,
      waypoints: [
        ...trip.waypoints!.slice(0, waypointIndex),
        updatedWaypoint,
        ...trip.waypoints!.slice(waypointIndex + 1)
      ],
      updatedAt: new Date()
    };
    
    localStorageService.saveTrip(updatedTrip);
    return Promise.resolve(updatedWaypoint);
  }

  async removeWaypoint(tripId: string, waypointId: string): Promise<void> {
    const trips = localStorageService.getTrips();
    const trip = trips.find(t => t.id === tripId);
    
    if (!trip) {
      throw new Error('Trip not found');
    }
    
    const updatedTrip = {
      ...trip,
      waypoints: trip.waypoints?.filter(w => w.id !== waypointId) || [],
      updatedAt: new Date()
    };
    
    localStorageService.saveTrip(updatedTrip);
    return Promise.resolve();
  }

  async reorderWaypoints(tripId: string, waypointIds: string[]): Promise<void> {
    const trips = localStorageService.getTrips();
    const trip = trips.find(t => t.id === tripId);
    
    if (!trip) {
      throw new Error('Trip not found');
    }
    
    // Create a map of waypoints by ID
    const waypointMap = new Map<string, Waypoint>();
    trip.waypoints?.forEach(w => waypointMap.set(w.id, w));
    
    // Reorder waypoints based on the provided IDs
    const reorderedWaypoints = waypointIds
      .map(id => waypointMap.get(id))
      .filter((w): w is Waypoint => w !== undefined);
    
    const updatedTrip = {
      ...trip,
      waypoints: reorderedWaypoints,
      updatedAt: new Date()
    };
    
    localStorageService.saveTrip(updatedTrip);
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