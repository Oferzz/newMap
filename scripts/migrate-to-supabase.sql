-- Supabase Migration Script
-- Run this in your Supabase SQL Editor to set up the database schema

-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "postgis";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";

-- Create profiles table (extends auth.users)
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

-- Enable RLS on profiles
ALTER TABLE profiles ENABLE ROW LEVEL SECURITY;

-- Profiles policies
CREATE POLICY "Public profiles are viewable by everyone"
  ON profiles FOR SELECT
  USING (privacy_settings->>'profile_visibility' = 'public');

CREATE POLICY "Users can view own profile"
  ON profiles FOR SELECT
  USING (auth.uid() = id);

CREATE POLICY "Users can update own profile"
  ON profiles FOR UPDATE
  USING (auth.uid() = id);

-- Trips table
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
  is_guest_data BOOLEAN DEFAULT false,
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Enable RLS on trips
ALTER TABLE trips ENABLE ROW LEVEL SECURITY;

-- Trips policies
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

-- Places table
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
  is_guest_data BOOLEAN DEFAULT false,
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Enable RLS on places
ALTER TABLE places ENABLE ROW LEVEL SECURITY;

-- Places policies
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

-- Trip collaborators table
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

-- Enable RLS on trip_collaborators
ALTER TABLE trip_collaborators ENABLE ROW LEVEL SECURITY;

-- Trip collaborators policies
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

-- Storage buckets
INSERT INTO storage.buckets (id, name, public)
VALUES 
  ('avatars', 'avatars', true),
  ('trip-covers', 'trip-covers', true),
  ('place-media', 'place-media', true)
ON CONFLICT DO NOTHING;

-- Storage policies for avatars
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

-- Functions and triggers

-- Updated timestamp trigger
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = NOW();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Apply triggers to tables
CREATE TRIGGER update_profiles_updated_at BEFORE UPDATE ON profiles
  FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_trips_updated_at BEFORE UPDATE ON trips
  FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_places_updated_at BEFORE UPDATE ON places
  FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_trip_collaborators_updated_at BEFORE UPDATE ON trip_collaborators
  FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Profile creation trigger
CREATE OR REPLACE FUNCTION handle_new_user()
RETURNS TRIGGER AS $$
BEGIN
  INSERT INTO public.profiles (id, username, display_name)
  VALUES (
    NEW.id,
    COALESCE(NEW.raw_user_meta_data->>'username', split_part(NEW.email, '@', 1)),
    COALESCE(NEW.raw_user_meta_data->>'display_name', split_part(NEW.email, '@', 1))
  );
  RETURN NEW;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

CREATE TRIGGER on_auth_user_created
  AFTER INSERT ON auth.users
  FOR EACH ROW EXECUTE FUNCTION handle_new_user();

-- Indexes for performance
CREATE INDEX idx_profiles_username ON profiles(username);
CREATE INDEX idx_profiles_roles ON profiles USING GIN(roles);

CREATE INDEX idx_trips_owner_id ON trips(owner_id);
CREATE INDEX idx_trips_visibility ON trips(visibility);
CREATE INDEX idx_trips_dates ON trips(start_date, end_date);
CREATE INDEX idx_trips_tags ON trips USING GIN(tags);
CREATE INDEX idx_trips_guest_data ON trips(owner_id, is_guest_data);

CREATE INDEX idx_places_location ON places USING GIST(location);
CREATE INDEX idx_places_trip_id ON places(trip_id);
CREATE INDEX idx_places_owner_id ON places(owner_id);
CREATE INDEX idx_places_name_trgm ON places USING GIN(name gin_trgm_ops);
CREATE INDEX idx_places_tags ON places USING GIN(tags);
CREATE INDEX idx_places_guest_data ON places(owner_id, is_guest_data);

CREATE INDEX idx_trip_collaborators_trip_id ON trip_collaborators(trip_id);
CREATE INDEX idx_trip_collaborators_user_id ON trip_collaborators(user_id);

-- Function to clean up old guest data
CREATE OR REPLACE FUNCTION cleanup_old_guest_data()
RETURNS void AS $$
BEGIN
  -- Delete guest data older than 30 days
  DELETE FROM trips 
  WHERE is_guest_data = true 
    AND created_at < NOW() - INTERVAL '30 days';
    
  DELETE FROM places 
  WHERE is_guest_data = true 
    AND created_at < NOW() - INTERVAL '30 days';
    
  -- Delete anonymous users with no data (be careful with this)
  DELETE FROM auth.users
  WHERE email IS NULL
    AND created_at < NOW() - INTERVAL '30 days'
    AND id NOT IN (SELECT DISTINCT owner_id FROM trips WHERE owner_id IS NOT NULL)
    AND id NOT IN (SELECT DISTINCT owner_id FROM places WHERE owner_id IS NOT NULL);
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Function to mark user data as permanent after email confirmation
CREATE OR REPLACE FUNCTION mark_user_data_permanent()
RETURNS TRIGGER AS $$
BEGIN
  -- When user email is confirmed, mark their data as permanent
  IF NEW.email_confirmed_at IS NOT NULL AND OLD.email_confirmed_at IS NULL THEN
    UPDATE trips SET is_guest_data = false WHERE owner_id = NEW.id;
    UPDATE places SET is_guest_data = false WHERE owner_id = NEW.id;
  END IF;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER on_user_email_confirmed
  AFTER UPDATE ON auth.users
  FOR EACH ROW
  WHEN (NEW.email_confirmed_at IS NOT NULL AND OLD.email_confirmed_at IS NULL)
  EXECUTE FUNCTION mark_user_data_permanent();