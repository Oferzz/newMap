# Deployment Checklist for Trip Planning Platform

## Pre-Deployment Steps

### 1. Environment Variables Setup in Render Dashboard
Before deploying, ensure these environment variables are set in the Render dashboard:

#### API Service (newMap-api)
- [ ] `MAPBOX_API_KEY` - Your Mapbox API key
- [ ] `JWT_SECRET` - Will be auto-generated, but you can set a custom one
- [ ] `SENTRY_DSN` - (Optional) For error tracking

#### Web Service (newMap-web)
- [ ] `VITE_MAPBOX_TOKEN` - Your Mapbox public token

### 2. GitHub Repository Setup
- [x] Push code to GitHub repository
- [x] Ensure main branch is protected
- [x] Set up branch protection rules
- [x] Enable GitHub Actions

### 3. Render Account Setup
- [ ] Create a Render account at https://render.com
- [ ] Add payment method (required for PostgreSQL and Redis)
- [ ] Connect GitHub account to Render

## Deployment Steps

### 1. Initial Deployment
1. **Connect Repository**
   - [ ] Go to Render Dashboard
   - [ ] Click "New +" → "Blueprint"
   - [ ] Connect your GitHub repository
   - [ ] Select the repository: `Oferzz/newMap`
   - [ ] Choose branch: `main`

2. **Review Resources**
   - [ ] Verify all services are detected from `render.yaml`
   - [ ] Confirm pricing (approximately $21/month for starter plan)
   - [ ] Review environment variables

3. **Deploy**
   - [ ] Click "Apply" to create all resources
   - [ ] Wait for initial deployment (10-15 minutes)

### 2. Post-Deployment Configuration

1. **Database Setup**
   - [ ] Verify PostgreSQL is running
   - [ ] Check migrations have run successfully
   - [ ] Enable PostGIS extension (should be automatic)

2. **Redis Configuration**
   - [ ] Verify Redis is accessible
   - [ ] Check connection from API service

3. **API Service**
   - [ ] Test health endpoint: `https://newMap-api.onrender.com/api/health`
   - [ ] Verify JWT authentication is working
   - [ ] Test API endpoints with Postman/curl

4. **Frontend Service**
   - [ ] Access the web app
   - [ ] Verify Mapbox is loading
   - [ ] Test API connectivity
   - [ ] Check responsive design

### 3. Domain Configuration (Optional)
1. **Add Custom Domain**
   - [ ] Add custom domain in Render dashboard
   - [ ] Update DNS records
   - [ ] Enable auto-renew SSL

2. **Update Environment Variables**
   - [ ] Update `VITE_API_URL` to use custom domain
   - [ ] Update CORS settings if needed

## Monitoring and Maintenance

### 1. Set Up Monitoring
- [ ] Enable Render's built-in metrics
- [ ] Set up alerts for downtime
- [ ] Configure log retention
- [ ] Set up Sentry for error tracking (optional)

### 2. Regular Maintenance
- [ ] Monitor disk usage for media storage
- [ ] Review Redis memory usage
- [ ] Check database performance
- [ ] Review monthly costs

### 3. Backup Strategy
- [ ] Enable automatic PostgreSQL backups
- [ ] Set up media backup to external storage (S3/Cloudinary)
- [ ] Document recovery procedures

## Troubleshooting

### Common Issues and Solutions

1. **Build Failures**
   - Check build logs in Render dashboard
   - Verify Node.js and Go versions
   - Check for missing dependencies

2. **Database Connection Issues**
   - Verify DATABASE_URL is correctly set
   - Check PostgreSQL logs
   - Ensure migrations completed

3. **Media Upload Issues**
   - Verify disk is mounted correctly
   - Check file permissions
   - Monitor disk space

4. **Performance Issues**
   - Scale up instances if needed
   - Review Redis cache hit rates
   - Optimize database queries

## Security Checklist

- [ ] All secrets are stored as environment variables
- [ ] HTTPS is enforced
- [ ] CORS is properly configured
- [ ] Rate limiting is enabled
- [ ] Input validation is working
- [ ] File upload restrictions are in place

## Final Verification

- [ ] All health checks passing
- [ ] User registration and login working
- [ ] Trip creation and management functional
- [ ] Map features working correctly
- [ ] Media upload and serving functional
- [ ] Real-time features operational
- [ ] Mobile responsive design verified

## Support Resources

- Render Documentation: https://render.com/docs
- Render Status Page: https://status.render.com
- GitHub Repository: https://github.com/Oferzz/newMap
- API Documentation: (Add Swagger URL when available)

## Notes

- Initial deployment may take 10-15 minutes
- First request after deployment may be slow (cold start)
- Monitor costs in Render dashboard
- Set up budget alerts to avoid surprises