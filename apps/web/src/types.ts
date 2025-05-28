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
  category: string;
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
}

export interface Trip {
  id: string;
  name: string;
  description: string;
  startDate: string;
  endDate: string;
  waypoints: any[];
  owner_id?: string;
  participants?: any[];
  visibility?: string;
}