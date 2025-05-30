# Trip Planning Platform

A collaborative trip planning platform with real-time features, built with Go, React, PostgreSQL, and Redis.

## Features

### Core Features (No Authentication Required)
- 🗺️ Interactive map-based trip planning with Mapbox
- 📍 Save locations and create routes (stored in browser cache)
- 🔍 Place search and exploration
- 🎯 Drag-and-drop itinerary planning
- 📱 Responsive design for mobile and desktop
- ⚡ Instant access - start planning immediately

### Collaboration Features (Authentication Required)
- 👥 Real-time collaboration with multiple users
- 🔒 Role-based access control (Owner, Admin, Editor, Viewer)
- 💬 Real-time suggestions and comments
- 🔄 WebSocket support for live updates
- 📊 Trip statistics and analytics
- ☁️ Cloud storage and sync across devices
- 🚀 Fast performance with Redis caching

> **Note**: Users can save locations, create routes, search, and explore the platform without authentication. All data is stored locally in browser cache. For sharing trips with others and real-time collaboration, authentication is required.

## Tech Stack

- **Backend**: Go 1.23+ with Gin framework
- **Frontend**: React 18+ with TypeScript, Redux Toolkit
- **Database**: PostgreSQL with PostGIS extension
- **Cache**: Redis for performance optimization
- **Maps**: Mapbox GL JS
- **Real-time**: Socket.io for WebSocket communication
- **Deployment**: Docker, Render.com, GitHub Actions CI/CD

## Quick Start with Docker

### Prerequisites

- Docker and Docker Compose
- Mapbox API key (get one at https://mapbox.com)
- Git

### 1. Clone the repository

```bash
git clone https://github.com/yourusername/newMap.git
cd newMap
```

### 2. Set up environment variables

```bash
cp .env.example .env
# Edit .env and add your Mapbox API key and other configurations
```

### 3. Start the application

```bash
# For development
docker-compose up

# For production
docker-compose -f docker-compose.prod.yml up
```

The application will be available at:
- Frontend: http://localhost:3000
- API: http://localhost:8080
- API Documentation: http://localhost:8080/swagger

### 4. Using the deployment script

We provide a convenient deployment script for managing the application:

```bash
# Make the script executable
chmod +x deploy.sh

# Build images
./deploy.sh build

# Start services
./deploy.sh start

# View logs
./deploy.sh logs

# Run database migrations
./deploy.sh migrate

# Create database backup
./deploy.sh backup

# Stop services
./deploy.sh stop
```

## Local Development

### Backend Development

```bash
cd apps/api
go mod download
go run cmd/server/main.go
```

### Frontend Development

```bash
cd apps/web
npm install
npm run dev
```

### Running Tests

```bash
# Backend tests
cd apps/api
go test -v ./...

# Frontend tests
cd apps/web
npm test
```

## Deployment to Render

This application is configured for easy deployment to Render.com. See [DEPLOYMENT_GUIDE.md](./DEPLOYMENT_GUIDE.md) for detailed instructions.

### Quick Deploy to Render

1. Fork this repository
2. Create a Render account
3. Use the `render.yaml` blueprint or manually create:
   - PostgreSQL database
   - Redis instance
   - Web service for API (Docker)
   - Static site for frontend
4. Set environment variables in Render dashboard
5. Deploy!

## Project Structure

```
.
├── apps/
│   ├── api/                 # Go backend API
│   │   ├── cmd/            # Application entrypoints
│   │   ├── internal/       # Private application code
│   │   ├── migrations/     # Database migrations
│   │   └── Dockerfile      # API container definition
│   └── web/                # React frontend
│       ├── src/            # Source code
│       └── Dockerfile      # Frontend container definition
├── packages/               # Shared packages (types, utils)
├── .github/               # GitHub Actions workflows
├── docker-compose.yml     # Development environment
├── docker-compose.prod.yml # Production environment
├── render.yaml            # Render.com configuration
└── deploy.sh             # Deployment helper script
```

## API Documentation

### Public Endpoints (No Authentication Required)
- `GET /api/v1/places/search` - Search places (Mapbox integration)
- `GET /api/v1/health` - Health check

### Authentication Endpoints
- `POST /api/v1/auth/register` - Register new user
- `POST /api/v1/auth/login` - Login user
- `POST /api/v1/auth/refresh` - Refresh access token
- `POST /api/v1/auth/logout` - Logout user

### Trip Management (Authentication Required)
- `GET /api/v1/trips` - List user's trips
- `POST /api/v1/trips` - Create new trip
- `GET /api/v1/trips/:id` - Get trip details
- `PUT /api/v1/trips/:id` - Update trip
- `DELETE /api/v1/trips/:id` - Delete trip

### Collections (Authentication Required)
- `GET /api/v1/collections` - List user's collections
- `POST /api/v1/collections` - Create new collection
- `POST /api/v1/collections/:id/locations` - Add location to collection
- `DELETE /api/v1/collections/:id` - Delete collection

### Collaboration (Authentication Required)
- `POST /api/v1/trips/:id/collaborators` - Add collaborator
- `DELETE /api/v1/trips/:id/collaborators/:userId` - Remove collaborator
- `PUT /api/v1/trips/:id/collaborators/:userId/role` - Update collaborator role

### Places (Mixed Access)
- `GET /api/v1/places` - Search places (public)
- `POST /api/v1/places` - Create custom place (requires auth)
- `GET /api/v1/places/:id` - Get place details (public)
- `PUT /api/v1/places/:id` - Update place (requires auth)
- `DELETE /api/v1/places/:id` - Delete place (requires auth)

## Environment Variables

Key environment variables (see `.env.example` for full list):

```bash
# Database
DATABASE_URL=postgresql://user:pass@localhost:5432/newMap

# Redis
REDIS_URL=redis://localhost:6379

# JWT
JWT_SECRET=your-secure-secret-key

# External Services
MAPBOX_API_KEY=your-mapbox-api-key

# Frontend
VITE_API_URL=http://localhost:8080/api/v1
VITE_WS_URL=ws://localhost:8080
```

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## Security

- JWT-based authentication with refresh tokens
- Password hashing with bcrypt
- Input validation and sanitization
- Rate limiting on API endpoints
- CORS configuration
- SQL injection protection
- XSS protection

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Support

- Documentation: See `/docs` folder
- Issues: GitHub Issues
- Discussions: GitHub Discussions