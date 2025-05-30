-- Create collections table
CREATE TABLE collections (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    privacy VARCHAR(20) NOT NULL DEFAULT 'private' CHECK (privacy IN ('public', 'private', 'friends')),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create collection_locations table
CREATE TABLE collection_locations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    collection_id UUID NOT NULL REFERENCES collections(id) ON DELETE CASCADE,
    name VARCHAR(255),
    latitude DECIMAL(10, 8) NOT NULL,
    longitude DECIMAL(11, 8) NOT NULL,
    added_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create collection_collaborators table
CREATE TABLE collection_collaborators (
    collection_id UUID NOT NULL REFERENCES collections(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(20) NOT NULL DEFAULT 'viewer' CHECK (role IN ('viewer', 'editor')),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    PRIMARY KEY (collection_id, user_id)
);

-- Create indexes for better performance
CREATE INDEX idx_collections_user_id ON collections(user_id);
CREATE INDEX idx_collections_privacy ON collections(privacy);
CREATE INDEX idx_collections_updated_at ON collections(updated_at DESC);

CREATE INDEX idx_collection_locations_collection_id ON collection_locations(collection_id);
CREATE INDEX idx_collection_locations_coords ON collection_locations(latitude, longitude);

CREATE INDEX idx_collection_collaborators_user_id ON collection_collaborators(user_id);

-- Add triggers for updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_collections_updated_at
    BEFORE UPDATE ON collections
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();