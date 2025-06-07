# Supabase Migration Plan - Updated

## Key Updates Based on Supabase Documentation

After reviewing Supabase's actual implementation, here are the key adjustments to our migration plan:

### 1. Authentication Adjustments

#### 1.1 User Types
- **Permanent Users**: Tied to email/phone/third-party identity (our main users)
- **Anonymous Users**: Have user ID but no identities (useful for guest mode)
- Anonymous users use `authenticated` role for RLS, not `anon` role

#### 1.2 JWT Implementation
- Supabase uses JSON Web Tokens (JWTs) automatically
- Integrates with Row Level Security (RLS) seamlessly
- Access tokens are managed by Supabase client SDK

#### 1.3 User Metadata Storage
```typescript
// User signup with metadata
const { data, error } = await supabase.auth.signUp({
  email: 'user@example.com',
  password: 'password',
  options: {
    data: {
      username: 'johndoe',
      display_name: 'John Doe',
      // Additional metadata
    }
  }
})
```

### 2. Updated Authentication Flow

#### 2.1 Guest Mode Implementation
```typescript
// Create anonymous user for guest mode
const { data, error } = await supabase.auth.signInAnonymously()

// Later convert to permanent user
const { data, error } = await supabase.auth.updateUser({
  email: 'user@example.com',
  password: 'password'
})
```

#### 2.2 Email Confirmation Settings
- By default, users need to verify email before logging in
- Can disable with `Confirm email` setting in project dashboard
- When disabled, both `user` and `session` are returned immediately

### 3. Updated Database Schema

#### 3.1 User Profile Trigger (Corrected)
```sql
-- Function to handle new user creation
CREATE OR REPLACE FUNCTION public.handle_new_user()
RETURNS trigger AS $$
BEGIN
  INSERT INTO public.profiles (id, username, display_name, email)
  VALUES (
    new.id,
    COALESCE(new.raw_user_meta_data->>'username', split_part(new.email, '@', 1)),
    COALESCE(new.raw_user_meta_data->>'display_name', split_part(new.email, '@', 1)),
    new.email
  );
  RETURN new;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Trigger on auth.users
CREATE TRIGGER on_auth_user_created
  AFTER INSERT ON auth.users
  FOR EACH ROW EXECUTE FUNCTION public.handle_new_user();
```

#### 3.2 RLS Policies (Updated)
```sql
-- Anonymous users can create temporary data
CREATE POLICY "Authenticated users can create places"
  ON places FOR INSERT
  WITH CHECK (auth.role() = 'authenticated');

-- Both permanent and anonymous users can view their data
CREATE POLICY "Users can view own places"
  ON places FOR SELECT
  USING (auth.uid() = owner_id);
```

### 4. Frontend Updates

#### 4.1 Supabase Client Configuration
```typescript
import { createClient } from '@supabase/supabase-js'

// Your project details from dashboard
const SUPABASE_URL = 'https://xrzjkhivkbcjdfirunyz.supabase.co'
const SUPABASE_ANON_KEY = 'your-anon-key-from-dashboard'

export const supabase = createClient(SUPABASE_URL, SUPABASE_ANON_KEY, {
  auth: {
    autoRefreshToken: true,
    persistSession: true,
    detectSessionInUrl: true
  }
})
```

#### 4.2 Authentication State Management
```typescript
// Listen to auth state changes
supabase.auth.onAuthStateChange((event, session) => {
  if (event === 'SIGNED_IN') {
    // Handle sign in
  } else if (event === 'SIGNED_OUT') {
    // Handle sign out
  } else if (event === 'USER_UPDATED') {
    // Handle user metadata updates
  }
})

// Get current user
const { data: { user } } = await supabase.auth.getUser()
```

### 5. Migration Strategy Updates

#### 5.1 User Migration Script (Updated)
```typescript
import { createClient } from '@supabase/supabase-js'

const supabase = createClient(
  'https://xrzjkhivkbcjdfirunyz.supabase.co',
  process.env.SUPABASE_SERVICE_KEY!, // Service key for admin operations
  {
    auth: {
      autoRefreshToken: false,
      persistSession: false
    }
  }
)

async function migrateUsers() {
  const oldUsers = await getOldUsers()
  
  for (const user of oldUsers) {
    try {
      // Create user without requiring email confirmation
      const { data, error } = await supabase.auth.admin.createUser({
        email: user.email,
        email_confirm: true, // Skip email confirmation
        user_metadata: {
          username: user.username,
          display_name: user.display_name,
          migrated: true,
          old_user_id: user.id
        }
      })
      
      if (error) throw error
      
      // Profile will be created automatically via trigger
      console.log(`Migrated user: ${user.email}`)
      
    } catch (error) {
      console.error(`Failed to migrate user ${user.email}:`, error)
    }
  }
}
```

#### 5.2 Password Reset for Migrated Users
```typescript
// Send password reset email
const { error } = await supabase.auth.resetPasswordForEmail(email, {
  redirectTo: 'https://newmap-qojk.onrender.com/auth/reset-password',
})
```

### 6. API Integration

#### 6.1 Backend Supabase Client
```go
// Use official Supabase Go client
import (
    "github.com/supabase-community/supabase-go"
)

func NewSupabaseClient() *supabase.Client {
    url := "https://xrzjkhivkbcjdfirunyz.supabase.co"
    key := os.Getenv("SUPABASE_SERVICE_KEY") // Use service key for backend
    
    client, err := supabase.NewClient(url, key, &supabase.ClientOptions{})
    if err != nil {
        panic(err)
    }
    return client
}
```

#### 6.2 Verify JWT in Backend
```go
// Middleware to verify Supabase JWT
func SupabaseAuthMiddleware(client *supabase.Client) gin.HandlerFunc {
    return func(c *gin.Context) {
        token := c.GetHeader("Authorization")
        if token == "" {
            c.AbortWithStatusJSON(401, gin.H{"error": "No authorization header"})
            return
        }
        
        // Remove "Bearer " prefix
        token = strings.TrimPrefix(token, "Bearer ")
        
        // Verify token with Supabase
        user, err := client.Auth.User(context.Background(), token)
        if err != nil {
            c.AbortWithStatusJSON(401, gin.H{"error": "Invalid token"})
            return
        }
        
        c.Set("user_id", user.ID)
        c.Set("user", user)
        c.Next()
    }
}
```

### 7. Environment Variables

#### 7.1 Frontend (.env)
```bash
# Public anon key (safe to expose)
VITE_SUPABASE_URL=https://xrzjkhivkbcjdfirunyz.supabase.co
VITE_SUPABASE_ANON_KEY=your-anon-key-from-api-settings
```

#### 7.2 Backend (.env)
```bash
# Service role key (keep secret!)
SUPABASE_URL=https://xrzjkhivkbcjdfirunyz.supabase.co
SUPABASE_SERVICE_KEY=your-service-role-key-from-api-settings
```

### 8. Real-time Subscriptions

```typescript
// Subscribe to changes
const channel = supabase
  .channel('db-changes')
  .on(
    'postgres_changes',
    {
      event: '*',
      schema: 'public',
      table: 'trips',
      filter: `owner_id=eq.${user.id}`
    },
    (payload) => {
      console.log('Change received!', payload)
    }
  )
  .subscribe()

// Cleanup
channel.unsubscribe()
```

### 9. Storage Integration

```typescript
// Upload file
const { data, error } = await supabase.storage
  .from('avatars')
  .upload(`${user.id}/avatar.png`, file)

// Get public URL
const { data } = supabase.storage
  .from('avatars')
  .getPublicUrl(`${user.id}/avatar.png`)
```

### 10. Testing Considerations

#### 10.1 Local Development
- Use Supabase CLI for local development
- `npx supabase init`
- `npx supabase start` - Starts local Supabase instance
- `npx supabase db reset` - Reset local database

#### 10.2 Testing RLS Policies
```sql
-- Test RLS as different users
SET request.jwt.claims = '{"sub": "user-id-1"}';
SELECT * FROM trips; -- Should only see user-id-1's trips

SET request.jwt.claims = '{"sub": "user-id-2"}';
SELECT * FROM trips; -- Should only see user-id-2's trips
```

## Project-Specific Details

- **Project URL**: https://xrzjkhivkbcjdfirunyz.supabase.co
- **Project ID**: xrzjkhivkbcjdfirunyz
- **Region**: Likely US East (check dashboard)
- **API Keys**: Available in project settings

## Next Steps

1. Get API keys from Supabase dashboard
2. Set up local development with Supabase CLI
3. Create initial schema with RLS policies
4. Implement authentication in a feature branch
5. Test thoroughly before migrating production data