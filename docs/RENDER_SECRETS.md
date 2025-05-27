# Render Secrets Configuration

This document explains how the application uses Render's secret files feature to securely manage sensitive configuration.

## Overview

The application is configured to read secrets from `/etc/secrets/tokens` when deployed on Render. This file should contain key-value pairs for sensitive configuration like API keys and database credentials.

## Required Secrets

Add these secrets to your `/etc/secrets/tokens` file in Render:

```
MAPBOX_ACCESS_TOKEN=pk.your_mapbox_access_token_here
MONGODB_URI=mongodb+srv://... (if needed for migration)
```

## How It Works

1. **Automatic Loading**: The application automatically loads secrets from `/etc/secrets/tokens` on startup
2. **Environment Variable Mapping**: Secrets are mapped to environment variables
3. **Fallback Support**: The app falls back to standard environment variables if the secrets file doesn't exist

## Setting Up Secrets in Render

1. Go to your Render service dashboard
2. Navigate to the "Environment" tab
3. Click on "Secret Files"
4. Create a file named `tokens` with path `/etc/secrets/tokens`
5. Add your secrets in the format:
   ```
   MAPBOX_ACCESS_TOKEN=your_actual_token_here
   MONGODB_URI=your_mongodb_connection_string
   ```

## Supported Secret Names

| Secret Name | Purpose | Maps To |
|------------|---------|---------|
| `MAPBOX_ACCESS_TOKEN` | Mapbox API access | `MAPBOX_API_KEY` |
| `MONGODB_URI` | MongoDB connection (legacy) | Used for data migration |

## Implementation Details

### Backend (Go)

The Go application loads secrets in `internal/config/config.go`:

```go
func loadRenderSecrets() {
    secretsFile := "/etc/secrets/tokens"
    // Reads file and sets environment variables
}
```

### Startup Script

The `startup.sh` script ensures secrets are loaded before the application starts:

```bash
load_render_secrets() {
    SECRETS_FILE="/etc/secrets/tokens"
    # Loads each key-value pair as environment variable
}
```

## Security Best Practices

1. **Never commit secrets** to the repository
2. **Use secret files** instead of plain environment variables for sensitive data
3. **Rotate secrets regularly**
4. **Limit access** to the Render dashboard

## Troubleshooting

### Secrets Not Loading

1. Check the Render logs for "Loading secrets from Render secrets file..."
2. Verify the file exists at `/etc/secrets/tokens`
3. Ensure the file format is correct (KEY=value, one per line)

### Mapbox Token Issues

- The app supports both `MAPBOX_ACCESS_TOKEN` and `MAPBOX_API_KEY`
- The startup script automatically maps between them
- Frontend builds require the token to be set as `VITE_MAPBOX_TOKEN` in Render environment variables

### Database Connection

- The app uses PostgreSQL, not MongoDB
- `MONGODB_URI` is only kept for backward compatibility or migration purposes
- Ensure `DATABASE_URL` is set in Render environment variables

## Local Development

For local development, use a `.env` file instead:

```bash
cp .env.example .env
# Edit .env with your local values
```

The application will load from `.env` when `/etc/secrets/tokens` is not available.