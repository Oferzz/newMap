package users

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	ErrUserNotFound     = errors.New("user not found")
	ErrEmailExists      = errors.New("email already exists")
	ErrUsernameExists   = errors.New("username already exists")
)

type Repository interface {
	Create(ctx context.Context, user *User) error
	GetByID(ctx context.Context, id primitive.ObjectID) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByUsername(ctx context.Context, username string) (*User, error)
	Update(ctx context.Context, id primitive.ObjectID, update *UpdateUserInput) error
	UpdatePassword(ctx context.Context, id primitive.ObjectID, passwordHash string) error
	UpdateLastLogin(ctx context.Context, id primitive.ObjectID) error
	Delete(ctx context.Context, id primitive.ObjectID) error
	List(ctx context.Context, filter bson.M, opts *options.FindOptions) ([]*User, error)
	Count(ctx context.Context, filter bson.M) (int64, error)
}

type mongoRepository struct {
	collection *mongo.Collection
}

func NewRepository(db *mongo.Database) Repository {
	return &mongoRepository{
		collection: db.Collection("users"),
	}
}

func (r *mongoRepository) Create(ctx context.Context, user *User) error {
	user.ID = primitive.NewObjectID()
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	user.IsActive = true
	user.IsVerified = false
	
	_, err := r.collection.InsertOne(ctx, user)
	if mongo.IsDuplicateKeyError(err) {
		if err.Error() == "email" {
			return ErrEmailExists
		}
		if err.Error() == "username" {
			return ErrUsernameExists
		}
	}
	return err
}

func (r *mongoRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*User, error) {
	var user User
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&user)
	if err == mongo.ErrNoDocuments {
		return nil, ErrUserNotFound
	}
	return &user, err
}

func (r *mongoRepository) GetByEmail(ctx context.Context, email string) (*User, error) {
	var user User
	err := r.collection.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err == mongo.ErrNoDocuments {
		return nil, ErrUserNotFound
	}
	return &user, err
}

func (r *mongoRepository) GetByUsername(ctx context.Context, username string) (*User, error) {
	var user User
	err := r.collection.FindOne(ctx, bson.M{"username": username}).Decode(&user)
	if err == mongo.ErrNoDocuments {
		return nil, ErrUserNotFound
	}
	return &user, err
}

func (r *mongoRepository) Update(ctx context.Context, id primitive.ObjectID, update *UpdateUserInput) error {
	updateDoc := bson.M{"updated_at": time.Now()}
	
	if update.FullName != nil {
		updateDoc["full_name"] = *update.FullName
	}
	if update.Bio != nil {
		updateDoc["bio"] = *update.Bio
	}
	if update.Avatar != nil {
		updateDoc["avatar"] = *update.Avatar
	}
	
	result, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{"$set": updateDoc},
	)
	
	if err != nil {
		return err
	}
	
	if result.MatchedCount == 0 {
		return ErrUserNotFound
	}
	
	return nil
}

func (r *mongoRepository) UpdatePassword(ctx context.Context, id primitive.ObjectID, passwordHash string) error {
	result, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{"$set": bson.M{
			"password_hash": passwordHash,
			"updated_at":    time.Now(),
		}},
	)
	
	if err != nil {
		return err
	}
	
	if result.MatchedCount == 0 {
		return ErrUserNotFound
	}
	
	return nil
}

func (r *mongoRepository) UpdateLastLogin(ctx context.Context, id primitive.ObjectID) error {
	now := time.Now()
	result, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{"$set": bson.M{
			"last_login_at": &now,
			"updated_at":    now,
		}},
	)
	
	if err != nil {
		return err
	}
	
	if result.MatchedCount == 0 {
		return ErrUserNotFound
	}
	
	return nil
}

func (r *mongoRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}
	
	if result.DeletedCount == 0 {
		return ErrUserNotFound
	}
	
	return nil
}

func (r *mongoRepository) List(ctx context.Context, filter bson.M, opts *options.FindOptions) ([]*User, error) {
	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	
	var users []*User
	if err := cursor.All(ctx, &users); err != nil {
		return nil, err
	}
	
	return users, nil
}

func (r *mongoRepository) Count(ctx context.Context, filter bson.M) (int64, error) {
	return r.collection.CountDocuments(ctx, filter)
}