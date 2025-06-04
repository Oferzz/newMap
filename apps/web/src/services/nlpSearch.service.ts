import { api, ApiResponse } from './api';

// Types for Natural Language Search
export interface NLPSearchResult {
  id: string;
  type: 'activity' | 'place';
  source: {
    [key: string]: any;
  };
  score: number;
}

export interface NLPSearchResponse {
  total: number;
  results: NLPSearchResult[];
  took: number;
  query_understanding?: {
    intent: string;
    filters: Record<string, any>;
    confidence: number;
    explanation: string;
  };
}

export interface NLPSearchParams {
  query: string;
  limit?: number;
  offset?: number;
}

export interface NLPQueryParseResponse {
  intent: string;
  search_text: string;
  filters: Record<string, any>;
  location?: {
    name?: string;
    latitude?: number;
    longitude?: number;
    radius?: number;
  };
  confidence: number;
  keywords: string[];
  explanation: string;
}

export interface SearchSuggestion {
  suggestion: string;
  type: 'activity' | 'place' | 'query';
  count?: number;
}

class NLPSearchService {
  /**
   * Perform natural language search
   */
  async search(params: NLPSearchParams): Promise<ApiResponse<NLPSearchResponse>> {
    const { query, limit = 20, offset = 0 } = params;
    
    const queryParams = new URLSearchParams({
      q: query,
      limit: limit.toString(),
      offset: offset.toString(),
    });

    return api.get<ApiResponse<NLPSearchResponse>>(`/search?${queryParams}`, {
      skipAuth: false, // Allow both authenticated and unauthenticated access
    });
  }

  /**
   * Get search suggestions for autocomplete
   */
  async getSuggestions(prefix: string, limit = 10): Promise<ApiResponse<string[]>> {
    const queryParams = new URLSearchParams({
      prefix,
      limit: limit.toString(),
    });

    return api.get<ApiResponse<string[]>>(`/search/suggestions?${queryParams}`, {
      skipAuth: false,
    });
  }

  /**
   * Parse a natural language query to understand intent and filters
   */
  async parseQuery(query: string): Promise<ApiResponse<NLPQueryParseResponse>> {
    const queryParams = new URLSearchParams({
      q: query,
    });

    return api.get<ApiResponse<NLPQueryParseResponse>>(`/search/parse?${queryParams}`, {
      skipAuth: false,
    });
  }

  /**
   * Transform NLP search results to legacy SearchResult format for compatibility
   */
  transformResultsToLegacyFormat(results: NLPSearchResult[]): Array<{
    id: string;
    type: 'place' | 'trip' | 'user';
    name: string;
    description?: string;
    coordinates?: {
      lat: number;
      lng: number;
    };
  }> {
    return results.map(result => {
      const source = result.source;
      let name = '';
      let type: 'place' | 'trip' | 'user' = 'place';
      let description = '';
      let coordinates: { lat: number; lng: number } | undefined;

      if (result.type === 'activity') {
        // Map activity to trip for legacy compatibility
        type = 'trip';
        name = source.title || source.name || 'Unnamed Activity';
        description = source.description || '';
        
        // Extract coordinates from route if available
        if (source.route_geojson && source.route_geojson.coordinates) {
          const coords = source.route_geojson.coordinates;
          if (coords.length > 0) {
            coordinates = {
              lat: coords[0][1],
              lng: coords[0][0],
            };
          }
        }
      } else if (result.type === 'place') {
        type = 'place';
        name = source.name || 'Unnamed Place';
        description = source.description || '';
        
        // Extract coordinates from location
        if (source.location && source.location.coordinates) {
          coordinates = {
            lat: source.location.coordinates[1],
            lng: source.location.coordinates[0],
          };
        }
      }

      return {
        id: result.id,
        type,
        name,
        description,
        coordinates,
      };
    });
  }
}

export const nlpSearchService = new NLPSearchService();