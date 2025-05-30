import { describe, it, expect } from 'vitest';
import { normalizePlaceType, normalizeTripType, isPlace, isTrip, getNormalizedProperty } from './typeHelpers';
import { Place as ServicePlace } from '../services/places.service';
import { Trip as ServiceTrip } from '../services/trips.service';

describe('typeHelpers', () => {
  describe('normalizePlaceType', () => {
    it('should convert service Place to normalized Place with address', () => {
      const servicePlace: ServicePlace = {
        id: '1',
        name: 'Test Place',
        description: 'A test place',
        type: 'poi',
        street_address: '123 Main St',
        city: 'Test City',
        state: 'Test State',
        country: 'Test Country',
        postal_code: '12345',
        created_by: 'user123',
        category: ['restaurant'],
        tags: ['food'],
        rating_count: 10,
        average_rating: 4.5,
        amenities: [],
        privacy: 'public',
        status: 'active',
        created_at: '2023-01-01',
        updated_at: '2023-01-02',
      };

      const normalized = normalizePlaceType(servicePlace);

      expect(normalized.address).toBe('123 Main St, Test City, Test State, Test Country');
      expect(normalized.rating).toBe(4.5);
      expect(normalized.photos).toEqual([]);
    });
  });

  describe('normalizeTripType', () => {
    it('should convert service Trip to normalized Trip with legacy properties', () => {
      const serviceTrip: ServiceTrip = {
        id: '1',
        title: 'Test Trip',
        description: 'A test trip',
        owner_id: 'user123',
        cover_image: 'image.jpg',
        privacy: 'public',
        status: 'planning',
        start_date: '2023-06-01',
        end_date: '2023-06-10',
        timezone: 'UTC',
        tags: ['vacation'],
        view_count: 100,
        share_count: 10,
        suggestion_count: 5,
        created_at: '2023-01-01',
        updated_at: '2023-01-02',
      };

      const normalized = normalizeTripType(serviceTrip);

      expect(normalized.name).toBe('Test Trip');
      expect(normalized.startDate).toBe('2023-06-01');
      expect(normalized.endDate).toBe('2023-06-10');
      expect(normalized.privacy).toBe('public');
    });
  });

  describe('isPlace', () => {
    it('should identify Place by location property', () => {
      const place = {
        id: '1',
        name: 'Test',
        description: 'Test',
        location: { type: 'Point', coordinates: [0, 0] },
      };

      expect(isPlace(place as any)).toBe(true);
    });

    it('should identify Place by address property', () => {
      const place = {
        id: '1',
        name: 'Test',
        description: 'Test',
        address: '123 Main St',
      };

      expect(isPlace(place as any)).toBe(true);
    });

    it('should identify Place by street_address property', () => {
      const place = {
        id: '1',
        name: 'Test',
        description: 'Test',
        street_address: '123 Main St',
      };

      expect(isPlace(place as any)).toBe(true);
    });
  });

  describe('isTrip', () => {
    it('should identify Trip by owner_id property', () => {
      const trip = {
        id: '1',
        description: 'Test',
        owner_id: 'user123',
      };

      expect(isTrip(trip as any)).toBe(true);
    });

    it('should identify Trip by ownerID property', () => {
      const trip = {
        id: '1',
        description: 'Test',
        ownerID: 'user123',
      };

      expect(isTrip(trip as any)).toBe(true);
    });

    it('should identify Trip by waypoints property', () => {
      const trip = {
        id: '1',
        description: 'Test',
        waypoints: [],
      };

      expect(isTrip(trip as any)).toBe(true);
    });
  });

  describe('getNormalizedProperty', () => {
    it('should get owner_id from different property names', () => {
      expect(getNormalizedProperty({ owner_id: 'user1' } as any, 'owner_id')).toBe('user1');
      expect(getNormalizedProperty({ ownerID: 'user2' } as any, 'owner_id')).toBe('user2');
      expect(getNormalizedProperty({ created_by: 'user3' } as any, 'owner_id')).toBe('user3');
    });

    it('should get title from different property names', () => {
      expect(getNormalizedProperty({ title: 'Trip 1' } as any, 'title')).toBe('Trip 1');
      expect(getNormalizedProperty({ name: 'Trip 2' } as any, 'title')).toBe('Trip 2');
    });

    it('should extract street address from full address', () => {
      expect(getNormalizedProperty({ street_address: '123 Main St' } as any, 'street_address')).toBe('123 Main St');
      expect(getNormalizedProperty({ address: '456 Oak Ave, Test City, Test State' } as any, 'street_address')).toBe('456 Oak Ave');
    });
  });
});