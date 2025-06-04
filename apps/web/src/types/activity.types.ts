export interface Activity {
  id: string;
  title: string;
  description: string;
  activity_type: string;
  created_by: string;
  privacy: 'public' | 'friends' | 'private';
  route?: Route;
  metadata?: ActivityMetadata;
  like_count: number;
  comment_count: number;
  view_count: number;
  created_at: string;
  updated_at: string;
}

export interface Route {
  type: 'out-and-back' | 'loop' | 'point-to-point';
  waypoints: Waypoint[];
  distance?: number; // in kilometers
  elevation_gain?: number; // in meters
  elevation_loss?: number; // in meters
}

export interface Waypoint {
  lat: number;
  lng: number;
  elevation?: number;
}

export interface ActivityMetadata {
  difficulty: 'easy' | 'moderate' | 'hard' | 'expert';
  duration: number; // in hours
  distance: number; // in kilometers
  elevation_gain: number; // in meters
  terrain: string[];
  water_features: string[];
  gear: string[];
  seasons: string[];
  conditions: string[];
  tags: string[];
}

export interface ShareLink {
  id: string;
  activity_id: string;
  token: string;
  url: string;
  expires_at?: string;
  created_at: string;
  created_by: string;
  view_count: number;
  settings: ShareSettings;
}

export interface ShareSettings {
  allow_comments: boolean;
  allow_downloads: boolean;
  require_password?: boolean;
  password?: string;
  max_views?: number;
}