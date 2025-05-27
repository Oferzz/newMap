#!/bin/bash
# Startup script for Render deployment
# This script loads secrets and starts the application

set -e

echo "Starting Trip Planner API..."

# Function to load secrets from Render's secret file
load_render_secrets() {
    SECRETS_FILE="/etc/secrets/tokens"
    
    if [ -f "$SECRETS_FILE" ]; then
        echo "Loading secrets from Render secrets file..."
        
        # Read the file line by line
        while IFS='=' read -r key value; do
            # Skip empty lines and comments
            if [ -n "$key" ] && [[ ! "$key" =~ ^[[:space:]]*# ]]; then
                # Trim whitespace
                key=$(echo "$key" | xargs)
                value=$(echo "$value" | xargs)
                
                # Export the variable
                export "$key"="$value"
                echo "Loaded secret: $key"
            fi
        done < "$SECRETS_FILE"
        
        # Map MAPBOX_ACCESS_TOKEN to MAPBOX_API_KEY if needed
        if [ -n "$MAPBOX_ACCESS_TOKEN" ] && [ -z "$MAPBOX_API_KEY" ]; then
            export MAPBOX_API_KEY="$MAPBOX_ACCESS_TOKEN"
            echo "Mapped MAPBOX_ACCESS_TOKEN to MAPBOX_API_KEY"
        fi
        
        # Map MONGODB_URI to DATABASE_URL if PostgreSQL URL is not set
        # This is for backward compatibility or migration purposes
        if [ -n "$MONGODB_URI" ] && [ -z "$DATABASE_URL" ]; then
            echo "Warning: MONGODB_URI found but app uses PostgreSQL. Please update to use DATABASE_URL"
        fi
    else
        echo "No Render secrets file found at $SECRETS_FILE"
    fi
}

# Function to wait for database
wait_for_database() {
    echo "Waiting for database to be ready..."
    
    # Parse DATABASE_URL to get host and port
    if [[ $DATABASE_URL =~ postgresql://([^:]+):([^@]+)@([^:]+):([0-9]+)/(.+) ]]; then
        DB_HOST="${BASH_REMATCH[3]}"
        DB_PORT="${BASH_REMATCH[4]}"
        
        until pg_isready -h "$DB_HOST" -p "$DB_PORT" 2>/dev/null; do
            echo "Database is unavailable - sleeping"
            sleep 2
        done
        
        echo "Database is ready!"
    else
        echo "Could not parse DATABASE_URL, skipping database wait"
    fi
}

# Function to run migrations
run_migrations() {
    if [ "${RUN_MIGRATIONS}" = "true" ]; then
        echo "Running database migrations..."
        ./server migrate up
        echo "Migrations completed!"
    else
        echo "Skipping migrations (RUN_MIGRATIONS != true)"
    fi
}

# Main execution
echo "=== Trip Planner API Startup ==="
echo "Environment: ${ENVIRONMENT:-development}"

# Load secrets from Render
load_render_secrets

# Verify critical environment variables
if [ -z "$DATABASE_URL" ]; then
    echo "ERROR: DATABASE_URL is not set!"
    exit 1
fi

if [ -z "$JWT_SECRET" ]; then
    echo "WARNING: JWT_SECRET is not set, using default (not secure for production!)"
fi

if [ -z "$MAPBOX_API_KEY" ] && [ -z "$MAPBOX_ACCESS_TOKEN" ]; then
    echo "WARNING: No Mapbox API key found (MAPBOX_API_KEY or MAPBOX_ACCESS_TOKEN)"
fi

# Wait for database
wait_for_database

# Run migrations
run_migrations

# Start the server
echo "Starting server on port ${PORT:-8080}..."
exec ./server