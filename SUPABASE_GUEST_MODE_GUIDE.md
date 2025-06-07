# Supabase Guest Mode Implementation Guide

## Overview

This guide details how to implement the guest mode (anonymous users) feature using Supabase while maintaining the current freemium model where users can use core features without signing up.

## Current Guest Mode vs Supabase Anonymous Users

### Current Implementation
- Data stored in browser localStorage/IndexedDB
- No server-side storage for guest users
- Data migrated to cloud when user signs up

### Supabase Anonymous Users
- Creates a real user in auth.users with no email/identity
- Can store data server-side with RLS protection
- Can be upgraded to permanent user later
- Maintains same user ID after upgrade

## Implementation Strategy

### 1. Anonymous User Creation

```typescript
// src/hooks/useAuth.ts
export const useAuth = () => {
  const [user, setUser] = useState<User | null>(null)
  const [isGuest, setIsGuest] = useState(false)
  
  useEffect(() => {
    // Check for existing session
    supabase.auth.getSession().then(({ data: { session } }) => {
      if (session) {
        setUser(session.user)
        setIsGuest(!session.user.email) // Anonymous if no email
      } else {
        // Create anonymous session for new users
        createGuestSession()
      }
    })
    
    // Listen for auth changes
    const { data: { subscription } } = supabase.auth.onAuthStateChange(
      (event, session) => {
        setUser(session?.user ?? null)
        setIsGuest(session?.user && !session.user.email)
      }
    )
    
    return () => subscription.unsubscribe()
  }, [])
  
  const createGuestSession = async () => {
    const { data, error } = await supabase.auth.signInAnonymously()
    if (data.user) {
      setUser(data.user)
      setIsGuest(true)
    }
  }
  
  return { user, isGuest, isAuthenticated: !!user }
}
```

### 2. Data Service Updates

```typescript
// src/services/storage/supabaseDataService.ts
export class SupabaseDataService implements DataService {
  private supabase: SupabaseClient
  
  constructor() {
    this.supabase = createClient(
      import.meta.env.VITE_SUPABASE_URL,
      import.meta.env.VITE_SUPABASE_ANON_KEY
    )
  }
  
  async saveTrip(trip: Trip): Promise<Trip> {
    const { data: { user } } = await this.supabase.auth.getUser()
    if (!user) throw new Error('No authenticated user')
    
    const { data, error } = await this.supabase
      .from('trips')
      .insert({
        ...trip,
        owner_id: user.id,
        // Mark as guest data for potential cleanup
        is_guest_data: !user.email
      })
      .select()
      .single()
    
    if (error) throw error
    return data
  }
  
  async getTrips(): Promise<Trip[]> {
    const { data: { user } } = await this.supabase.auth.getUser()
    if (!user) return []
    
    const { data, error } = await this.supabase
      .from('trips')
      .select('*')
      .eq('owner_id', user.id)
      .order('created_at', { ascending: false })
    
    if (error) throw error
    return data || []
  }
}
```

### 3. Guest to User Upgrade

```typescript
// src/components/auth/UpgradeAccount.tsx
export const UpgradeAccount: React.FC = () => {
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [loading, setLoading] = useState(false)
  
  const handleUpgrade = async (e: React.FormEvent) => {
    e.preventDefault()
    setLoading(true)
    
    try {
      // Update anonymous user to permanent user
      const { error } = await supabase.auth.updateUser({
        email,
        password
      })
      
      if (error) throw error
      
      // Send confirmation email
      await supabase.auth.resend({
        type: 'signup',
        email
      })
      
      toast.success('Account upgraded! Check your email to confirm.')
    } catch (error) {
      toast.error('Failed to upgrade account')
    } finally {
      setLoading(false)
    }
  }
  
  return (
    <form onSubmit={handleUpgrade}>
      <h2>Save Your Data Permanently</h2>
      <p>Create an account to access your data from any device</p>
      
      <input
        type="email"
        value={email}
        onChange={(e) => setEmail(e.target.value)}
        placeholder="Email"
        required
      />
      
      <input
        type="password"
        value={password}
        onChange={(e) => setPassword(e.target.value)}
        placeholder="Password"
        required
      />
      
      <button type="submit" disabled={loading}>
        Create Account
      </button>
    </form>
  )
}
```

### 4. Database Schema Updates

```sql
-- Add guest data tracking
ALTER TABLE trips ADD COLUMN is_guest_data BOOLEAN DEFAULT false;
ALTER TABLE places ADD COLUMN is_guest_data BOOLEAN DEFAULT false;

-- Index for guest data queries
CREATE INDEX idx_trips_guest_data ON trips(owner_id, is_guest_data);
CREATE INDEX idx_places_guest_data ON places(owner_id, is_guest_data);

-- Function to mark data as permanent after email confirmation
CREATE OR REPLACE FUNCTION mark_user_data_permanent()
RETURNS TRIGGER AS $$
BEGIN
  -- When user email is confirmed, mark their data as permanent
  IF NEW.email_confirmed_at IS NOT NULL AND OLD.email_confirmed_at IS NULL THEN
    UPDATE trips SET is_guest_data = false WHERE owner_id = NEW.id;
    UPDATE places SET is_guest_data = false WHERE owner_id = NEW.id;
  END IF;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER on_user_email_confirmed
  AFTER UPDATE ON auth.users
  FOR EACH ROW
  WHEN (NEW.email_confirmed_at IS NOT NULL AND OLD.email_confirmed_at IS NULL)
  EXECUTE FUNCTION mark_user_data_permanent();
```

### 5. RLS Policies for Guest Users

```sql
-- Allow anonymous users to manage their own data
CREATE POLICY "Anonymous users can manage own trips"
  ON trips
  FOR ALL
  USING (auth.uid() = owner_id)
  WITH CHECK (auth.uid() = owner_id);

-- Prevent anonymous users from seeing others' data
CREATE POLICY "Anonymous users cannot see public trips"
  ON trips
  FOR SELECT
  USING (
    auth.uid() = owner_id OR 
    (visibility = 'public' AND auth.jwt()->>'email' IS NOT NULL)
  );
```

### 6. Guest Data Cleanup

```sql
-- Function to clean up old guest data
CREATE OR REPLACE FUNCTION cleanup_old_guest_data()
RETURNS void AS $$
BEGIN
  -- Delete guest data older than 30 days
  DELETE FROM trips 
  WHERE is_guest_data = true 
    AND created_at < NOW() - INTERVAL '30 days';
    
  DELETE FROM places 
  WHERE is_guest_data = true 
    AND created_at < NOW() - INTERVAL '30 days';
    
  -- Delete anonymous users with no data
  DELETE FROM auth.users
  WHERE email IS NULL
    AND created_at < NOW() - INTERVAL '30 days'
    AND id NOT IN (SELECT DISTINCT owner_id FROM trips)
    AND id NOT IN (SELECT DISTINCT owner_id FROM places);
END;
$$ LANGUAGE plpgsql;

-- Schedule cleanup (using pg_cron or external scheduler)
SELECT cron.schedule('cleanup-guest-data', '0 3 * * *', 'SELECT cleanup_old_guest_data()');
```

### 7. Local Storage Fallback

```typescript
// src/services/storage/hybridDataService.ts
export class HybridDataService implements DataService {
  private localService: LocalDataService
  private supabaseService: SupabaseDataService
  
  constructor() {
    this.localService = new LocalDataService()
    this.supabaseService = new SupabaseDataService()
  }
  
  async saveTrip(trip: Trip): Promise<Trip> {
    try {
      // Try Supabase first
      const savedTrip = await this.supabaseService.saveTrip(trip)
      // Sync to local as backup
      await this.localService.saveTrip(savedTrip)
      return savedTrip
    } catch (error) {
      // Fallback to local only
      console.warn('Falling back to local storage', error)
      return this.localService.saveTrip(trip)
    }
  }
  
  async getTrips(): Promise<Trip[]> {
    try {
      // Get from Supabase
      const trips = await this.supabaseService.getTrips()
      // Sync to local
      await this.localService.syncTrips(trips)
      return trips
    } catch (error) {
      // Fallback to local
      console.warn('Using local data', error)
      return this.localService.getTrips()
    }
  }
}
```

### 8. UI Updates for Guest Mode

```typescript
// src/components/common/GuestModeBanner.tsx
export const GuestModeBanner: React.FC = () => {
  const { isGuest } = useAuth()
  const [dismissed, setDismissed] = useState(false)
  
  if (!isGuest || dismissed) return null
  
  return (
    <div className="bg-yellow-50 border-b border-yellow-200 px-4 py-3">
      <div className="flex items-center justify-between">
        <div className="flex items-center">
          <AlertCircle className="h-5 w-5 text-yellow-600 mr-2" />
          <p className="text-sm text-yellow-800">
            You're using NewMap as a guest. Your data is temporary and will be deleted after 30 days.
          </p>
        </div>
        <div className="flex items-center space-x-4">
          <Link
            to="/upgrade"
            className="text-sm font-medium text-yellow-800 hover:text-yellow-900"
          >
            Create Account
          </Link>
          <button
            onClick={() => setDismissed(true)}
            className="text-yellow-600 hover:text-yellow-700"
          >
            <X className="h-4 w-4" />
          </button>
        </div>
      </div>
    </div>
  )
}
```

### 9. Migration from Current System

```typescript
// src/services/migration/guestDataMigration.ts
export async function migrateLocalDataToSupabase() {
  const localService = new LocalDataService()
  const supabase = createClient(...)
  
  // Get all local data
  const localTrips = await localService.getTrips()
  const localPlaces = await localService.getPlaces()
  
  // Create anonymous session if needed
  const { data: { user } } = await supabase.auth.getUser()
  if (!user) {
    await supabase.auth.signInAnonymously()
  }
  
  // Migrate trips
  for (const trip of localTrips) {
    await supabase.from('trips').insert({
      ...trip,
      owner_id: user.id,
      is_guest_data: !user.email
    })
  }
  
  // Migrate places
  for (const place of localPlaces) {
    await supabase.from('places').insert({
      ...place,
      owner_id: user.id,
      is_guest_data: !user.email
    })
  }
  
  // Clear local storage after successful migration
  await localService.clear()
}
```

## Benefits of This Approach

1. **Seamless Experience**: Users don't need to sign up to start using the app
2. **Data Persistence**: Guest data is stored server-side with 30-day retention
3. **Easy Upgrade**: Anonymous users can upgrade without losing data
4. **Security**: RLS policies protect guest data
5. **Scalability**: Server-side storage for all users
6. **Offline Support**: Local storage fallback when offline

## Considerations

1. **Storage Costs**: Anonymous users will use database storage
2. **Cleanup Policy**: Implement regular cleanup of old guest data
3. **Rate Limiting**: Consider limits for anonymous users
4. **Analytics**: Track guest-to-user conversion rates
5. **GDPR**: Ensure compliance with data retention policies