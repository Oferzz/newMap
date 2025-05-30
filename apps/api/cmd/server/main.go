package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Oferzz/newMap/apps/api/internal/cache"
	"github.com/Oferzz/newMap/apps/api/internal/config"
	"github.com/Oferzz/newMap/apps/api/internal/database"
	"github.com/Oferzz/newMap/apps/api/internal/domain/collections"
	"github.com/Oferzz/newMap/apps/api/internal/domain/places"
	"github.com/Oferzz/newMap/apps/api/internal/domain/trips"
	"github.com/Oferzz/newMap/apps/api/internal/domain/users"
	"github.com/Oferzz/newMap/apps/api/internal/health"
	"github.com/Oferzz/newMap/apps/api/internal/media"
	"github.com/Oferzz/newMap/apps/api/internal/middleware"
	"github.com/Oferzz/newMap/apps/api/internal/utils"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	log.Println("Starting newMap API server...")
	
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}
	
	log.Printf("Configuration loaded. Port: %s, Environment: %s", cfg.Server.Port, cfg.Server.Environment)

	// Connect to PostgreSQL
	log.Println("Connecting to PostgreSQL...")
	db, err := database.NewPostgresDB(&cfg.Database)
	if err != nil {
		log.Fatal("Failed to connect to PostgreSQL:", err)
	}
	defer db.Close()
	log.Println("PostgreSQL connected successfully")

	// Run migrations
	log.Println("Running database migrations...")
	if err := db.RunMigrations(cfg.Database.MigrationsPath); err != nil {
		log.Printf("Warning: Failed to run migrations: %v", err)
	}

	// Create database extensions
	ctx := context.Background()
	if err := db.CreateExtensions(ctx); err != nil {
		log.Printf("Warning: Failed to create extensions: %v", err)
	}

	// Connect to Redis (optional - don't fail if not available)
	var redisClient *database.RedisClient
	var cacheService cache.Cache
	redisClient, err = database.NewRedisClient(&cfg.Redis)
	if err != nil {
		log.Printf("Warning: Failed to connect to Redis, caching disabled: %v", err)
		// Use a no-op cache implementation if Redis is not available
		cacheService = cache.NewNoOpCache()
	} else {
		defer redisClient.Close()
		cacheService = cache.NewRedisCache(redisClient)
		log.Println("Redis connected, caching enabled")
	}

	// Initialize JWT manager
	jwtManager := utils.NewJWTManager(&cfg.JWT)

	// Initialize media storage
	mediaStorage, err := media.NewDiskStorage(&cfg.Media)
	if err != nil {
		log.Fatal("Failed to initialize media storage:", err)
	}

	// Initialize repositories
	userRepo := users.NewPostgresRepository(db.DB.DB)
	tripRepo := trips.NewPostgresRepository(db.DB)
	placeRepo := places.NewPostgresRepository(db.DB)
	collectionRepo := collections.NewPostgresRepository(db.DB)

	// Initialize services
	userService := users.NewPostgreSQLService(userRepo)
	
	// Use cached trip service if Redis is available
	baseTripService := trips.NewService(tripRepo, userRepo)
	var tripService trips.Service
	if cacheService != nil {
		tripService = trips.NewCachedServicePg(baseTripService, cacheService)
	} else {
		tripService = baseTripService
	}
	
	placeService := places.NewServicePg(placeRepo, tripRepo, cfg.App.MapboxAPIKey)
	mediaService := media.NewService(db.DB, mediaStorage)
	collectionService := collections.NewService(collectionRepo)

	// Initialize handlers
	userHandler := users.NewHandler(userService)
	tripHandler := trips.NewHandler(tripService)
	placeHandler := places.NewHandler(placeService)
	mediaHandler := media.NewHandler(mediaService)
	collectionHandler := collections.NewHandler(collectionService)
	healthHandler := health.NewHandler(db.DB, redisClient)

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(jwtManager)
	rbacMiddleware := middleware.NewRBACMiddleware(userRepo, tripRepo)

	// Setup router
	router := setupRouter(cfg, userHandler, tripHandler, placeHandler, mediaHandler, collectionHandler, healthHandler, authMiddleware, rbacMiddleware, mediaStorage)

	// Create server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	// Start server in goroutine
	go func() {
		log.Printf("Server starting on port %s", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Failed to start server:", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exited")
}

func setupRouter(cfg *config.Config, userHandler *users.Handler, tripHandler *trips.Handler, placeHandler *places.Handler, mediaHandler *media.Handler, collectionHandler *collections.Handler, healthHandler *health.Handler, authMiddleware *middleware.AuthMiddleware, rbacMiddleware *middleware.RBACMiddleware, mediaStorage media.Storage) *gin.Engine {
	if cfg.Server.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// CORS middleware
	corsConfig := cors.Config{
		AllowOrigins:     cfg.App.AllowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}
	router.Use(cors.New(corsConfig))

	// Health check routes
	healthHandler.RegisterRoutes(router)

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Auth routes
		auth := v1.Group("/auth")
		{
			auth.POST("/register", userHandler.Register)
			auth.POST("/login", userHandler.Login)
			auth.POST("/refresh", userHandler.RefreshToken)
		}

		// User routes
		userRoutes := v1.Group("/users")
		{
			userRoutes.GET("/me", authMiddleware.RequireAuth(), userHandler.GetProfile)
			userRoutes.PUT("/me", authMiddleware.RequireAuth(), userHandler.UpdateProfile)
			userRoutes.PUT("/me/password", authMiddleware.RequireAuth(), userHandler.ChangePassword)
			// userRoutes.DELETE("/me", authMiddleware.RequireAuth(), userHandler.DeleteAccount) // TODO: Implement DeleteAccount
		}

		// Trip routes
		tripRoutes := v1.Group("/trips")
		{
			// Public routes (authentication optional)
			tripRoutes.GET("", authMiddleware.OptionalAuth(), tripHandler.List)
			tripRoutes.GET("/:id", authMiddleware.OptionalAuth(), tripHandler.GetByID)

			// Protected routes (authentication required)
			tripRoutes.Use(authMiddleware.RequireAuth())
			{
				// Create trip (any authenticated user)
				tripRoutes.POST("", rbacMiddleware.RequireSystemPermission(users.PermissionTripCreate), tripHandler.Create)
				
				// Trip-specific routes (permission based on trip role)
				tripRoutes.PUT("/:id", rbacMiddleware.RequireTripPermission(users.PermissionTripUpdate), tripHandler.Update)
				tripRoutes.DELETE("/:id", rbacMiddleware.RequireTripOwnership(), tripHandler.Delete)
				
				// Collaborator management
				tripRoutes.POST("/:id/collaborators", rbacMiddleware.RequireTripPermission(users.PermissionTripUpdate), tripHandler.InviteCollaborator)
				tripRoutes.DELETE("/:id/collaborators/:userId", rbacMiddleware.RequireTripOwnership(), tripHandler.RemoveCollaborator)
				tripRoutes.PUT("/:id/collaborators/role", rbacMiddleware.RequireTripOwnership(), tripHandler.UpdateCollaboratorRole)
				tripRoutes.POST("/:id/leave", tripHandler.LeaveTrip)
			}
		}

		// Place routes
		placeRoutes := v1.Group("/places")
		{
			// Public place routes (no authentication required)
			placeRoutes.GET("/search", placeHandler.Search) // Public search endpoint
			
			// All other place routes require authentication
			placeRoutes.Use(authMiddleware.RequireAuth())
			{
				// List places (with filters)
				placeRoutes.GET("", placeHandler.List)
				placeRoutes.GET("/:id", placeHandler.GetByID)
				
				// Create place (requires permission on trip)
				placeRoutes.POST("", placeHandler.Create)
				
				// Update/Delete place (requires permission on trip)
				placeRoutes.PUT("/:id", placeHandler.Update)
				placeRoutes.DELETE("/:id", placeHandler.Delete)
				
				// Special operations
				placeRoutes.PUT("/:id/visited", placeHandler.MarkAsVisited)
				// placeRoutes.GET("/:id/children", placeHandler.GetChildren) // TODO: Implement GetChildren
			}
		}

		// Trip places routes (convenience endpoints)
		tripRoutes.GET("/:id/places", authMiddleware.RequireAuth(), placeHandler.GetByTripID)

		// Collection routes
		collectionRoutes := v1.Group("/collections")
		{
			collectionRoutes.Use(authMiddleware.RequireAuth())
			{
				// Collection CRUD
				collectionRoutes.POST("", collectionHandler.CreateCollection)
				collectionRoutes.GET("", collectionHandler.GetUserCollections)
				collectionRoutes.GET("/:id", collectionHandler.GetCollection)
				collectionRoutes.PUT("/:id", collectionHandler.UpdateCollection)
				collectionRoutes.DELETE("/:id", collectionHandler.DeleteCollection)
				
				// Location management
				collectionRoutes.POST("/:id/locations", collectionHandler.AddLocationToCollection)
				collectionRoutes.DELETE("/:id/locations/:locationId", collectionHandler.RemoveLocationFromCollection)
				
				// Collaborator management
				collectionRoutes.POST("/:id/collaborators", collectionHandler.AddCollaborator)
				collectionRoutes.DELETE("/:id/collaborators/:userId", collectionHandler.RemoveCollaborator)
			}
		}

		// Media routes
		mediaRoutes := v1.Group("/media")
		{
			mediaRoutes.Use(authMiddleware.RequireAuth())
			mediaRoutes.Use(media.ValidateFileUpload(cfg.Media.MaxFileSize))
			mediaHandler.RegisterRoutes(mediaRoutes)
		}
	}

	// Serve media files (for development)
	if cfg.Server.Environment != "production" {
		router.GET("/media/*filepath", mediaHandler.ServeMedia(mediaStorage))
	}

	return router
}