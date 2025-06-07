# Supabase Authentication Migration Guide

## Overview
This document provides detailed implementation steps for migrating from the current JWT-based authentication to Supabase Auth.

## Current Authentication Flow

### Current JWT Implementation
```go
// Current JWT token structure
type Claims struct {
    UserID string `json:"user_id"`
    Email  string `json:"email"`
    jwt.RegisteredClaims
}

// Current login flow
1. User submits email/password
2. Backend validates credentials against PostgreSQL
3. Generate JWT access token (15 min) and refresh token (7 days)
4. Return tokens to client
5. Client stores tokens and includes in Authorization header
```

## Supabase Authentication Implementation

### 1. Backend Migration

#### 1.1 Supabase Client Setup
```go
// internal/auth/supabase_client.go
package auth

import (
    "context"
    "fmt"
    "github.com/supabase-community/supabase-go"
)

type SupabaseClient struct {
    client *supabase.Client
}

func NewSupabaseClient(url, serviceKey string) (*SupabaseClient, error) {
    client, err := supabase.NewClient(url, serviceKey, nil)
    if err != nil {
        return nil, fmt.Errorf("failed to create supabase client: %w", err)
    }
    return &SupabaseClient{client: client}, nil
}
```

#### 1.2 Updated Auth Middleware
```go
// internal/middleware/supabase_auth.go
package middleware

import (
    "net/http"
    "strings"
    
    "github.com/gin-gonic/gin"
    "github.com/supabase-community/supabase-go"
)

func SupabaseAuth(client *supabase.Client) gin.HandlerFunc {
    return func(c *gin.Context) {
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
            c.Abort()
            return
        }
        
        token := strings.Replace(authHeader, "Bearer ", "", 1)
        
        // Verify token with Supabase
        user, err := client.Auth.User(context.Background(), token)
        if err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
            c.Abort()
            return
        }
        
        // Set user context
        c.Set("user_id", user.ID)
        c.Set("email", user.Email)
        c.Set("user_metadata", user.UserMetadata)
        
        c.Next()
    }
}
```

#### 1.3 User Service Migration
```go
// internal/domain/users/service_supabase.go
package users

import (
    "context"
    "fmt"
    
    "github.com/supabase-community/supabase-go"
)

type SupabaseUserService struct {
    client *supabase.Client
    repo   Repository
}

func (s *SupabaseUserService) Login(ctx context.Context, email, password string) (*LoginResponse, error) {
    // Use Supabase Auth
    authResp, err := s.client.Auth.SignInWithPassword(ctx, supabase.UserCredentials{
        Email:    email,
        Password: password,
    })
    if err != nil {
        return nil, fmt.Errorf("login failed: %w", err)
    }
    
    // Get user profile from our database
    profile, err := s.repo.GetByAuthID(ctx, authResp.User.ID)
    if err != nil {
        return nil, fmt.Errorf("failed to get user profile: %w", err)
    }
    
    return &LoginResponse{
        User:         profile,
        AccessToken:  authResp.AccessToken,
        RefreshToken: authResp.RefreshToken,
        ExpiresIn:    authResp.ExpiresIn,
    }, nil
}

func (s *SupabaseUserService) Register(ctx context.Context, req *RegisterRequest) (*User, error) {
    // Create Supabase auth user
    authResp, err := s.client.Auth.SignUp(ctx, supabase.UserCredentials{
        Email:    req.Email,
        Password: req.Password,
        Data: map[string]interface{}{
            "username":     req.Username,
            "display_name": req.DisplayName,
        },
    })
    if err != nil {
        return nil, fmt.Errorf("signup failed: %w", err)
    }
    
    // Create user profile in our database
    user := &User{
        ID:          authResp.User.ID,
        Email:       req.Email,
        Username:    req.Username,
        DisplayName: req.DisplayName,
    }
    
    if err := s.repo.Create(ctx, user); err != nil {
        // Rollback: delete auth user if profile creation fails
        _ = s.client.Auth.AdminDeleteUser(ctx, authResp.User.ID)
        return nil, fmt.Errorf("failed to create user profile: %w", err)
    }
    
    return user, nil
}
```

### 2. Frontend Migration

#### 2.1 Supabase Client Configuration
```typescript
// src/lib/supabase.ts
import { createClient } from '@supabase/supabase-js'
import type { Database } from './database.types'

const supabaseUrl = import.meta.env.VITE_SUPABASE_URL
const supabaseAnonKey = import.meta.env.VITE_SUPABASE_ANON_KEY

export const supabase = createClient<Database>(supabaseUrl, supabaseAnonKey, {
  auth: {
    persistSession: true,
    autoRefreshToken: true,
    detectSessionInUrl: true
  }
})
```

#### 2.2 Updated Auth Service
```typescript
// src/services/auth.service.ts
import { supabase } from '../lib/supabase'

export const authService = {
  async login(email: string, password: string) {
    const { data, error } = await supabase.auth.signInWithPassword({
      email,
      password,
    })
    
    if (error) throw error
    
    // Fetch user profile
    const { data: profile } = await supabase
      .from('profiles')
      .select('*')
      .eq('id', data.user.id)
      .single()
    
    return {
      user: data.user,
      profile,
      session: data.session
    }
  },

  async register(email: string, password: string, username: string) {
    const { data, error } = await supabase.auth.signUp({
      email,
      password,
      options: {
        data: { username }
      }
    })
    
    if (error) throw error
    
    return data
  },

  async logout() {
    const { error } = await supabase.auth.signOut()
    if (error) throw error
  },

  async refreshSession() {
    const { data, error } = await supabase.auth.refreshSession()
    if (error) throw error
    return data
  },

  onAuthStateChange(callback: (event: any, session: any) => void) {
    return supabase.auth.onAuthStateChange(callback)
  }
}
```

#### 2.3 Updated Redux Auth Slice
```typescript
// src/store/slices/authSlice.ts
import { createSlice, PayloadAction } from '@reduxjs/toolkit'
import { User, Session } from '@supabase/supabase-js'

interface AuthState {
  user: User | null
  session: Session | null
  profile: UserProfile | null
  isAuthenticated: boolean
  isLoading: boolean
}

const authSlice = createSlice({
  name: 'auth',
  initialState: {
    user: null,
    session: null,
    profile: null,
    isAuthenticated: false,
    isLoading: true
  } as AuthState,
  reducers: {
    setAuth: (state, action: PayloadAction<{ user: User; session: Session; profile: UserProfile }>) => {
      state.user = action.payload.user
      state.session = action.payload.session
      state.profile = action.payload.profile
      state.isAuthenticated = true
      state.isLoading = false
    },
    clearAuth: (state) => {
      state.user = null
      state.session = null
      state.profile = null
      state.isAuthenticated = false
      state.isLoading = false
    },
    setLoading: (state, action: PayloadAction<boolean>) => {
      state.isLoading = action.payload
    }
  }
})
```

### 3. User Migration Strategy

#### 3.1 Migration Script
```typescript
// scripts/migrate-users.ts
import { createClient } from '@supabase/supabase-js'
import { Pool } from 'pg'

const supabaseAdmin = createClient(
  process.env.SUPABASE_URL!,
  process.env.SUPABASE_SERVICE_KEY!
)

const pgPool = new Pool({
  connectionString: process.env.OLD_DATABASE_URL
})

async function migrateUsers() {
  const { rows: users } = await pgPool.query('SELECT * FROM users')
  
  for (const user of users) {
    try {
      // Create auth user
      const { data: authUser, error } = await supabaseAdmin.auth.admin.createUser({
        email: user.email,
        email_confirm: true,
        user_metadata: {
          username: user.username,
          display_name: user.display_name,
          migrated_at: new Date().toISOString(),
          old_user_id: user.id
        }
      })
      
      if (error) {
        console.error(`Failed to migrate user ${user.email}:`, error)
        continue
      }
      
      // Create profile
      const { error: profileError } = await supabaseAdmin
        .from('profiles')
        .insert({
          id: authUser.user.id,
          username: user.username,
          display_name: user.display_name,
          bio: user.bio,
          avatar_url: user.avatar_url,
          roles: user.roles,
          created_at: user.created_at,
          updated_at: user.updated_at
        })
      
      if (profileError) {
        console.error(`Failed to create profile for ${user.email}:`, profileError)
      }
      
      console.log(`Migrated user: ${user.email}`)
    } catch (err) {
      console.error(`Error migrating user ${user.email}:`, err)
    }
  }
}
```

#### 3.2 Password Reset Flow for Migrated Users
```typescript
// src/components/auth/MigrationPasswordReset.tsx
import { useState } from 'react'
import { supabase } from '../../lib/supabase'

export function MigrationPasswordReset() {
  const [email, setEmail] = useState('')
  const [sent, setSent] = useState(false)
  
  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    
    const { error } = await supabase.auth.resetPasswordForEmail(email, {
      redirectTo: `${window.location.origin}/reset-password`,
    })
    
    if (!error) {
      setSent(true)
    }
  }
  
  return (
    <div className="migration-notice">
      <h2>Welcome Back!</h2>
      <p>We've upgraded our authentication system. Please reset your password to continue.</p>
      
      {!sent ? (
        <form onSubmit={handleSubmit}>
          <input
            type="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            placeholder="Enter your email"
            required
          />
          <button type="submit">Send Reset Link</button>
        </form>
      ) : (
        <p>Check your email for the password reset link!</p>
      )}
    </div>
  )
}
```

### 4. API Route Updates

#### 4.1 Remove Old Auth Routes
```go
// Remove these routes from main.go
// r.POST("/api/v1/auth/login", authHandler.Login)
// r.POST("/api/v1/auth/logout", authHandler.Logout)
// r.POST("/api/v1/auth/refresh", authHandler.RefreshToken)
```

#### 4.2 Add Supabase Webhook Handler
```go
// internal/auth/webhook_handler.go
package auth

import (
    "crypto/hmac"
    "crypto/sha256"
    "encoding/hex"
    "encoding/json"
    "io"
    "net/http"
    
    "github.com/gin-gonic/gin"
)

type WebhookPayload struct {
    Type   string          `json:"type"`
    Table  string          `json:"table"`
    Record json.RawMessage `json:"record"`
    OldRecord json.RawMessage `json:"old_record"`
}

func HandleSupabaseWebhook(secret string) gin.HandlerFunc {
    return func(c *gin.Context) {
        body, err := io.ReadAll(c.Request.Body)
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
            return
        }
        
        // Verify webhook signature
        signature := c.GetHeader("X-Supabase-Signature")
        if !verifyWebhookSignature(body, signature, secret) {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid signature"})
            return
        }
        
        var payload WebhookPayload
        if err := json.Unmarshal(body, &payload); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
            return
        }
        
        // Handle different webhook types
        switch payload.Type {
        case "INSERT":
            if payload.Table == "auth.users" {
                // Handle new user registration
                handleNewUser(payload.Record)
            }
        case "DELETE":
            if payload.Table == "auth.users" {
                // Handle user deletion
                handleUserDeletion(payload.Record)
            }
        }
        
        c.JSON(http.StatusOK, gin.H{"status": "ok"})
    }
}
```

### 5. Testing Strategy

#### 5.1 Integration Tests
```go
// internal/auth/supabase_auth_test.go
package auth

import (
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestSupabaseAuthentication(t *testing.T) {
    client := setupTestSupabaseClient()
    
    t.Run("Valid token should authenticate", func(t *testing.T) {
        // Create test user
        user, token := createTestUser(t, client)
        
        // Verify token
        authUser, err := client.Auth.User(context.Background(), token)
        
        assert.NoError(t, err)
        assert.Equal(t, user.Email, authUser.Email)
    })
    
    t.Run("Invalid token should fail", func(t *testing.T) {
        _, err := client.Auth.User(context.Background(), "invalid-token")
        assert.Error(t, err)
    })
}
```

### 6. Rollback Plan

If issues arise during migration:

1. **Feature Flag System**
```go
if config.UseSupabaseAuth {
    r.Use(middleware.SupabaseAuth(supabaseClient))
} else {
    r.Use(middleware.JWTAuth())
}
```

2. **Dual Auth Support**
- Support both JWT and Supabase tokens during transition
- Gradually migrate users
- Monitor error rates and performance

3. **Data Sync**
- Keep user tables in sync during migration
- Log all authentication attempts for debugging
- Have rollback scripts ready