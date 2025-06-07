#!/bin/bash

# Setup Supabase Environment Variables
# This script helps you configure environment variables for Supabase migration

echo "🚀 Setting up Supabase environment for NewMap"
echo "=============================================="

# Check if Supabase CLI is installed
if ! command -v supabase &> /dev/null; then
    echo "⚠️  Supabase CLI not found. Install it with:"
    echo "   npm install -g supabase"
    echo "   or visit: https://supabase.com/docs/guides/cli"
    exit 1
fi

# Function to prompt for input with default value
prompt_with_default() {
    local prompt="$1"
    local default="$2"
    local result
    
    if [[ -n "$default" ]]; then
        read -p "$prompt [$default]: " result
        echo "${result:-$default}"
    else
        read -p "$prompt: " result
        echo "$result"
    fi
}

echo ""
echo "📋 Please provide your Supabase project details:"
echo "   You can find these in your Supabase project dashboard at:"
echo "   https://supabase.com/dashboard/project/YOUR_PROJECT_ID/settings/api"
echo ""

# Get project details
PROJECT_URL=$(prompt_with_default "Supabase Project URL (e.g., https://xyzabc.supabase.co)" "")
ANON_KEY=$(prompt_with_default "Supabase Anon Key (public)" "")
SERVICE_KEY=$(prompt_with_default "Supabase Service Role Key (secret)" "")

# Validate inputs
if [[ -z "$PROJECT_URL" || -z "$ANON_KEY" || -z "$SERVICE_KEY" ]]; then
    echo "❌ Error: All fields are required!"
    exit 1
fi

# Create or update backend .env file
echo ""
echo "🔧 Creating backend environment file..."
BACKEND_ENV_FILE="apps/api/.env"

cat > "$BACKEND_ENV_FILE" << EOF
# Supabase Configuration
SUPABASE_PROJECT_URL=$PROJECT_URL
SUPABASE_PROJECT_KEY=$SERVICE_KEY
SUPABASE_ANON_KEY=$ANON_KEY

# Local Database (fallback if Supabase not configured)
DATABASE_URL=postgresql://newMap_user:newMap_pass@localhost:5432/newMap?sslmode=disable

# Redis Cache
REDIS_URL=redis://localhost:6379

# JWT Configuration (can be removed when using Supabase Auth)
JWT_SECRET=your-super-secret-jwt-key-change-in-production
JWT_ISSUER=newMap

# Server Configuration
PORT=8080
ENVIRONMENT=development

# External Services
MAPBOX_API_KEY=your_mapbox_token_here

# Media Configuration
MEDIA_PATH=/tmp/media
CDN_URL=http://localhost:8080

# CORS
ALLOWED_ORIGINS=http://localhost:3000,http://localhost:5173
EOF

echo "✅ Backend environment file created: $BACKEND_ENV_FILE"

# Create or update frontend .env file
echo ""
echo "🔧 Creating frontend environment file..."
FRONTEND_ENV_FILE="apps/web/.env"

cat > "$FRONTEND_ENV_FILE" << EOF
# API Configuration
VITE_API_URL=http://localhost:8080

# Supabase Configuration
VITE_SUPABASE_PROJECT_URL=$PROJECT_URL
VITE_SUPABASE_PROJECT_KEY=$ANON_KEY

# Mapbox Configuration
VITE_MAPBOX_TOKEN=your_mapbox_token_here

# Feature Flags
VITE_ENABLE_OFFLINE_MODE=false
VITE_ENABLE_REAL_TIME=true
EOF

echo "✅ Frontend environment file created: $FRONTEND_ENV_FILE"

# Create .env file for Docker Compose
echo ""
echo "🔧 Creating Docker Compose environment file..."
DOCKER_ENV_FILE=".env"

cat > "$DOCKER_ENV_FILE" << EOF
# Supabase Configuration
SUPABASE_PROJECT_URL=$PROJECT_URL
SUPABASE_PROJECT_KEY=$SERVICE_KEY
SUPABASE_ANON_KEY=$ANON_KEY

# Mapbox API Key
MAPBOX_API_KEY=your_mapbox_token_here

# Local Database Configuration (for fallback)
DB_USER=newMap_user
DB_PASSWORD=newMap_pass
DB_NAME=newMap

# Redis Configuration
REDIS_PASSWORD=changeme

# JWT Configuration
JWT_SECRET=your-super-secret-jwt-key-change-in-production

# Environment
ENVIRONMENT=development
ALLOWED_ORIGINS=http://localhost:3000,http://localhost:5173

# URLs for production
VITE_API_URL=http://localhost:8080/api/v1
VITE_WS_URL=ws://localhost:8080
CDN_URL=http://localhost:8080
EOF

echo "✅ Docker Compose environment file created: $DOCKER_ENV_FILE"

echo ""
echo "🎉 Environment setup complete!"
echo ""
echo "📋 Next Steps:"
echo "   1. Update your Mapbox API keys in the .env files"
echo "   2. Run the Supabase migration script in your Supabase SQL Editor:"
echo "      cat scripts/migrate-to-supabase.sql"
echo "   3. Install dependencies:"
echo "      cd apps/web && npm install"
echo "      cd apps/api && go mod tidy"
echo "   4. Start the application:"
echo "      docker-compose up (uses local DB + Supabase)"
echo "      or"
echo "      cd apps/api && go run cmd/server/main.go (backend only)"
echo "      cd apps/web && npm run dev (frontend only)"
echo ""
echo "🔐 Security Notes:"
echo "   - The .env files contain sensitive keys - don't commit them to git"
echo "   - Service role key should only be used in backend/server environments"
echo "   - Anon key is safe to use in frontend applications"
echo ""
echo "📚 Documentation:"
echo "   - Migration guide: SUPABASE_MIGRATION_PLAN_UPDATED.md"
echo "   - Authentication guide: SUPABASE_AUTH_MIGRATION.md"
echo "   - Database guide: SUPABASE_DATABASE_MIGRATION.md"