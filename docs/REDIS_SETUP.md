# Redis Setup for Render Deployment

This document explains how to configure Redis caching for the newMap application on Render.

## Redis Configuration

The application is configured to use Redis for caching to improve performance. Redis is optional - the application will gracefully fall back to in-memory caching if Redis is not available.

### Environment Variables

The application checks for Redis configuration in this order:

1. `REDIS_URL` - Custom Redis connection string
2. `INTERNAL_REDIS_URL` - Render's managed Redis (recommended)
3. Falls back to local Redis at `redis://localhost:6379`

### Render Redis Add-on Setup

To enable Redis caching on Render:

1. **Add Redis Service**
   - Go to your Render dashboard
   - Click "New +" and select "Redis"
   - Choose your plan (Free tier available)
   - Name it (e.g., "newmap-redis")
   - Click "Create Redis"

2. **Connect to Your Web Service**
   - Go to your web service settings
   - Navigate to "Environment" tab
   - Render will automatically provide `INTERNAL_REDIS_URL`
   - No manual configuration needed

3. **Verify Connection**
   - Check application logs for "Redis connected, caching enabled"
   - If Redis is unavailable, you'll see "Redis not available, using no-op cache"

## Redis Usage in Application

### Caching Strategy

The application uses Redis for:

1. **Trip Data Caching**
   - Frequently accessed trips
   - User trip lists
   - Public trip feeds

2. **Search Result Caching**
   - NLP search queries
   - Elasticsearch results
   - Autocomplete suggestions

3. **Session Data**
   - User authentication tokens
   - Rate limiting counters

### Cache Keys

The application uses a structured cache key naming convention:

```
trips:{user_id}:list           # User's trip list
trips:{trip_id}:details        # Individual trip data
search:{hash}:results          # Search results
places:{area}:public           # Public places in area
users:{user_id}:permissions    # User permissions cache
```

### TTL (Time To Live)

Different data types have different cache durations:

- **Trip Lists**: 5 minutes
- **Trip Details**: 15 minutes
- **Search Results**: 1 hour
- **User Permissions**: 30 minutes
- **Public Data**: 2 hours

## Performance Benefits

With Redis enabled, you can expect:

- **Faster Trip Loading**: 2-3x improvement for cached trips
- **Better Search Performance**: Sub-second response for cached queries
- **Reduced Database Load**: 40-60% reduction in PostgreSQL queries
- **Improved Concurrent Users**: Better handling of simultaneous requests

## Monitoring Redis

### Application Logs

Monitor these log messages:

```bash
# Successful connection
"Redis connected, caching enabled"

# Connection failure (will use fallback)
"Redis not available, using no-op cache"

# Cache statistics (debug mode)
"Cache hit rate: 75% (150/200 requests)"
```

### Render Dashboard

In your Redis service dashboard, monitor:

- **Memory Usage**: Should stay under your plan limits
- **Commands/sec**: Indicates cache activity
- **Hit Rate**: Higher is better (>70% is good)

## Troubleshooting

### Redis Connection Issues

1. **Check Environment Variables**
   ```bash
   # In Render shell
   echo $INTERNAL_REDIS_URL
   ```

2. **Verify Redis Service Status**
   - Check Render Redis dashboard
   - Ensure service is running
   - Check for any service alerts

3. **Application Logs**
   ```bash
   # Look for Redis-related messages
   grep -i redis /var/log/app.log
   ```

### Performance Issues

1. **High Memory Usage**
   - Check cache key patterns
   - Consider reducing TTL values
   - Monitor for memory leaks

2. **Low Hit Rate**
   - Review caching strategy
   - Check for cache invalidation issues
   - Consider warming cache on startup

### Cache Invalidation

The application automatically invalidates cache for:

- **Trip Updates**: Clears trip-specific cache
- **User Changes**: Clears user-specific cache
- **Search Index Updates**: Clears search cache

## Development vs Production

### Local Development

For local development without Redis:

```bash
# Optional: Run Redis locally
docker run -p 6379:6379 redis:alpine

# Or just run without Redis (uses in-memory cache)
go run cmd/server/main.go
```

### Production Recommendations

1. **Use Render's Managed Redis**: Simpler maintenance
2. **Monitor Cache Performance**: Set up alerts for hit rates
3. **Plan for Growth**: Upgrade Redis plan as usage increases
4. **Backup Strategy**: Render handles Redis backups automatically

## Cost Optimization

### Free Tier Usage

Render's free Redis tier includes:
- 25MB storage
- Suitable for small applications
- No persistence (data may be lost on restart)

### Upgrade Considerations

Consider upgrading when:
- Memory usage consistently >80%
- Need data persistence
- Require higher performance
- Multiple environments (staging/prod)

## Security

### Access Control

- Redis is only accessible within Render's private network
- No external access by default
- Use strong passwords if configuring custom Redis

### Data Sensitivity

The application caches:
- ✅ Public trip data
- ✅ Search results
- ✅ User permissions
- ❌ Sensitive user data (passwords, tokens)
- ❌ Private messages
- ❌ Payment information

## Future Enhancements

Planned Redis usage improvements:

1. **Real-time Features**: Use Redis pub/sub for live updates
2. **Advanced Caching**: Implement cache warming strategies
3. **Analytics**: Store usage analytics in Redis
4. **Session Management**: Distribute sessions across instances