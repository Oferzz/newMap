-- Drop triggers
DROP TRIGGER IF EXISTS update_collections_updated_at ON collections;
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop indexes
DROP INDEX IF EXISTS idx_collection_collaborators_user_id;
DROP INDEX IF EXISTS idx_collection_locations_coords;
DROP INDEX IF EXISTS idx_collection_locations_collection_id;
DROP INDEX IF EXISTS idx_collections_updated_at;
DROP INDEX IF EXISTS idx_collections_privacy;
DROP INDEX IF EXISTS idx_collections_user_id;

-- Drop tables (order matters due to foreign keys)
DROP TABLE IF EXISTS collection_collaborators;
DROP TABLE IF EXISTS collection_locations;
DROP TABLE IF EXISTS collections;