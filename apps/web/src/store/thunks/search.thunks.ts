import { createAsyncThunk } from '@reduxjs/toolkit';
import { placesService } from '../../services/places.service';
import { tripsService } from '../../services/trips.service';
import { SearchResults } from '../../types';
import { setLoading, setError } from '../slices/searchSlice';
import { setSearchResults, setIsSearching } from '../slices/uiSlice';
import { normalizeSearchResults } from '../../utils/typeHelpers';

interface SearchParams {
  query: string;
  filters?: {
    type: string;
    radius?: number;
    onlyMine?: boolean;
  };
}

export const searchAllThunk = createAsyncThunk(
  'search/searchAll',
  async ({ query, filters }: SearchParams, { dispatch }) => {
    try {
      dispatch(setLoading(true));
      dispatch(setIsSearching(true));
      
      const results = {
        places: [] as any[],
        trips: [] as any[],
        users: [] as any[]
      };

      // Search based on filter type
      if (!filters || filters.type === 'all' || filters.type === 'places') {
        try {
          const placesResponse = await placesService.search({
            q: query,
            radius: filters?.radius,
            limit: 10
          });
          results.places = placesResponse.data || [];
        } catch (error) {
          console.error('Error searching places:', error);
        }
      }

      if (!filters || filters.type === 'all' || filters.type === 'trips') {
        try {
          const tripsResponse = await tripsService.search(query, {
            limit: 10
          });
          results.trips = tripsResponse.data || [];
        } catch (error) {
          console.error('Error searching trips:', error);
        }
      }

      // TODO: Add user search when API is available
      if (!filters || filters.type === 'all' || filters.type === 'users') {
        // Placeholder for user search
        results.users = [];
      }

      const normalizedResults = normalizeSearchResults(results as any);
      dispatch(setSearchResults(normalizedResults as SearchResults));
      dispatch(setIsSearching(true));
      
      return results;
    } catch (error: any) {
      dispatch(setError(error.message || 'Search failed'));
      throw error;
    } finally {
      dispatch(setLoading(false));
    }
  }
);