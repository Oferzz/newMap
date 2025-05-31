export interface SearchResult {
  id: string;
  type: 'place' | 'trip' | 'user';
  name: string;
  description?: string;
  coordinates?: {
    lat: number;
    lng: number;
  };
}

export interface SearchResults {
  places: Place[];
  trips: Trip[];
  users: any[];
}

export interface CollectionLocation {
  id: string;
  collection_id: string;
  name: string | null;
  latitude: number;
  longitude: number;
  added_at: string;
}

export interface Collection {
  id: string;
  name: string;
  description: string | null;
  user_id: string;
  privacy: string;
  locations: CollectionLocation[];
  created_at: string;
  updated_at: string;
}

// Union type to support both API and Redux slice formats
export type Place = {
  id: string;
  name: string;
  description: string;
  created_by?: string;
  category: string[] | string;
  average_rating?: number;
  rating_count?: number;
  opening_hours?: any;
  privacy?: string;
  created_at?: string;
  updated_at?: string;
  media?: any[];
  collaborators?: any[];
  photos?: string[];
  rating?: number;
} & (
  // API format
  | {
      type: string;
      parent_id?: string;
      location?: {
        type: string;
        coordinates: [number, number];
      };
      bounds?: any;
      street_address?: string;
      city?: string;
      state?: string;
      country?: string;
      postal_code?: string;
      tags?: string[];
      contact_info?: any;
      amenities?: string[];
      status?: string;
      address?: string;
    }
  // Redux slice format
  | {
      type: 'poi' | 'area' | 'region';
      location?: {
        type: string;
        coordinates: [number, number];
      };
      address: string;
      city: string;
      state: string;
      country: string;
      postalCode: string;
      tags: string[];
      contactInfo?: {
        phone?: string;
        email?: string;
        website?: string;
      };
      images: string[];
      createdBy: string;
      createdAt: Date;
      updatedAt: Date;
      street_address?: string;
      postal_code?: string;
      amenities?: string[];
      status?: string;
    }
);

// Union type to support both API and Redux slice formats  
export type Trip = {
  id: string;
  description: string;
  collaborators?: any[];
  waypoints?: any[];
  created_at?: string;
  updated_at?: string;
  media?: any[];
} & (
  // API format
  | {
      title: string;
      owner_id: string;
      cover_image: string;
      privacy: string;
      status: string;
      start_date?: string;
      end_date?: string;
      timezone: string;
      tags: string[];
      view_count: number;
      share_count: number;
      suggestion_count: number;
      name?: string;
      startDate?: string;
      endDate?: string;
      participants?: any[];
      visibility?: string;
    }
  // Redux slice format
  | {
      title: string;
      startDate: Date;
      endDate: Date;
      coverImage: string;
      status: 'planning' | 'active' | 'completed';
      privacy: 'public' | 'friends' | 'private';
      ownerID: string;
      tags: string[];
      createdAt: Date;
      updatedAt: Date;
      name?: string;
      owner_id?: string;
      cover_image?: string;
      start_date?: string;
      end_date?: string;
      timezone?: string;
      view_count?: number;
      share_count?: number;
      suggestion_count?: number;
    }
);