#!/bin/bash
# Startup script for Render deployment
# Simple startup that relies on Render's environment variables

set -e

echo "=== Trip Planner API Startup ==="
echo "Environment: ${ENVIRONMENT:-production}"

# Verify critical environment variables
echo "Checking environment variables..."

if [ -z "$DATABASE_URL" ]; then
    echo "ERROR: DATABASE_URL is not set!"
    echo "Please set DATABASE_URL in your Render environment variables"
    exit 1
fi

# Log which environment variables are set (without values for security)
[ -n "$DATABASE_URL" ] && echo "✓ DATABASE_URL is set"
[ -n "$REDIS_URL" ] && echo "✓ REDIS_URL is set"
[ -n "$JWT_SECRET" ] && echo "✓ JWT_SECRET is set" || echo "⚠ JWT_SECRET not set (using default)"
[ -n "$PORT" ] && echo "✓ PORT is set to $PORT" || echo "✓ PORT defaulting to 8080"

# Wait for database to be ready
echo "Checking database connection..."
if [[ $DATABASE_URL =~ postgresql://([^:]+):([^@]+)@([^:]+):([0-9]+)/(.+) ]]; then
    DB_HOST="${BASH_REMATCH[3]}"
    DB_PORT="${BASH_REMATCH[4]}"
    
    # Try a few times to connect
    for i in {1..10}; do
        if pg_isready -h "$DB_HOST" -p "$DB_PORT" 2>/dev/null; then
            echo "✓ Database is ready!"
            break
        else
            echo "Waiting for database... (attempt $i/10)"
            sleep 2
        fi
    done
else
    echo "⚠ Could not parse DATABASE_URL for health check"
fi

# Start the server
echo "Starting server on port ${PORT:-8080}..."
exec ./server