# Supabase Database Migration Guide

## Overview
This document details the database schema migration from the current PostgreSQL setup to Supabase, including RLS policies, indexes, and data migration strategies.

## Current Schema Analysis

### Key PostgreSQL Features in Use
- UUID generation with uuid-ossp
- PostGIS for geospatial data
- Arrays and JSONB columns
- Full-text search with pg_trgm
- Complex foreign key relationships
- Custom indexes for performance

## Supabase Schema Design

### 1. Enable Required Extensions
```sql
-- Run in Supabase SQL Editor
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "postgis";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";
```

### 2. Core Tables Migration

#### 2.1 Profiles Table (extends auth.users)
```sql
-- Profiles table to extend Supabase auth.users
CREATE TABLE profiles (
  id UUID REFERENCES auth.users(id) ON DELETE CASCADE PRIMARY KEY,
  username TEXT UNIQUE NOT NULL,
  display_name TEXT,
  bio TEXT,
  avatar_url TEXT,
  location TEXT,
  website TEXT,
  roles TEXT[] DEFAULT '{user}'::TEXT[],
  privacy_settings JSONB DEFAULT '{
    "profile_visibility": "public",
    "location_sharing": "friends",
    "activity_visibility": "public"
  }'::JSONB,
  notification_preferences JSONB DEFAULT '{
    "email_notifications": true,
    "push_notifications": true,
    "trip_invites": true,
    "trip_updates": true,
    "new_suggestions": true
  }'::JSONB,
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW(),
  last_active TIMESTAMPTZ DEFAULT NOW()
);

-- RLS Policies for profiles
ALTER TABLE profiles ENABLE ROW LEVEL SECURITY;

CREATE POLICY "Public profiles are viewable by everyone"
  ON profiles FOR SELECT
  USING (privacy_settings->>'profile_visibility' = 'public');

CREATE POLICY "Users can update own profile"
  ON profiles FOR UPDATE
  USING (auth.uid() = id);

-- Indexes
CREATE INDEX idx_profiles_username ON profiles(username);
CREATE INDEX idx_profiles_roles ON profiles USING GIN(roles);
```

#### 2.2 Trips Table with RLS
```sql
CREATE TABLE trips (
  id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
  owner_id UUID REFERENCES profiles(id) ON DELETE CASCADE NOT NULL,
  title VARCHAR(255) NOT NULL,
  description TEXT,
  start_date DATE,
  end_date DATE,
  status VARCHAR(50) DEFAULT 'planning',
  visibility VARCHAR(50) DEFAULT 'private',
  settings JSONB DEFAULT '{}'::JSONB,
  tags TEXT[] DEFAULT '{}'::TEXT[],
  budget DECIMAL(10, 2),
  currency VARCHAR(3) DEFAULT 'USD',
  cover_image_url TEXT,
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Enable RLS
ALTER TABLE trips ENABLE ROW LEVEL SECURITY;

-- RLS Policies
CREATE POLICY "Users can view their own trips"
  ON trips FOR SELECT
  USING (auth.uid() = owner_id);

CREATE POLICY "Users can view public trips"
  ON trips FOR SELECT
  USING (visibility = 'public');

CREATE POLICY "Users can view shared trips"
  ON trips FOR SELECT
  USING (
    EXISTS (
      SELECT 1 FROM trip_collaborators
      WHERE trip_collaborators.trip_id = trips.id
      AND trip_collaborators.user_id = auth.uid()
    )
  );

CREATE POLICY "Users can create trips"
  ON trips FOR INSERT
  WITH CHECK (auth.uid() = owner_id);

CREATE POLICY "Owners can update their trips"
  ON trips FOR UPDATE
  USING (auth.uid() = owner_id);

CREATE POLICY "Owners can delete their trips"
  ON trips FOR DELETE
  USING (auth.uid() = owner_id);

-- Indexes
CREATE INDEX idx_trips_owner_id ON trips(owner_id);
CREATE INDEX idx_trips_visibility ON trips(visibility);
CREATE INDEX idx_trips_dates ON trips(start_date, end_date);
CREATE INDEX idx_trips_tags ON trips USING GIN(tags);
```

#### 2.3 Places Table with PostGIS
```sql
CREATE TABLE places (
  id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
  trip_id UUID REFERENCES trips(id) ON DELETE CASCADE,
  owner_id UUID REFERENCES profiles(id) ON DELETE CASCADE NOT NULL,
  name VARCHAR(255) NOT NULL,
  description TEXT,
  address TEXT,
  location GEOGRAPHY(POINT, 4326) NOT NULL,
  place_type VARCHAR(50),
  tags TEXT[] DEFAULT '{}'::TEXT[],
  metadata JSONB DEFAULT '{}'::JSONB,
  media_urls TEXT[] DEFAULT '{}'::TEXT[],
  visit_date DATE,
  rating INTEGER CHECK (rating >= 1 AND rating <= 5),
  notes TEXT,
  visibility VARCHAR(50) DEFAULT 'private',
  is_public BOOLEAN DEFAULT false,
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Enable RLS
ALTER TABLE places ENABLE ROW LEVEL SECURITY;

-- RLS Policies
CREATE POLICY "Users can view their own places"
  ON places FOR SELECT
  USING (auth.uid() = owner_id);

CREATE POLICY "Users can view public places"
  ON places FOR SELECT
  USING (is_public = true OR visibility = 'public');

CREATE POLICY "Users can view places in shared trips"
  ON places FOR SELECT
  USING (
    trip_id IS NOT NULL AND EXISTS (
      SELECT 1 FROM trips
      WHERE trips.id = places.trip_id
      AND (
        trips.owner_id = auth.uid() OR
        trips.visibility = 'public' OR
        EXISTS (
          SELECT 1 FROM trip_collaborators
          WHERE trip_collaborators.trip_id = trips.id
          AND trip_collaborators.user_id = auth.uid()
        )
      )
    )
  );

CREATE POLICY "Users can create places"
  ON places FOR INSERT
  WITH CHECK (auth.uid() = owner_id);

CREATE POLICY "Owners can update their places"
  ON places FOR UPDATE
  USING (auth.uid() = owner_id);

CREATE POLICY "Owners can delete their places"
  ON places FOR DELETE
  USING (auth.uid() = owner_id);

-- Spatial and text indexes
CREATE INDEX idx_places_location ON places USING GIST(location);
CREATE INDEX idx_places_trip_id ON places(trip_id);
CREATE INDEX idx_places_owner_id ON places(owner_id);
CREATE INDEX idx_places_name_trgm ON places USING GIN(name gin_trgm_ops);
CREATE INDEX idx_places_tags ON places USING GIN(tags);
```

#### 2.4 Trip Collaborators with Permissions
```sql
CREATE TABLE trip_collaborators (
  id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
  trip_id UUID REFERENCES trips(id) ON DELETE CASCADE NOT NULL,
  user_id UUID REFERENCES profiles(id) ON DELETE CASCADE NOT NULL,
  role VARCHAR(50) DEFAULT 'viewer',
  permissions JSONB DEFAULT '{
    "can_edit": false,
    "can_delete": false,
    "can_invite": false,
    "can_manage_collaborators": false
  }'::JSONB,
  invited_by UUID REFERENCES profiles(id),
  invited_at TIMESTAMPTZ DEFAULT NOW(),
  accepted_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW(),
  UNIQUE(trip_id, user_id)
);

-- Enable RLS
ALTER TABLE trip_collaborators ENABLE ROW LEVEL SECURITY;

-- RLS Policies
CREATE POLICY "Users can view collaborations they're part of"
  ON trip_collaborators FOR SELECT
  USING (
    auth.uid() = user_id OR
    EXISTS (
      SELECT 1 FROM trips
      WHERE trips.id = trip_collaborators.trip_id
      AND trips.owner_id = auth.uid()
    )
  );

CREATE POLICY "Trip owners can manage collaborators"
  ON trip_collaborators FOR ALL
  USING (
    EXISTS (
      SELECT 1 FROM trips
      WHERE trips.id = trip_collaborators.trip_id
      AND trips.owner_id = auth.uid()
    )
  );

CREATE POLICY "Collaborators with invite permission can add others"
  ON trip_collaborators FOR INSERT
  WITH CHECK (
    EXISTS (
      SELECT 1 FROM trip_collaborators tc
      WHERE tc.trip_id = trip_collaborators.trip_id
      AND tc.user_id = auth.uid()
      AND (tc.permissions->>'can_invite')::boolean = true
    )
  );
```

### 3. Storage Configuration

#### 3.1 Media Storage Buckets
```sql
-- Create storage buckets (run in Supabase dashboard)
INSERT INTO storage.buckets (id, name, public)
VALUES 
  ('avatars', 'avatars', true),
  ('trip-covers', 'trip-covers', true),
  ('place-media', 'place-media', true);

-- Storage policies
CREATE POLICY "Avatar images are publicly accessible"
  ON storage.objects FOR SELECT
  USING (bucket_id = 'avatars');

CREATE POLICY "Users can upload their own avatar"
  ON storage.objects FOR INSERT
  WITH CHECK (
    bucket_id = 'avatars' AND
    (auth.uid())::text = (storage.foldername(name))[1]
  );

CREATE POLICY "Users can update their own avatar"
  ON storage.objects FOR UPDATE
  USING (
    bucket_id = 'avatars' AND
    (auth.uid())::text = (storage.foldername(name))[1]
  );

CREATE POLICY "Users can delete their own avatar"
  ON storage.objects FOR DELETE
  USING (
    bucket_id = 'avatars' AND
    (auth.uid())::text = (storage.foldername(name))[1]
  );
```

### 4. Functions and Triggers

#### 4.1 Updated Timestamp Trigger
```sql
-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = NOW();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Apply to all tables
CREATE TRIGGER update_profiles_updated_at BEFORE UPDATE ON profiles
  FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_trips_updated_at BEFORE UPDATE ON trips
  FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_places_updated_at BEFORE UPDATE ON places
  FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
```

#### 4.2 Profile Creation Trigger
```sql
-- Automatically create profile on user signup
CREATE OR REPLACE FUNCTION handle_new_user()
RETURNS TRIGGER AS $$
BEGIN
  INSERT INTO public.profiles (id, username, display_name)
  VALUES (
    NEW.id,
    COALESCE(NEW.raw_user_meta_data->>'username', NEW.email),
    COALESCE(NEW.raw_user_meta_data->>'display_name', NEW.email)
  );
  RETURN NEW;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

CREATE TRIGGER on_auth_user_created
  AFTER INSERT ON auth.users
  FOR EACH ROW EXECUTE FUNCTION handle_new_user();
```

### 5. Search and Query Optimizations

#### 5.1 Full-Text Search Configuration
```sql
-- Create text search configuration
CREATE TEXT SEARCH CONFIGURATION trip_search (COPY = english);

-- Add text search columns
ALTER TABLE trips ADD COLUMN search_vector tsvector
  GENERATED ALWAYS AS (
    setweight(to_tsvector('trip_search', coalesce(title, '')), 'A') ||
    setweight(to_tsvector('trip_search', coalesce(description, '')), 'B') ||
    setweight(to_tsvector('trip_search', coalesce(array_to_string(tags, ' '), '')), 'C')
  ) STORED;

CREATE INDEX idx_trips_search ON trips USING GIN(search_vector);

-- Search function
CREATE OR REPLACE FUNCTION search_trips(search_query text)
RETURNS SETOF trips AS $$
BEGIN
  RETURN QUERY
  SELECT *
  FROM trips
  WHERE search_vector @@ plainto_tsquery('trip_search', search_query)
  ORDER BY ts_rank(search_vector, plainto_tsquery('trip_search', search_query)) DESC;
END;
$$ LANGUAGE plpgsql;
```

#### 5.2 Geospatial Query Functions
```sql
-- Find places near a point
CREATE OR REPLACE FUNCTION find_places_nearby(
  lat double precision,
  lng double precision,
  radius_meters integer DEFAULT 5000
)
RETURNS SETOF places AS $$
BEGIN
  RETURN QUERY
  SELECT *
  FROM places
  WHERE ST_DWithin(
    location::geography,
    ST_SetSRID(ST_MakePoint(lng, lat), 4326)::geography,
    radius_meters
  )
  ORDER BY location <-> ST_SetSRID(ST_MakePoint(lng, lat), 4326)::geography;
END;
$$ LANGUAGE plpgsql;

-- RLS-aware version
CREATE OR REPLACE FUNCTION find_public_places_nearby(
  lat double precision,
  lng double precision,
  radius_meters integer DEFAULT 5000
)
RETURNS TABLE (
  id UUID,
  name VARCHAR(255),
  description TEXT,
  distance_meters DOUBLE PRECISION
) AS $$
BEGIN
  RETURN QUERY
  SELECT 
    p.id,
    p.name,
    p.description,
    ST_Distance(
      p.location::geography,
      ST_SetSRID(ST_MakePoint(lng, lat), 4326)::geography
    ) as distance_meters
  FROM places p
  WHERE p.is_public = true
    AND ST_DWithin(
      p.location::geography,
      ST_SetSRID(ST_MakePoint(lng, lat), 4326)::geography,
      radius_meters
    )
  ORDER BY distance_meters;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;
```

### 6. Data Migration Scripts

#### 6.1 Batch Migration Script
```javascript
// scripts/migrate-data.js
const { Client: PGClient } = require('pg');
const { createClient } = require('@supabase/supabase-js');

const OLD_DB_URL = process.env.OLD_DATABASE_URL;
const SUPABASE_URL = process.env.SUPABASE_URL;
const SUPABASE_SERVICE_KEY = process.env.SUPABASE_SERVICE_KEY;

const pgClient = new PGClient({ connectionString: OLD_DB_URL });
const supabase = createClient(SUPABASE_URL, SUPABASE_SERVICE_KEY);

async function migrateTrips() {
  const { rows: trips } = await pgClient.query(`
    SELECT * FROM trips 
    ORDER BY created_at
    LIMIT 1000
  `);
  
  const batchSize = 100;
  for (let i = 0; i < trips.length; i += batchSize) {
    const batch = trips.slice(i, i + batchSize);
    
    const { error } = await supabase
      .from('trips')
      .insert(batch.map(trip => ({
        id: trip.id,
        owner_id: trip.owner_id,
        title: trip.title,
        description: trip.description,
        start_date: trip.start_date,
        end_date: trip.end_date,
        status: trip.status,
        visibility: trip.visibility,
        settings: trip.settings,
        tags: trip.tags,
        budget: trip.budget,
        currency: trip.currency,
        cover_image_url: trip.cover_image_url,
        created_at: trip.created_at,
        updated_at: trip.updated_at
      })));
    
    if (error) {
      console.error(`Error migrating trips batch ${i}:`, error);
    } else {
      console.log(`Migrated trips ${i} to ${i + batch.length}`);
    }
  }
}

async function migratePlaces() {
  const { rows: places } = await pgClient.query(`
    SELECT 
      *,
      ST_X(location::geometry) as lng,
      ST_Y(location::geometry) as lat
    FROM places 
    ORDER BY created_at
  `);
  
  for (const place of places) {
    const { error } = await supabase.rpc('insert_place_with_location', {
      p_id: place.id,
      p_trip_id: place.trip_id,
      p_owner_id: place.owner_id,
      p_name: place.name,
      p_description: place.description,
      p_address: place.address,
      p_lat: place.lat,
      p_lng: place.lng,
      p_place_type: place.place_type,
      p_tags: place.tags,
      p_metadata: place.metadata,
      p_media_urls: place.media_urls,
      p_visit_date: place.visit_date,
      p_rating: place.rating,
      p_notes: place.notes,
      p_visibility: place.visibility,
      p_is_public: place.is_public,
      p_created_at: place.created_at,
      p_updated_at: place.updated_at
    });
    
    if (error) {
      console.error(`Error migrating place ${place.id}:`, error);
    }
  }
}
```

### 7. Performance Monitoring

#### 7.1 Query Performance Views
```sql
-- Create view for monitoring slow queries
CREATE VIEW slow_queries AS
SELECT 
  query,
  calls,
  total_time,
  mean_time,
  min_time,
  max_time
FROM pg_stat_statements
WHERE mean_time > 100
ORDER BY mean_time DESC;

-- Monitor table sizes
CREATE VIEW table_sizes AS
SELECT
  schemaname,
  tablename,
  pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) AS size,
  pg_total_relation_size(schemaname||'.'||tablename) AS size_bytes
FROM pg_tables
WHERE schemaname = 'public'
ORDER BY size_bytes DESC;
```

### 8. Backup and Recovery

#### 8.1 Backup Strategy
```bash
#!/bin/bash
# backup-supabase.sh

# Daily backup script
SUPABASE_DB_URL="postgresql://postgres:[password]@[host]:[port]/postgres"
BACKUP_DIR="/backups/supabase"
DATE=$(date +%Y%m%d_%H%M%S)

# Create backup
pg_dump $SUPABASE_DB_URL \
  --no-owner \
  --no-privileges \
  --exclude-schema=auth \
  --exclude-schema=storage \
  -f "$BACKUP_DIR/backup_$DATE.sql"

# Compress
gzip "$BACKUP_DIR/backup_$DATE.sql"

# Delete backups older than 30 days
find $BACKUP_DIR -name "backup_*.sql.gz" -mtime +30 -delete
```

### 9. Testing RLS Policies

```sql
-- Test RLS policies
-- Run as different users to verify access

-- Test as user1
SET LOCAL role = 'authenticated';
SET LOCAL request.jwt.claims = '{"sub": "user1-uuid"}';

-- Should only see own trips and public trips
SELECT * FROM trips;

-- Test as user2
SET LOCAL request.jwt.claims = '{"sub": "user2-uuid"}';
SELECT * FROM trips;

-- Reset
RESET role;
```