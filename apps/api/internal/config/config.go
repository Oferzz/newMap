package config

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	JWT      JWTConfig
	App      AppConfig
	Media    MediaConfig
	Supabase SupabaseConfig
}

type ServerConfig struct {
	Port         string
	Environment  string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

type DatabaseConfig struct {
	URI             string
	Name            string
	MaxPoolSize     int
	MinPoolSize     int
	MaxIdleTime     int // in minutes
	MigrationsPath  string
	SSLMode         string
}

type RedisConfig struct {
	URL         string
	Password    string
	DB          int
	MaxRetries  int
	PoolSize    int
}

type JWTConfig struct {
	Secret           string
	AccessExpiry     time.Duration
	RefreshExpiry    time.Duration
	Issuer           string
}

type AppConfig struct {
	Name            string
	Version         string
	AllowedOrigins  []string
	MaxUploadSize   int64
	RateLimitPerMin int
	MapboxAPIKey    string
	MongoDBURI      string // For backward compatibility if needed
}

type MediaConfig struct {
	StoragePath      string
	CDNURL           string
	MaxFileSize      int64
	AllowedMimeTypes []string
	ThumbnailQuality int
}

type SupabaseConfig struct {
	URL        string
	ServiceKey string
	AnonKey    string
}

// loadRenderSecrets loads secrets from Render's secret file if it exists
func loadRenderSecrets() {
	// Render now provides environment variables directly, not through files
	// This function is kept for backward compatibility but does nothing
	return
}

func Load() (*Config, error) {
	// Load .env file first (for local development)
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}
	
	// Load Render secrets (will override .env if keys exist)
	loadRenderSecrets()

	cfg := &Config{
		Server: ServerConfig{
			Port:         getEnv("PORT", "8080"),
			Environment:  getEnv("ENVIRONMENT", "development"),
			ReadTimeout:  getDurationEnv("SERVER_READ_TIMEOUT", 15*time.Second),
			WriteTimeout: getDurationEnv("SERVER_WRITE_TIMEOUT", 15*time.Second),
		},
		Database: DatabaseConfig{
			URI:            getEnv("DATABASE_URL", "postgresql://localhost:5432/trip_platform?sslmode=disable"),
			Name:           getEnv("DB_NAME", "trip_platform"),
			MaxPoolSize:    getIntEnv("DB_MAX_CONNECTIONS", 100),
			MinPoolSize:    getIntEnv("DB_MIN_CONNECTIONS", 10),
			MaxIdleTime:    getIntEnv("DB_MAX_IDLE_TIME", 10),
			MigrationsPath: getEnv("DB_MIGRATIONS_PATH", "./migrations"),
			SSLMode:        getEnv("DB_SSL_MODE", "disable"),
		},
		Redis: RedisConfig{
			URL:        getEnv("REDIS_URL", getEnv("INTERNAL_REDIS_URL", "redis://localhost:6379")),
			Password:   getEnv("REDIS_PASSWORD", ""),
			DB:         getIntEnv("REDIS_DB", 0),
			MaxRetries: getIntEnv("REDIS_MAX_RETRIES", 3),
			PoolSize:   getIntEnv("REDIS_POOL_SIZE", 10),
		},
		JWT: JWTConfig{
			Secret:        getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
			AccessExpiry:  getDurationEnv("JWT_ACCESS_EXPIRY", 15*time.Minute),
			RefreshExpiry: getDurationEnv("JWT_REFRESH_EXPIRY", 7*24*time.Hour),
			Issuer:        getEnv("JWT_ISSUER", "trip-platform"),
		},
		App: AppConfig{
			Name:            "Trip Planning Platform",
			Version:         "1.0.0",
			AllowedOrigins:  getAllowedOrigins(),
			MaxUploadSize:   getInt64Env("MAX_UPLOAD_SIZE", 10*1024*1024), // 10MB
			RateLimitPerMin: getIntEnv("RATE_LIMIT_PER_MIN", 60),
			MapboxAPIKey:    getEnv("MAPBOX_ACCESS_TOKEN", getEnv("MAPBOX_API_KEY", "")), // Support both naming conventions
			MongoDBURI:      getEnv("MONGODB_URI", ""), // For backward compatibility
		},
		Media: MediaConfig{
			StoragePath:      getEnv("MEDIA_PATH", "/data/media"),
			CDNURL:           getEnv("CDN_URL", "http://localhost:8080/media"),
			MaxFileSize:      getInt64Env("MAX_FILE_SIZE", 50*1024*1024), // 50MB
			AllowedMimeTypes: []string{"image/jpeg", "image/png", "image/webp", "video/mp4"},
			ThumbnailQuality: getIntEnv("THUMBNAIL_QUALITY", 85),
		},
		Supabase: SupabaseConfig{
			URL:        getEnv("SUPABASE_PROJECT_URL", ""),
			ServiceKey: getEnv("SUPABASE_PROJECT_KEY", ""),
			AnonKey:    getEnv("SUPABASE_ANON_KEY", ""),
		},
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getInt64Env(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.ParseInt(value, 10, 64); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getUint64Env(key string, defaultValue uint64) uint64 {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.ParseUint(value, 10, 64); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

func getAllowedOrigins() []string {
	// Check for environment variable first
	if originsEnv := os.Getenv("ALLOWED_ORIGINS"); originsEnv != "" {
		// Split by comma and trim spaces
		origins := make([]string, 0)
		for _, origin := range strings.Split(originsEnv, ",") {
			trimmed := strings.TrimSpace(origin)
			if trimmed != "" {
				origins = append(origins, trimmed)
			}
		}
		return origins
	}
	
	// Default origins for development and production
	return []string{
		"http://localhost:3000",
		"http://localhost:5173", 
		"https://newmap-fe.onrender.com",
		"https://newmap-qojk.onrender.com",
	}
}