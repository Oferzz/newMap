package database

import (
	"context"
	"fmt"
	"time"

	"github.com/Oferzz/newMap/apps/api/internal/config"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoDB struct {
	Client   *mongo.Client
	Database *mongo.Database
	config   *config.DatabaseConfig
}

func NewMongoDB(cfg *config.DatabaseConfig) (*MongoDB, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOptions := options.Client().
		ApplyURI(cfg.URI).
		SetMaxPoolSize(cfg.MaxPoolSize).
		SetMinPoolSize(cfg.MinPoolSize).
		SetMaxConnIdleTime(cfg.MaxConnIdleTime)

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	db := &MongoDB{
		Client:   client,
		Database: client.Database(cfg.Name),
		config:   cfg,
	}

	if err := db.createIndexes(ctx); err != nil {
		return nil, fmt.Errorf("failed to create indexes: %w", err)
	}

	return db, nil
}

func (db *MongoDB) createIndexes(ctx context.Context) error {
	// Create indexes for users collection
	usersCollection := db.Database.Collection("users")
	_, err := usersCollection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: map[string]interface{}{"email": 1}, Options: options.Index().SetUnique(true)},
		{Keys: map[string]interface{}{"username": 1}, Options: options.Index().SetUnique(true)},
		{Keys: map[string]interface{}{"created_at": -1}},
	})
	if err != nil {
		return fmt.Errorf("failed to create users indexes: %w", err)
	}

	// Create indexes for trips collection
	tripsCollection := db.Database.Collection("trips")
	_, err = tripsCollection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: map[string]interface{}{"owner_id": 1}},
		{Keys: map[string]interface{}{"collaborators.user_id": 1}},
		{Keys: map[string]interface{}{"status": 1}},
		{Keys: map[string]interface{}{"start_date": 1}},
		{Keys: map[string]interface{}{"created_at": -1}},
		{Keys: map[string]interface{}{"name": "text", "description": "text"}},
	})
	if err != nil {
		return fmt.Errorf("failed to create trips indexes: %w", err)
	}

	// Create indexes for places collection
	placesCollection := db.Database.Collection("places")
	_, err = placesCollection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: map[string]interface{}{"trip_id": 1}},
		{Keys: map[string]interface{}{"location": "2dsphere"}},
		{Keys: map[string]interface{}{"category": 1}},
		{Keys: map[string]interface{}{"created_at": -1}},
		{Keys: map[string]interface{}{"name": "text", "description": "text"}},
	})
	if err != nil {
		return fmt.Errorf("failed to create places indexes: %w", err)
	}

	return nil
}

func (db *MongoDB) Close(ctx context.Context) error {
	return db.Client.Disconnect(ctx)
}

func (db *MongoDB) Ping(ctx context.Context) error {
	return db.Client.Ping(ctx, nil)
}

func (db *MongoDB) Collection(name string) *mongo.Collection {
	return db.Database.Collection(name)
}