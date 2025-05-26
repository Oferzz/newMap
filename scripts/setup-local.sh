#!/bin/bash

# Local Development Setup Script
set -e

echo "🚀 Setting up Trip Planning Platform for local development..."

# Check prerequisites
echo "📋 Checking prerequisites..."

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed. Please install Go 1.21 or later."
    exit 1
fi

# Check if Node.js is installed
if ! command -v node &> /dev/null; then
    echo "❌ Node.js is not installed. Please install Node.js 18 or later."
    exit 1
fi

# Check if Docker is installed
if ! command -v docker &> /dev/null; then
    echo "⚠️  Docker is not installed. You'll need it for PostgreSQL and Redis."
fi

echo "✅ Prerequisites check passed!"

# Create .env files
echo "📝 Creating environment files..."

# Backend .env
cat > apps/api/.env.local << EOF
# Database
DATABASE_URL=postgres://tripuser:trippass@localhost:5432/tripdb?sslmode=disable

# Redis
REDIS_URL=redis://localhost:6379

# JWT
JWT_SECRET=your-super-secret-jwt-key-change-in-production
JWT_ISSUER=trip-planner-local

# Server
PORT=8080
ENVIRONMENT=development

# Media
MEDIA_PATH=./uploads
CDN_URL=http://localhost:8080

# Mapbox
MAPBOX_API_KEY=your-mapbox-api-key-here
EOF

# Frontend .env
cat > apps/web/.env.local << EOF
# API Configuration
VITE_API_URL=http://localhost:8080/api/v1
VITE_WS_URL=ws://localhost:8080

# Mapbox
VITE_MAPBOX_TOKEN=your-mapbox-public-token-here

# Features
VITE_ENABLE_PWA=false
VITE_ENABLE_ANALYTICS=false
EOF

echo "✅ Environment files created!"

# Start services with Docker Compose
if command -v docker &> /dev/null; then
    echo "🐳 Starting PostgreSQL and Redis with Docker..."
    
    # Create docker-compose for local development
    cat > docker-compose.local.yml << EOF
version: '3.8'

services:
  postgres:
    image: postgis/postgis:16-3.4
    container_name: trip-planner-postgres
    environment:
      POSTGRES_USER: tripuser
      POSTGRES_PASSWORD: trippass
      POSTGRES_DB: tripdb
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data

  redis:
    image: redis:7-alpine
    container_name: trip-planner-redis
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data

volumes:
  postgres_data:
  redis_data:
EOF

    docker-compose -f docker-compose.local.yml up -d
    
    echo "⏳ Waiting for services to start..."
    sleep 5
fi

# Install backend dependencies
echo "📦 Installing backend dependencies..."
cd apps/api
go mod download
go mod tidy

# Run migrations
echo "🗄️  Running database migrations..."
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
migrate -path ./migrations -database "postgres://tripuser:trippass@localhost:5432/tripdb?sslmode=disable" up

# Create media directory
mkdir -p uploads

cd ../..

# Install frontend dependencies
echo "📦 Installing frontend dependencies..."
cd apps/web
npm install

cd ../..

echo "✅ Setup complete!"
echo ""
echo "📋 Next steps:"
echo "1. Add your Mapbox API keys to the .env.local files"
echo "2. Start the backend: cd apps/api && go run cmd/server/main.go"
echo "3. Start the frontend: cd apps/web && npm run dev"
echo "4. Open http://localhost:5173 in your browser"
echo ""
echo "🛑 To stop services: docker-compose -f docker-compose.local.yml down"
echo ""
echo "Happy coding! 🎉"