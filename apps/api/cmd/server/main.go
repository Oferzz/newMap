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

	"github.com/Oferzz/newMap/apps/api/internal/config"
	"github.com/Oferzz/newMap/apps/api/internal/database"
	"github.com/Oferzz/newMap/apps/api/internal/domain/trips"
	"github.com/Oferzz/newMap/apps/api/internal/domain/users"
	"github.com/Oferzz/newMap/apps/api/internal/middleware"
	"github.com/Oferzz/newMap/apps/api/internal/utils"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// Connect to MongoDB
	mongodb, err := database.NewMongoDB(&cfg.Database)
	if err != nil {
		log.Fatal("Failed to connect to MongoDB:", err)
	}
	defer mongodb.Close(context.Background())

	// Initialize JWT manager
	jwtManager := utils.NewJWTManager(&cfg.JWT)

	// Initialize repositories
	userRepo := users.NewRepository(mongodb.Database)
	tripRepo := trips.NewRepository(mongodb.Database)

	// Initialize services
	userService := users.NewService(userRepo, jwtManager)
	tripService := trips.NewService(tripRepo, userRepo)

	// Initialize handlers
	userHandler := users.NewHandler(userService)
	tripHandler := trips.NewHandler(tripService)

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(jwtManager)
	rbacMiddleware := middleware.NewRBACMiddleware(userRepo, tripRepo)

	// Setup router
	router := setupRouter(cfg, userHandler, tripHandler, authMiddleware, rbacMiddleware)

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

func setupRouter(cfg *config.Config, userHandler *users.Handler, tripHandler *trips.Handler, authMiddleware *middleware.AuthMiddleware, rbacMiddleware *middleware.RBACMiddleware) *gin.Engine {
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

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "healthy",
			"time":   time.Now().UTC(),
		})
	})

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
			userRoutes.DELETE("/me", authMiddleware.RequireAuth(), userHandler.DeleteAccount)
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
	}

	return router
}