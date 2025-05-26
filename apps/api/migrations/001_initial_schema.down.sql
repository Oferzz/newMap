-- Drop triggers
DROP TRIGGER IF EXISTS update_suggestions_updated_at ON suggestions;
DROP TRIGGER IF EXISTS update_trip_waypoints_updated_at ON trip_waypoints;
DROP TRIGGER IF EXISTS update_places_updated_at ON places;
DROP TRIGGER IF EXISTS update_trips_updated_at ON trips;
DROP TRIGGER IF EXISTS update_users_updated_at ON users;

-- Drop function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop indexes
DROP INDEX IF EXISTS idx_suggestions_status;
DROP INDEX IF EXISTS idx_suggestions_user;
DROP INDEX IF EXISTS idx_suggestions_target;
DROP INDEX IF EXISTS idx_media_location;
DROP INDEX IF EXISTS idx_media_uploaded_by;
DROP INDEX IF EXISTS idx_waypoints_place;
DROP INDEX IF EXISTS idx_waypoints_trip;
DROP INDEX IF EXISTS idx_collaborators_user;
DROP INDEX IF EXISTS idx_collaborators_trip;
DROP INDEX IF EXISTS idx_places_search;
DROP INDEX IF EXISTS idx_places_parent;
DROP INDEX IF EXISTS idx_places_created_by;
DROP INDEX IF EXISTS idx_places_bounds;
DROP INDEX IF EXISTS idx_places_location;
DROP INDEX IF EXISTS idx_trips_search;
DROP INDEX IF EXISTS idx_trips_dates;
DROP INDEX IF EXISTS idx_trips_privacy;
DROP INDEX IF EXISTS idx_trips_status;
DROP INDEX IF EXISTS idx_trips_owner;
DROP INDEX IF EXISTS idx_users_status;
DROP INDEX IF EXISTS idx_users_username;
DROP INDEX IF EXISTS idx_users_email;

-- Drop tables in reverse order of dependencies
DROP TABLE IF EXISTS place_media;
DROP TABLE IF EXISTS place_collaborators;
DROP TABLE IF EXISTS user_friends;
DROP TABLE IF EXISTS suggestion_comments;
DROP TABLE IF EXISTS suggestions;
DROP TABLE IF EXISTS media_usage;
DROP TABLE IF EXISTS media;
DROP TABLE IF EXISTS trip_waypoints;
DROP TABLE IF EXISTS trip_collaborators;
DROP TABLE IF EXISTS places;
DROP TABLE IF EXISTS trips;
DROP TABLE IF EXISTS users;

-- Drop extensions
DROP EXTENSION IF EXISTS "pg_trgm";
DROP EXTENSION IF EXISTS "postgis";
DROP EXTENSION IF EXISTS "uuid-ossp";