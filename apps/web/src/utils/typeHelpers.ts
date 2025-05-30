import { Place as ServicePlace } from '../services/places.service';
import { Trip as ServiceTrip } from '../services/trips.service';
import { Place, Trip } from '../types';

/**
 * Convert a service Place to the types.ts Place format with convenience properties
 */
export function normalizePlaceType(place: ServicePlace): Place {
  return {
    ...place,
    // Add convenience address property by combining address fields
    address: [place.street_address, place.city, place.state, place.country]
      .filter(Boolean)
      .join(', '),
    // Map media to photos for backward compatibility
    photos: place.media?.map((m: any) => m.url || m.file_url).filter(Boolean) || [],
    // Use average_rating as rating for backward compatibility
    rating: place.average_rating,
  } as Place;
}

/**
 * Convert a service Trip to the types.ts Trip format with convenience properties
 */
export function normalizeTripType(trip: ServiceTrip): Trip {
  return {
    ...trip,
    // Add convenience properties for backward compatibility
    name: trip.title,
    startDate: trip.start_date,
    endDate: trip.end_date,
    participants: trip.collaborators,
    visibility: trip.privacy,
  } as Trip;
}

/**
 * Type guard to check if an item is a Place
 */
export function isPlace(item: Place | Trip): item is Place {
  return 'location' in item || 'address' in item || 'street_address' in item;
}

/**
 * Type guard to check if an item is a Trip  
 */
export function isTrip(item: Place | Trip): item is Trip {
  return 'owner_id' in item || 'ownerID' in item || 'waypoints' in item;
}

/**
 * Get a normalized property value that works across different type formats
 */
export function getNormalizedProperty<T extends Place | Trip>(
  item: T,
  property: 'owner_id' | 'title' | 'cover_image' | 'start_date' | 'end_date' | 'street_address'
): string | undefined {
  switch (property) {
    case 'owner_id':
      return (item as any).owner_id || (item as any).ownerID || (item as any).created_by;
    case 'title':
      return (item as any).title || (item as any).name;
    case 'cover_image':
      return (item as any).cover_image || (item as any).coverImage;
    case 'start_date':
      return (item as any).start_date || (item as any).startDate;
    case 'end_date':
      return (item as any).end_date || (item as any).endDate;
    case 'street_address':
      if ('street_address' in item) return (item as any).street_address;
      if ('address' in item && typeof (item as any).address === 'string') {
        // Extract street portion from full address  
        return (item as any).address.split(',')[0]?.trim();
      }
      return undefined;
    default:
      return undefined;
  }
}

/**
 * Normalize search results to ensure type compatibility
 */
export function normalizeSearchResults(results: {
  places: ServicePlace[];
  trips: ServiceTrip[];
  users: any[];
}): {
  places: Place[];
  trips: Trip[];
  users: any[];
} {
  return {
    places: results.places.map(normalizePlaceType),
    trips: results.trips.map(normalizeTripType),
    users: results.users,
  };
}