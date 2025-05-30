import { Collection, Place, Trip } from '../../types';

// Storage keys
export const STORAGE_KEYS = {
  TRIPS: 'newmap_trips',
  COLLECTIONS: 'newmap_collections',
  PLACES: 'newmap_places',
  ROUTES: 'newmap_routes',
  PREFERENCES: 'newmap_preferences',
  TEMPORARY_MARKERS: 'newmap_temp_markers'
} as const;

export interface LocalStorageData {
  trips: Trip[];
  collections: Collection[];
  places: Place[];
  routes: any[]; // TODO: Define Route type
  temporaryMarkers: Array<{ id: string; coordinates: [number, number] }>;
}

class LocalStorageService {
  private isAvailable(): boolean {
    try {
      const testKey = '__localStorage_test__';
      localStorage.setItem(testKey, 'test');
      localStorage.removeItem(testKey);
      return true;
    } catch {
      return false;
    }
  }

  private getItem<T>(key: string, defaultValue: T): T {
    if (!this.isAvailable()) return defaultValue;
    
    try {
      const item = localStorage.getItem(key);
      return item ? JSON.parse(item) : defaultValue;
    } catch (error) {
      console.error(`Error reading from localStorage (${key}):`, error);
      return defaultValue;
    }
  }

  private setItem<T>(key: string, value: T): void {
    if (!this.isAvailable()) return;
    
    try {
      localStorage.setItem(key, JSON.stringify(value));
    } catch (error) {
      console.error(`Error writing to localStorage (${key}):`, error);
      // Handle storage quota exceeded
      if (error instanceof DOMException && error.code === 22) {
        this.handleStorageQuotaExceeded();
      }
    }
  }

  private handleStorageQuotaExceeded(): void {
    // Remove oldest items to make space
    const trips = this.getTrips();
    if (trips.length > 10) {
      // Keep only the 10 most recent trips
      const recentTrips = trips
        .sort((a, b) => {
          const dateA = a.updated_at || a.created_at || '';
          const dateB = b.updated_at || b.created_at || '';
          return new Date(dateB).getTime() - new Date(dateA).getTime();
        })
        .slice(0, 10);
      this.setItem(STORAGE_KEYS.TRIPS, recentTrips);
    }
  }

  // Collections
  getCollections(): Collection[] {
    return this.getItem(STORAGE_KEYS.COLLECTIONS, []);
  }

  saveCollection(collection: Collection): void {
    const collections = this.getCollections();
    const existingIndex = collections.findIndex(c => c.id === collection.id);
    
    if (existingIndex >= 0) {
      collections[existingIndex] = collection;
    } else {
      collections.push(collection);
    }
    
    this.setItem(STORAGE_KEYS.COLLECTIONS, collections);
  }

  deleteCollection(collectionId: string): void {
    const collections = this.getCollections().filter(c => c.id !== collectionId);
    this.setItem(STORAGE_KEYS.COLLECTIONS, collections);
  }

  addLocationToCollection(collectionId: string, location: { latitude: number; longitude: number; name?: string }): void {
    const collections = this.getCollections();
    const collection = collections.find(c => c.id === collectionId);
    
    if (collection) {
      const newLocation = {
        id: `loc_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`,
        collection_id: collectionId,
        name: location.name || null,
        latitude: location.latitude,
        longitude: location.longitude,
        added_at: new Date().toISOString()
      };
      
      collection.locations = [...(collection.locations || []), newLocation];
      this.saveCollection(collection);
    }
  }

  // Places
  getPlaces(): Place[] {
    return this.getItem(STORAGE_KEYS.PLACES, []);
  }

  savePlace(place: Place): void {
    const places = this.getPlaces();
    const existingIndex = places.findIndex(p => p.id === place.id);
    
    if (existingIndex >= 0) {
      places[existingIndex] = place;
    } else {
      places.push(place);
    }
    
    this.setItem(STORAGE_KEYS.PLACES, places);
  }

  deletePlace(placeId: string): void {
    const places = this.getPlaces().filter(p => p.id !== placeId);
    this.setItem(STORAGE_KEYS.PLACES, places);
  }

  // Trips
  getTrips(): Trip[] {
    return this.getItem(STORAGE_KEYS.TRIPS, []);
  }

  saveTrip(trip: Trip): void {
    const trips = this.getTrips();
    const existingIndex = trips.findIndex(t => t.id === trip.id);
    
    if (existingIndex >= 0) {
      trips[existingIndex] = trip;
    } else {
      trips.push(trip);
    }
    
    this.setItem(STORAGE_KEYS.TRIPS, trips);
  }

  deleteTrip(tripId: string): void {
    const trips = this.getTrips().filter(t => t.id !== tripId);
    this.setItem(STORAGE_KEYS.TRIPS, trips);
  }

  // Temporary Markers
  getTemporaryMarkers(): Array<{ id: string; coordinates: [number, number] }> {
    return this.getItem(STORAGE_KEYS.TEMPORARY_MARKERS, []);
  }

  saveTemporaryMarker(marker: { id: string; coordinates: [number, number] }): void {
    const markers = this.getTemporaryMarkers();
    markers.push(marker);
    this.setItem(STORAGE_KEYS.TEMPORARY_MARKERS, markers);
  }

  removeTemporaryMarker(markerId: string): void {
    const markers = this.getTemporaryMarkers().filter(m => m.id !== markerId);
    this.setItem(STORAGE_KEYS.TEMPORARY_MARKERS, markers);
  }

  clearTemporaryMarkers(): void {
    this.setItem(STORAGE_KEYS.TEMPORARY_MARKERS, []);
  }

  // Routes
  getRoutes(): any[] {
    return this.getItem(STORAGE_KEYS.ROUTES, []);
  }

  saveRoute(route: any): void {
    const routes = this.getRoutes();
    routes.push(route);
    this.setItem(STORAGE_KEYS.ROUTES, routes);
  }

  // Clear all data
  clearAllData(): void {
    Object.values(STORAGE_KEYS).forEach(key => {
      localStorage.removeItem(key);
    });
  }

  // Get all data for migration
  getAllData(): LocalStorageData {
    return {
      trips: this.getTrips(),
      collections: this.getCollections(),
      places: this.getPlaces(),
      routes: this.getRoutes(),
      temporaryMarkers: this.getTemporaryMarkers()
    };
  }

  // Check storage usage
  getStorageInfo(): { used: number; available: boolean } {
    if (!this.isAvailable()) {
      return { used: 0, available: false };
    }

    let totalSize = 0;
    for (const key in localStorage) {
      if (localStorage.hasOwnProperty(key)) {
        totalSize += localStorage[key].length + key.length;
      }
    }

    return {
      used: totalSize,
      available: true
    };
  }
}

export const localStorageService = new LocalStorageService();