# Deployment Guide for Trip Planning Platform

This guide explains how to deploy the Trip Planning Platform to Render.com.

## Prerequisites

1. A Render.com account
2. A GitHub repository with the code
3. PostgreSQL database (provided by Render)
4. Redis instance (provided by Render)
5. Mapbox API key

## Step-by-Step Deployment

### 1. Fork/Clone the Repository

First, ensure your code is pushed to GitHub:

```bash
git add .
git commit -m "feat: prepare for deployment"
git push origin main
```

### 2. Create Render Services

Log in to [Render.com](https://render.com) and create the following services:

#### A. PostgreSQL Database

1. Click "New +" → "PostgreSQL"
2. Configure:
   - Name: `trip-planner-db`
   - Database: `trip_planner`
   - User: `trip_planner_user`
   - Region: Choose closest to your users
   - Plan: Starter ($7/month) or Free (for testing)
3. Click "Create Database"
4. Wait for the database to be ready
5. Note the connection strings (Internal and External)

#### B. Redis Instance

1. Click "New +" → "Redis"
2. Configure:
   - Name: `trip-planner-cache`
   - Region: Same as database
   - Maxmemory Policy: `allkeys-lru`
   - Plan: Starter ($7/month) or Free (for testing)
3. Click "Create Redis"
4. Note the connection strings

#### C. Backend API Service

1. Click "New +" → "Web Service"
2. Connect your GitHub repository
3. Configure:
   - Name: `trip-planner-api`
   - Environment: Docker
   - Dockerfile Path: `./apps/api/Dockerfile.render`
   - Docker Context Directory: `.`
   - Branch: `main`
4. Add secret file:
   - Go to "Environment" → "Secret Files"
   - Create file: `/etc/secrets/tokens`
   - Add content:
     ```
     MAPBOX_ACCESS_TOKEN=your_mapbox_token_here
     MONGODB_URI=your_mongodb_uri_if_needed
     ```
5. Add environment variables:
   ```
   PORT=8080
   DATABASE_URL=[Internal Database URL from PostgreSQL]
   REDIS_URL=[Internal Redis URL]
   INTERNAL_REDIS_URL=[Internal Redis URL]
   JWT_SECRET=[Generate a secure random string]
   JWT_ISSUER=trip-planner
   MEDIA_PATH=/data/media
   CDN_URL=https://trip-planner-api.onrender.com
   DB_MIGRATIONS_PATH=./migrations
   ENVIRONMENT=production
   RUN_MIGRATIONS=true
   ```
   Note: MAPBOX_ACCESS_TOKEN is loaded from /etc/secrets/tokens
6. Add a disk:
   - Mount Path: `/data`
   - Size: 20GB (for media storage)
7. Click "Create Web Service"

#### D. Frontend Web Service

1. Click "New +" → "Static Site"
2. Connect your GitHub repository
3. Configure:
   - Name: `trip-planner-web`
   - Build Command: `cd apps/web && npm install && npm run build`
   - Publish Directory: `apps/web/dist`
   - Branch: `main`
4. Add environment variables:
   ```
   VITE_API_URL=https://trip-planner-api.onrender.com/api/v1
   VITE_WS_URL=wss://trip-planner-api.onrender.com
   VITE_MAPBOX_TOKEN=[Your Mapbox API key]
   ```
5. Click "Create Static Site"

### 3. Configure render.yaml

The repository includes a `render.yaml` file that defines all services. You can also use Render Blueprints:

1. Go to Blueprints in Render dashboard
2. Click "New Blueprint Instance"
3. Connect your GitHub repository
4. Select the branch with `render.yaml`
5. Review and apply the configuration

### 4. Set Up GitHub Secrets

For automated deployments, add these secrets to your GitHub repository:

1. Go to Settings → Secrets and variables → Actions
2. Add:
   - `RENDER_API_KEY`: Get from Render Account Settings
   - `RENDER_SERVICE_ID`: Your API service ID from Render
   - `SLACK_WEBHOOK` (optional): For deployment notifications

### 5. Run Database Migrations

After the API service is deployed:

1. Go to the API service in Render
2. Click "Shell" tab
3. Run:
   ```bash
   cd /app
   ./server migrate up
   ```

### 6. Verify Deployment

1. Check service logs in Render dashboard
2. Visit your API health endpoint: `https://trip-planner-api.onrender.com/api/health`
3. Visit your frontend: `https://trip-planner-web.onrender.com`

## Environment-Specific Configuration

### Production Environment Variables

Make sure these are set in Render:

- `ENVIRONMENT=production`
- `JWT_SECRET`: Use a strong, unique secret
- `ALLOWED_ORIGINS`: Set to your frontend URL
- `CORS_ORIGINS`: Set to your frontend URL

### Security Considerations

1. Enable Render's DDoS protection
2. Set up Cloudflare (optional) for CDN and additional security
3. Configure rate limiting in the API
4. Use environment-specific secrets
5. Enable HTTPS (automatic with Render)

## Monitoring and Maintenance

### Health Checks

The API includes health endpoints:
- `/api/health` - Basic health check
- `/api/health/ready` - Readiness check (includes DB connection)

### Logs

Access logs through:
1. Render dashboard → Service → Logs
2. Use Render's log streaming API
3. Integrate with external logging services

### Backups

1. Enable automatic PostgreSQL backups in Render
2. Set up regular exports for critical data
3. Test restore procedures regularly

## Troubleshooting

### Common Issues

1. **Database Connection Failed**
   - Check DATABASE_URL is using internal URL
   - Verify PostgreSQL service is running
   - Check network connectivity

2. **Redis Connection Failed**
   - Verify REDIS_URL is correct
   - Check Redis service status

3. **Frontend Can't Connect to API**
   - Verify CORS settings
   - Check API URL in frontend env vars
   - Ensure API service is running

4. **Media Upload Issues**
   - Verify disk is mounted at `/data`
   - Check file permissions
   - Ensure sufficient disk space

### Debug Commands

SSH into your API service:
```bash
# Check environment variables
env | grep -E 'DATABASE|REDIS|JWT'

# Test database connection
psql $DATABASE_URL -c "SELECT 1"

# Check disk usage
df -h /data

# View recent logs
tail -f /var/log/app.log
```

## Scaling

When ready to scale:

1. **Vertical Scaling**: Upgrade to higher Render plans
2. **Horizontal Scaling**: Increase min/max instances in render.yaml
3. **Database Scaling**: Upgrade PostgreSQL plan
4. **Caching**: Increase Redis memory allocation

## Cost Optimization

Estimated monthly costs:
- PostgreSQL Starter: $7
- Redis Starter: $7
- API Service (Starter): $7
- Static Site: Free
- **Total**: ~$21/month

For production, consider:
- Standard plans for better performance
- Reserved instances for cost savings
- CDN integration for static assets

## Support

- Render Documentation: https://render.com/docs
- Render Community: https://community.render.com
- GitHub Issues: Report bugs in the repository