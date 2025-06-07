# Supabase Migration Plan

## Executive Summary

This document outlines a comprehensive plan to migrate the NewMap platform from its current PostgreSQL/JWT/Redis architecture to Supabase, leveraging Supabase's built-in authentication, database, and real-time capabilities.

## Current Architecture Overview

- **Database**: PostgreSQL 16 with PostGIS, MongoDB (legacy), Redis (caching)
- **Authentication**: Custom JWT implementation with access/refresh tokens
- **User Management**: Custom user tables with RBAC system
- **Deployment**: Render.com with managed PostgreSQL and Redis

## Migration Benefits

1. **Simplified Architecture**: Replace custom auth with Supabase Auth
2. **Built-in Features**: Row Level Security (RLS), real-time subscriptions, storage
3. **Reduced Maintenance**: No need to manage JWT tokens, refresh logic, or auth middleware
4. **Enhanced Security**: Battle-tested auth system with MFA, OAuth providers
5. **Cost Optimization**: Potentially reduce costs by consolidating services

## Migration Phases

### Phase 1: Setup and Authentication Migration (Week 1-2)

#### 1.1 Supabase Project Setup
- [ ] Create Supabase project
- [ ] Configure environment (production/staging)
- [ ] Set up custom domain (optional)
- [ ] Configure CORS and allowed URLs

#### 1.2 Authentication Planning
- [ ] Map current JWT claims to Supabase auth metadata
- [ ] Plan user migration strategy (passwords can't be migrated directly)
- [ ] Design auth flow for existing users
- [ ] Configure auth providers (email/password initially)

#### 1.3 User Migration Strategy
```sql
-- Supabase auth.users will handle authentication
-- Create profile table for additional user data
CREATE TABLE profiles (
  id UUID REFERENCES auth.users PRIMARY KEY,
  username TEXT UNIQUE,
  display_name TEXT,
  bio TEXT,
  avatar_url TEXT,
  location TEXT,
  website TEXT,
  roles TEXT[],
  privacy_settings JSONB,
  notification_preferences JSONB,
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW()
);
```

### Phase 2: Database Schema Migration (Week 2-3)

#### 2.1 Schema Analysis and Conversion
- [ ] Export current PostgreSQL schema
- [ ] Adapt for Supabase (consider RLS requirements)
- [ ] Handle PostGIS → Supabase PostGIS
- [ ] Plan for MongoDB data migration

#### 2.2 Supabase-Specific Adaptations
```sql
-- Enable RLS on all tables
ALTER TABLE trips ENABLE ROW LEVEL SECURITY;
ALTER TABLE places ENABLE ROW LEVEL SECURITY;
ALTER TABLE trip_collaborators ENABLE ROW LEVEL SECURITY;

-- Example RLS policies
CREATE POLICY "Users can view their own trips" ON trips
  FOR SELECT USING (auth.uid() = owner_id);

CREATE POLICY "Users can edit their own trips" ON trips
  FOR UPDATE USING (auth.uid() = owner_id);

CREATE POLICY "Collaborators can view shared trips" ON trips
  FOR SELECT USING (
    auth.uid() IN (
      SELECT user_id FROM trip_collaborators 
      WHERE trip_id = trips.id
    )
  );
```

#### 2.3 Storage Migration
- [ ] Configure Supabase Storage buckets
- [ ] Plan media file migration from current storage
- [ ] Set up storage policies

### Phase 3: Data Migration Strategy (Week 3-4)

#### 3.1 User Data Migration
```javascript
// Migration script pseudo-code
async function migrateUsers() {
  // 1. Export existing users
  const users = await getExistingUsers();
  
  // 2. Create Supabase auth users
  for (const user of users) {
    // Create auth user (users will need to reset passwords)
    const { data: authUser } = await supabase.auth.admin.createUser({
      email: user.email,
      email_confirm: true,
      user_metadata: {
        username: user.username,
        migrated: true
      }
    });
    
    // 3. Create profile
    await supabase.from('profiles').insert({
      id: authUser.id,
      username: user.username,
      display_name: user.display_name,
      // ... other fields
    });
  }
}
```

#### 3.2 Application Data Migration
- [ ] Export trips, places, collaborators data
- [ ] Transform data for new schema if needed
- [ ] Bulk import to Supabase
- [ ] Verify data integrity

### Phase 4: Backend Code Refactoring (Week 4-6)

#### 4.1 Authentication Refactoring
```go
// Replace JWT utils with Supabase client
package auth

import (
  "github.com/supabase-community/supabase-go"
)

type SupabaseAuth struct {
  client *supabase.Client
}

func (s *SupabaseAuth) ValidateToken(token string) (*User, error) {
  user, err := s.client.Auth.User(token)
  if err != nil {
    return nil, err
  }
  // Map to internal user structure
  return mapSupabaseUser(user), nil
}
```

#### 4.2 Database Access Layer
```go
// Update repository pattern to use Supabase client
type SupabaseRepository struct {
  client *supabase.Client
}

func (r *SupabaseRepository) GetTripByID(id string) (*Trip, error) {
  var trip Trip
  err := r.client.From("trips").
    Select("*").
    Eq("id", id).
    Single().
    Execute(&trip)
  return &trip, err
}
```

#### 4.3 Remove/Replace Components
- [ ] Remove JWT generation/validation code
- [ ] Remove custom password hashing
- [ ] Update middleware to use Supabase auth
- [ ] Remove Redis caching (evaluate Supabase performance first)

### Phase 5: Frontend Integration Updates (Week 5-6)

#### 5.1 Supabase Client Setup
```typescript
// src/services/supabase.ts
import { createClient } from '@supabase/supabase-js'

export const supabase = createClient(
  process.env.VITE_SUPABASE_URL!,
  process.env.VITE_SUPABASE_ANON_KEY!
)
```

#### 5.2 Authentication Service Update
```typescript
// src/services/auth.service.ts
export const authService = {
  async login(email: string, password: string) {
    const { data, error } = await supabase.auth.signInWithPassword({
      email,
      password
    });
    return { data, error };
  },
  
  async logout() {
    return supabase.auth.signOut();
  },
  
  async getSession() {
    return supabase.auth.getSession();
  }
};
```

#### 5.3 Real-time Subscriptions
```typescript
// Replace Socket.io with Supabase real-time
const subscription = supabase
  .channel('trips')
  .on('postgres_changes', 
    { event: '*', schema: 'public', table: 'trips' },
    (payload) => {
      // Handle real-time updates
    }
  )
  .subscribe();
```

### Phase 6: Testing and Rollback Strategy (Week 6-7)

#### 6.1 Testing Plan
- [ ] Unit tests for new Supabase repositories
- [ ] Integration tests for auth flows
- [ ] E2E tests for critical user journeys
- [ ] Performance testing vs current system
- [ ] Security audit of RLS policies

#### 6.2 Rollback Strategy
- [ ] Maintain current system in parallel
- [ ] Feature flags for gradual rollout
- [ ] Data sync mechanism during transition
- [ ] Quick switch capability

### Phase 7: Deployment and Cutover (Week 7-8)

#### 7.1 Deployment Steps
1. Deploy Supabase-integrated backend (feature flagged)
2. Deploy updated frontend
3. Run data migration scripts
4. Enable for internal testing
5. Gradual rollout to users
6. Full cutover
7. Decommission old infrastructure

#### 7.2 Post-Migration
- [ ] Monitor performance metrics
- [ ] Address user feedback
- [ ] Optimize RLS policies
- [ ] Document new architecture

## Technical Considerations

### 1. PostGIS Compatibility
Supabase supports PostGIS, but verify all spatial queries work correctly:
```sql
-- Test spatial queries
SELECT * FROM places 
WHERE ST_DWithin(
  location::geography,
  ST_SetSRID(ST_MakePoint(lng, lat), 4326)::geography,
  radius_meters
);
```

### 2. Performance Optimization
- Implement proper indexes
- Optimize RLS policies
- Consider connection pooling with Supabase

### 3. Cost Analysis
- Supabase pricing tiers
- Data transfer costs
- Storage costs for media
- Compare with current Render.com costs

### 4. Security Considerations
- RLS policy audit
- API key management
- Environment variable updates
- CORS configuration

## Risk Mitigation

1. **Data Loss**: Comprehensive backups before migration
2. **Authentication Issues**: Password reset flow for all users
3. **Performance Degradation**: Thorough testing and optimization
4. **Feature Parity**: Ensure all features work post-migration

## Timeline Summary

- **Week 1-2**: Setup and Authentication
- **Week 2-3**: Schema Migration
- **Week 3-4**: Data Migration
- **Week 4-6**: Code Refactoring
- **Week 5-6**: Frontend Updates
- **Week 6-7**: Testing
- **Week 7-8**: Deployment

## Success Criteria

1. All users successfully migrated
2. No data loss
3. Performance meets or exceeds current system
4. All features functional
5. Reduced operational complexity
6. Cost savings realized

## Next Steps

1. Approval of migration plan
2. Create Supabase project
3. Set up development environment
4. Begin Phase 1 implementation