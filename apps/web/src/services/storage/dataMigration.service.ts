import { localStorageService, LocalStorageData } from './localStorage.service';
import { CloudDataService } from './cloudDataService';
import { store } from '../../store';
import { addNotification } from '../../store/slices/uiSlice';

export interface MigrationResult {
  success: boolean;
  migratedItems: {
    trips: number;
    collections: number;
    places: number;
  };
  errors: string[];
}

class DataMigrationService {
  async migrateLocalDataToCloud(): Promise<MigrationResult> {
    const result: MigrationResult = {
      success: false,
      migratedItems: {
        trips: 0,
        collections: 0,
        places: 0,
      },
      errors: [],
    };

    try {
      // Get all local data
      const localData = localStorageService.getAllData();
      
      // Create cloud service instance
      const cloudService = new CloudDataService();
      
      // Show migration started notification
      store.dispatch(addNotification({
        type: 'info',
        message: 'Migrating your local data to the cloud...',
      }));

      // Migrate collections
      if (localData.collections.length > 0) {
        for (const collection of localData.collections) {
          try {
            await cloudService.saveCollection({
              name: collection.name,
              description: collection.description,
              privacy: collection.privacy,
              user_id: collection.user_id,
              locations: collection.locations,
              created_at: collection.created_at,
              updated_at: collection.updated_at,
            });
            result.migratedItems.collections++;
          } catch (error) {
            result.errors.push(`Failed to migrate collection "${collection.name}": ${error}`);
          }
        }
      }

      // Migrate places
      if (localData.places.length > 0) {
        for (const place of localData.places) {
          try {
            await cloudService.savePlace({
              name: place.name,
              description: place.description,
              location: place.location,
              category: place.category,
              tags: place.tags,
              is_private: place.is_private,
              created_at: place.created_at,
              updated_at: place.updated_at,
            });
            result.migratedItems.places++;
          } catch (error) {
            result.errors.push(`Failed to migrate place "${place.name}": ${error}`);
          }
        }
      }

      // Migrate trips
      if (localData.trips.length > 0) {
        for (const trip of localData.trips) {
          try {
            await cloudService.saveTrip({
              name: trip.name,
              description: trip.description,
              start_date: trip.start_date,
              end_date: trip.end_date,
              visibility: trip.visibility,
              status: trip.status,
              tags: trip.tags,
              settings: trip.settings,
              waypoints: trip.waypoints,
              collaborators: trip.collaborators,
              suggestions: trip.suggestions,
              places: trip.places,
              activities: trip.activities,
              media: trip.media,
              created_at: trip.created_at,
              updated_at: trip.updated_at,
            });
            result.migratedItems.trips++;
          } catch (error) {
            result.errors.push(`Failed to migrate trip "${trip.name}": ${error}`);
          }
        }
      }

      // Mark as successful if at least some items were migrated
      result.success = 
        result.migratedItems.collections > 0 ||
        result.migratedItems.places > 0 ||
        result.migratedItems.trips > 0;

      // Clear local data if migration was successful
      if (result.success && result.errors.length === 0) {
        localStorageService.clearAllData();
        
        store.dispatch(addNotification({
          type: 'success',
          message: `Successfully migrated ${this.getTotalItems(result.migratedItems)} items to the cloud!`,
        }));
      } else if (result.success && result.errors.length > 0) {
        store.dispatch(addNotification({
          type: 'warning',
          message: `Migrated ${this.getTotalItems(result.migratedItems)} items with ${result.errors.length} errors.`,
        }));
      } else {
        store.dispatch(addNotification({
          type: 'error',
          message: 'Failed to migrate local data to the cloud.',
        }));
      }

      return result;
    } catch (error) {
      console.error('Data migration failed:', error);
      result.errors.push(`Migration failed: ${error}`);
      
      store.dispatch(addNotification({
        type: 'error',
        message: 'An error occurred during data migration.',
      }));
      
      return result;
    }
  }

  private getTotalItems(items: MigrationResult['migratedItems']): number {
    return items.trips + items.collections + items.places;
  }

  async checkForLocalData(): Promise<boolean> {
    const data = localStorageService.getAllData();
    return (
      data.collections.length > 0 ||
      data.places.length > 0 ||
      data.trips.length > 0
    );
  }

  async promptForMigration(): Promise<boolean> {
    // This could show a modal or confirmation dialog
    // For now, we'll return true to auto-migrate
    const hasLocalData = await this.checkForLocalData();
    
    if (hasLocalData) {
      // In a real implementation, you'd show a dialog here
      return confirm(
        'You have local data saved. Would you like to sync it to your account?\n\n' +
        'This will move your saved locations, collections, and trips to the cloud so you can access them from any device.'
      );
    }
    
    return false;
  }
}

export const dataMigrationService = new DataMigrationService();