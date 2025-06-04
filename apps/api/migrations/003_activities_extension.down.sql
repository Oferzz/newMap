-- Drop new tables
DROP TABLE IF EXISTS activity_share_links;
DROP TABLE IF EXISTS activity_conditions;
DROP TABLE IF EXISTS activity_ratings;
DROP TABLE IF EXISTS activity_completions;

-- Drop enum types
DROP TYPE IF EXISTS route_type_enum;
DROP TYPE IF EXISTS difficulty_enum;
DROP TYPE IF EXISTS activity_type_enum;

-- Drop indexes
DROP INDEX IF EXISTS idx_trips_activity_type;
DROP INDEX IF EXISTS idx_trips_difficulty;
DROP INDEX IF EXISTS idx_trips_visibility;
DROP INDEX IF EXISTS idx_trips_water_features;
DROP INDEX IF EXISTS idx_trips_terrain_types;
DROP INDEX IF EXISTS idx_trips_route_geojson;
DROP INDEX IF EXISTS idx_trips_shared_with;

-- Remove columns from trips table
ALTER TABLE trips DROP COLUMN IF EXISTS activity_type;
ALTER TABLE trips DROP COLUMN IF EXISTS difficulty_level;
ALTER TABLE trips DROP COLUMN IF EXISTS duration_hours;
ALTER TABLE trips DROP COLUMN IF EXISTS distance_km;
ALTER TABLE trips DROP COLUMN IF EXISTS elevation_gain_m;
ALTER TABLE trips DROP COLUMN IF EXISTS max_elevation_m;
ALTER TABLE trips DROP COLUMN IF EXISTS route_type;
ALTER TABLE trips DROP COLUMN IF EXISTS route_geojson;
ALTER TABLE trips DROP COLUMN IF EXISTS water_features;
ALTER TABLE trips DROP COLUMN IF EXISTS terrain_types;
ALTER TABLE trips DROP COLUMN IF EXISTS essential_gear;
ALTER TABLE trips DROP COLUMN IF EXISTS best_seasons;
ALTER TABLE trips DROP COLUMN IF EXISTS trail_conditions;
ALTER TABLE trips DROP COLUMN IF EXISTS accessibility_notes;
ALTER TABLE trips DROP COLUMN IF EXISTS parking_info;
ALTER TABLE trips DROP COLUMN IF EXISTS permits_required;
ALTER TABLE trips DROP COLUMN IF EXISTS hazards;
ALTER TABLE trips DROP COLUMN IF EXISTS emergency_contacts;
ALTER TABLE trips DROP COLUMN IF EXISTS visibility;
ALTER TABLE trips DROP COLUMN IF EXISTS shared_with;
ALTER TABLE trips DROP COLUMN IF EXISTS completion_count;
ALTER TABLE trips DROP COLUMN IF EXISTS average_rating;
ALTER TABLE trips DROP COLUMN IF EXISTS rating_count;
ALTER TABLE trips DROP COLUMN IF EXISTS featured;
ALTER TABLE trips DROP COLUMN IF EXISTS verified;