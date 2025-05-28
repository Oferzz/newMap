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

export interface Place {
  id: string;
  name: string;
  address: string;
  category: string | string[];
  location: {
    coordinates: [number, number];
  };
  created_by?: string;
  description?: string;
  rating?: number;
  photos?: string[];
  opening_hours?: {
    [key: string]: string;
  };
  average_rating?: number;
  rating_count?: number;
  street_address?: string;
  city?: string;
  state?: string;
  country?: string;
}

export interface Trip {
  id: string;
  name: string;
  title?: string;
  description: string;
  startDate: string;
  endDate: string;
  start_date?: string;
  end_date?: string;
  waypoints: any[];
  owner_id?: string;
  participants?: any[];
  collaborators?: any[];
  visibility?: string;
  privacy?: string;
  created_at?: string;
  cover_image?: string;
}