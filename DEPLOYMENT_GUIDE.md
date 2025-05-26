# Production Deployment Guide

## 🚀 Quick Start Deployment

### Prerequisites
1. GitHub account with the repository
2. Render.com account (free tier available)
3. Mapbox account for API keys

### Step-by-Step Deployment

#### 1. Fork/Clone Repository
```bash
git clone https://github.com/Oferzz/newMap.git
cd newMap
```

#### 2. Set Up Render Account
1. Go to [render.com](https://render.com) and sign up
2. Connect your GitHub account
3. Add a payment method (required for database and Redis)

#### 3. Deploy from Blueprint
1. In Render dashboard, click **"New +"** → **"Blueprint"**
2. Connect your GitHub repository
3. Select the `main` branch
4. Review the services that will be created:
   - Web Service (API) - $7/month
   - Static Site (Frontend) - Free
   - PostgreSQL Database - $7/month
   - Redis - $7/month
   - **Total: ~$21/month** (starter plan)

#### 4. Configure Environment Variables
Before clicking "Apply", set these environment variables:

**For API Service:**
```
MAPBOX_API_KEY=sk.your_mapbox_secret_key
```

**For Web Service:**
```
VITE_MAPBOX_TOKEN=pk.your_mapbox_public_token
```

#### 5. Deploy
1. Click **"Apply"** to start deployment
2. Wait 10-15 minutes for all services to deploy
3. Your app will be available at:
   - API: `https://trip-planner-api.onrender.com`
   - Web: `https://trip-planner-web.onrender.com`

## 📊 Architecture Overview

```
┌─────────────────┐     ┌─────────────────┐
│   React Web     │────▶│    Go API       │
│  (Static Site)  │     │  (Web Service)  │
└─────────────────┘     └─────┬───────────┘
                              │
                    ┌─────────┴─────────┐
                    ▼                   ▼
            ┌─────────────┐     ┌─────────────┐
            │ PostgreSQL  │     │    Redis    │
            │   (PostGIS) │     │   (Cache)   │
            └─────────────┘     └─────────────┘
```

## 🔧 Configuration Details

### Database Schema
The PostgreSQL database includes:
- PostGIS extension for geospatial queries
- Tables: users, trips, places, media
- Automatic migrations on deployment

### API Endpoints
- Health: `GET /api/health`
- Auth: `POST /api/v1/auth/register`, `POST /api/v1/auth/login`
- Users: `GET/PUT /api/v1/users/:id`
- Trips: Full CRUD at `/api/v1/trips`
- Places: Full CRUD at `/api/v1/places`

### Security Features
- JWT authentication
- Role-based access control (RBAC)
- Rate limiting (60 requests/minute)
- CORS configuration
- Input validation

## 🎯 Post-Deployment Tasks

### 1. Verify Deployment
```bash
# Check API health
curl https://trip-planner-api.onrender.com/api/health

# Response should be:
# {"status":"ok","timestamp":"...","services":{...}}
```

### 2. Create First Admin User
```bash
curl -X POST https://trip-planner-api.onrender.com/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@example.com",
    "username": "admin",
    "password": "your-secure-password",
    "display_name": "Admin User"
  }'
```

### 3. Set Up Monitoring
1. Enable Render's metrics in the dashboard
2. Set up alerts for service health
3. Configure log retention (7 days free, more with paid plans)

### 4. Custom Domain (Optional)
1. Add custom domain in Render dashboard
2. Update DNS records:
   ```
   Type: CNAME
   Name: app
   Value: trip-planner-web.onrender.com
   ```
3. SSL certificates are automatic

## 🔄 Continuous Deployment

### GitHub Actions Integration
The repository includes GitHub Actions workflows:
- **CI Pipeline**: Runs on every PR
- **Deploy Pipeline**: Auto-deploys on merge to main
- **Security Scans**: Weekly vulnerability scanning

### Manual Deployment
```bash
# From Render dashboard
# 1. Go to your service
# 2. Click "Manual Deploy"
# 3. Select branch and deploy
```

## 📈 Scaling Options

### Starter → Standard Upgrade
When you need more resources:

1. **API Service**: Upgrade to Standard ($25/month)
   - 2GB RAM → 4GB RAM
   - Shared CPU → Dedicated CPU
   - Auto-scaling to 10 instances

2. **Database**: Upgrade to Standard ($30/month)
   - 256MB RAM → 1GB RAM
   - 1GB storage → 16GB storage
   - Daily backups

3. **Redis**: Upgrade to Standard ($30/month)
   - 25MB → 1GB memory
   - Persistence enabled

### Performance Optimization
1. Enable CDN for static assets
2. Use Cloudflare for global distribution
3. Implement database read replicas
4. Add more Redis cache strategies

## 🐛 Troubleshooting

### Common Issues

#### Build Failures
```bash
# Check build logs
# Render Dashboard → Service → Events → Build Logs
```

#### Database Connection Errors
```bash
# Verify DATABASE_URL format
postgres://user:password@host:5432/database?sslmode=require
```

#### Cold Starts
- First request may take 10-30 seconds
- Consider upgrading to eliminate cold starts

### Debug Commands
```bash
# SSH into service (Standard plan+)
render ssh <service-name>

# View logs
render logs <service-name> --tail 100

# Run migrations manually
render run --service api -- migrate up
```

## 📚 Additional Resources

- [Render Documentation](https://render.com/docs)
- [API Documentation](https://trip-planner-api.onrender.com/docs) (when Swagger is added)
- [GitHub Repository](https://github.com/Oferzz/newMap)
- [Mapbox Documentation](https://docs.mapbox.com)

## 💡 Tips

1. **Cost Optimization**
   - Use free tier for staging environments
   - Enable auto-sleep for non-production
   - Monitor usage in billing dashboard

2. **Security**
   - Rotate JWT secrets quarterly
   - Use environment groups for shared secrets
   - Enable 2FA on Render account

3. **Performance**
   - Pre-warm services with health checks
   - Use aggressive caching strategies
   - Optimize images before upload

## 🆘 Support

- Render Support: support@render.com
- GitHub Issues: [Create Issue](https://github.com/Oferzz/newMap/issues)
- Community: [Render Community](https://community.render.com)