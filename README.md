# Trip Planning Platform

A collaborative trip planning platform with real-time features, role-based permissions, and map-based interactions.

## Tech Stack

- **Backend**: Go 1.21+ with Gin framework
- **Frontend**: React 18+ with TypeScript (coming soon)
- **Database**: MongoDB Atlas, Redis
- **Real-time**: Socket.io (coming soon)
- **Deployment**: Render.com with Cloudflare CDN

## Project Structure

```
.
├── apps/
│   ├── api/           # Go backend API
│   └── web/           # React frontend (coming soon)
├── packages/          # Shared packages
└── .github/           # GitHub Actions workflows
```

## Getting Started

### Prerequisites

- Go 1.21+
- MongoDB (local or Atlas)
- Redis (optional for caching)

### Backend Development

1. Navigate to the API directory:
```bash
cd apps/api
```

2. Copy environment variables:
```bash
cp .env.example .env
```

3. Install dependencies:
```bash
go mod download
```

4. Run the server:
```bash
go run cmd/server/main.go
```

### Running Tests

```bash
cd apps/api
go test -v ./...
```

## API Endpoints

### Authentication
- `POST /api/v1/auth/register` - Register new user
- `POST /api/v1/auth/login` - Login user
- `POST /api/v1/auth/refresh` - Refresh access token

### User Profile
- `GET /api/v1/users/me` - Get current user profile
- `PUT /api/v1/users/me` - Update profile
- `PUT /api/v1/users/me/password` - Change password
- `DELETE /api/v1/users/me` - Delete account

## Features

### Core Functionality
- ✅ User authentication with JWT
- ✅ User registration and login
- ✅ Password hashing with bcrypt
- ✅ MongoDB integration
- ✅ Comprehensive test coverage
- ✅ CI/CD with GitHub Actions
- 🚧 Role-based access control (RBAC)
- 🚧 Trip CRUD operations
- 🚧 Real-time collaboration

### Security
- JWT-based authentication
- Password hashing
- Input validation
- Rate limiting (coming soon)

## License

This project is licensed under the MIT License.