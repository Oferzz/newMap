-- Extend trips table to support activities with advanced metadata
ALTER TABLE trips ADD COLUMN IF NOT EXISTS activity_type VARCHAR(50) DEFAULT 'general';
ALTER TABLE trips ADD COLUMN IF NOT EXISTS difficulty_level VARCHAR(20);
ALTER TABLE trips ADD COLUMN IF NOT EXISTS duration_hours DECIMAL(5,2);
ALTER TABLE trips ADD COLUMN IF NOT EXISTS distance_km DECIMAL(8,2);
ALTER TABLE trips ADD COLUMN IF NOT EXISTS elevation_gain_m INTEGER;
ALTER TABLE trips ADD COLUMN IF NOT EXISTS max_elevation_m INTEGER;
ALTER TABLE trips ADD COLUMN IF NOT EXISTS route_type VARCHAR(50); -- 'out_and_back', 'loop', 'point_to_point', 'area'
ALTER TABLE trips ADD COLUMN IF NOT EXISTS route_geojson JSONB;
ALTER TABLE trips ADD COLUMN IF NOT EXISTS water_features TEXT[];
ALTER TABLE trips ADD COLUMN IF NOT EXISTS terrain_types TEXT[];
ALTER TABLE trips ADD COLUMN IF NOT EXISTS essential_gear TEXT[];
ALTER TABLE trips ADD COLUMN IF NOT EXISTS best_seasons TEXT[];
ALTER TABLE trips ADD COLUMN IF NOT EXISTS trail_conditions TEXT;
ALTER TABLE trips ADD COLUMN IF NOT EXISTS accessibility_notes TEXT;
ALTER TABLE trips ADD COLUMN IF NOT EXISTS parking_info JSONB;
ALTER TABLE trips ADD COLUMN IF NOT EXISTS permits_required TEXT[];
ALTER TABLE trips ADD COLUMN IF NOT EXISTS hazards TEXT[];
ALTER TABLE trips ADD COLUMN IF NOT EXISTS emergency_contacts JSONB;
ALTER TABLE trips ADD COLUMN IF NOT EXISTS visibility VARCHAR(20) DEFAULT 'private'; -- 'public', 'private'
ALTER TABLE trips ADD COLUMN IF NOT EXISTS shared_with UUID[];
ALTER TABLE trips ADD COLUMN IF NOT EXISTS completion_count INTEGER DEFAULT 0;
ALTER TABLE trips ADD COLUMN IF NOT EXISTS average_rating DECIMAL(3,2);
ALTER TABLE trips ADD COLUMN IF NOT EXISTS rating_count INTEGER DEFAULT 0;
ALTER TABLE trips ADD COLUMN IF NOT EXISTS featured BOOLEAN DEFAULT false;
ALTER TABLE trips ADD COLUMN IF NOT EXISTS verified BOOLEAN DEFAULT false;

-- Create activity-specific indexes
CREATE INDEX IF NOT EXISTS idx_trips_activity_type ON trips(activity_type);
CREATE INDEX IF NOT EXISTS idx_trips_difficulty ON trips(difficulty_level);
CREATE INDEX IF NOT EXISTS idx_trips_visibility ON trips(visibility);
CREATE INDEX IF NOT EXISTS idx_trips_water_features ON trips USING gin(water_features);
CREATE INDEX IF NOT EXISTS idx_trips_terrain_types ON trips USING gin(terrain_types);
CREATE INDEX IF NOT EXISTS idx_trips_route_geojson ON trips USING gin(route_geojson);
CREATE INDEX IF NOT EXISTS idx_trips_shared_with ON trips USING gin(shared_with);

-- Create activity_completions table for tracking user completions
CREATE TABLE IF NOT EXISTS activity_completions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    trip_id UUID NOT NULL REFERENCES trips(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    completed_at TIMESTAMPTZ NOT NULL,
    duration_minutes INTEGER,
    difficulty_rating INTEGER CHECK (difficulty_rating >= 1 AND difficulty_rating <= 5),
    overall_rating INTEGER CHECK (overall_rating >= 1 AND overall_rating <= 5),
    weather_conditions TEXT,
    trail_conditions TEXT,
    notes TEXT,
    photos UUID[],
    gpx_track JSONB,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(trip_id, user_id, completed_at)
);

-- Create activity_ratings table for detailed ratings
CREATE TABLE IF NOT EXISTS activity_ratings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    trip_id UUID NOT NULL REFERENCES trips(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    overall_rating INTEGER NOT NULL CHECK (overall_rating >= 1 AND overall_rating <= 5),
    difficulty_accuracy INTEGER CHECK (difficulty_accuracy >= 1 AND difficulty_accuracy <= 5),
    description_accuracy INTEGER CHECK (description_accuracy >= 1 AND description_accuracy <= 5),
    scenery_rating INTEGER CHECK (scenery_rating >= 1 AND scenery_rating <= 5),
    review_text TEXT,
    helpful_count INTEGER DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(trip_id, user_id)
);

-- Create activity_conditions table for real-time condition updates
CREATE TABLE IF NOT EXISTS activity_conditions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    trip_id UUID NOT NULL REFERENCES trips(id) ON DELETE CASCADE,
    reported_by UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    condition_type VARCHAR(50) NOT NULL, -- 'trail', 'weather', 'closure', 'hazard'
    severity VARCHAR(20), -- 'info', 'warning', 'danger'
    description TEXT NOT NULL,
    location GEOGRAPHY(POINT, 4326),
    photos UUID[],
    valid_from TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    valid_until TIMESTAMPTZ,
    verified BOOLEAN DEFAULT false,
    verified_by UUID REFERENCES users(id),
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- Create activity_share_links table for private sharing
CREATE TABLE IF NOT EXISTS activity_share_links (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    trip_id UUID NOT NULL REFERENCES trips(id) ON DELETE CASCADE,
    created_by UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    share_token VARCHAR(255) UNIQUE NOT NULL,
    permissions VARCHAR(20) DEFAULT 'view', -- 'view', 'edit'
    max_uses INTEGER,
    use_count INTEGER DEFAULT 0,
    expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    last_used_at TIMESTAMPTZ
);

-- Create indexes for new tables
CREATE INDEX idx_completions_trip ON activity_completions(trip_id);
CREATE INDEX idx_completions_user ON activity_completions(user_id);
CREATE INDEX idx_completions_date ON activity_completions(completed_at);

CREATE INDEX idx_ratings_trip ON activity_ratings(trip_id);
CREATE INDEX idx_ratings_user ON activity_ratings(user_id);

CREATE INDEX idx_conditions_trip ON activity_conditions(trip_id);
CREATE INDEX idx_conditions_type ON activity_conditions(condition_type);
CREATE INDEX idx_conditions_location ON activity_conditions USING GIST(location);

CREATE INDEX idx_share_links_token ON activity_share_links(share_token);
CREATE INDEX idx_share_links_trip ON activity_share_links(trip_id);

-- Add triggers for updated_at on new tables
CREATE TRIGGER update_activity_ratings_updated_at BEFORE UPDATE ON activity_ratings
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Create enum types for better data integrity
DO $$ BEGIN
    CREATE TYPE activity_type_enum AS ENUM (
        'hiking', 'biking', 'climbing', 'skiing', 'snowboarding',
        'kayaking', 'canoeing', 'rafting', 'swimming', 'surfing',
        'running', 'walking', 'backpacking', 'camping', 'fishing',
        'birdwatching', 'photography', 'sightseeing', 'general'
    );
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

DO $$ BEGIN
    CREATE TYPE difficulty_enum AS ENUM ('easy', 'moderate', 'hard', 'expert');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

DO $$ BEGIN
    CREATE TYPE route_type_enum AS ENUM ('out_and_back', 'loop', 'point_to_point', 'area');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;