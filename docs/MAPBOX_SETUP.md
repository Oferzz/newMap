# Mapbox API Setup for newMap

## Backend API Token Configuration

The backend API uses Mapbox Geocoding API to search for places. This requires a **secret access token** with the correct scopes.

### Creating the Correct Token

1. Go to [Mapbox Access Tokens](https://account.mapbox.com/access-tokens/)
2. Click "Create a token"
3. Give it a name like "newMap Backend API"
4. **Important**: Under "Secret scopes", check:
   - `styles:read`
   - `geocoding:read` (if available)
   - Or use "All public scopes" if geocoding:read is not listed
5. **URL restrictions**: Leave this EMPTY or add `*` (all URLs)
   - URL restrictions are for frontend use only
   - Backend API calls don't have a referrer URL
6. Click "Create token"

### Common Issues

#### 403 Forbidden Error
- **Cause**: Using a public token or a token with URL restrictions
- **Solution**: Create a secret token without URL restrictions

#### Token Types
- **Public tokens**: For frontend use (has URL restrictions)
- **Secret tokens**: For backend use (no URL restrictions)

### Setting the Token in Render

1. Go to your Render dashboard
2. Navigate to the newMap-api service
3. Go to "Environment" tab
4. Find `MAPBOX_API_KEY` (or add it)
5. Set the value to your secret access token
6. Save and the service will redeploy

### Testing the Token

You can test if your token works using curl:

```bash
curl "https://api.mapbox.com/geocoding/v5/mapbox.places/Tokyo.json?access_token=YOUR_TOKEN_HERE"
```

If you get a 403, the token doesn't have the right permissions.

## Frontend Token Configuration

The frontend uses a different token (public token) set in `VITE_MAPBOX_TOKEN`. This token should have:
- URL restrictions set to your frontend domains
- Only needs basic public scopes

## Security Notes

- Never commit tokens to git
- Use environment variables
- Backend tokens should be secret tokens
- Frontend tokens should be public tokens with URL restrictions