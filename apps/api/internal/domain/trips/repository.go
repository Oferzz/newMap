package trips

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
	ErrTripNotFound = errors.New("trip not found")
	ErrUnauthorized = errors.New("unauthorized")
)

type Repository interface {
	Create(ctx context.Context, trip *Trip) error
	GetByID(ctx context.Context, id primitive.ObjectID) (*Trip, error)
	Update(ctx context.Context, id primitive.ObjectID, update bson.M) error
	Delete(ctx context.Context, id primitive.ObjectID) error
	List(ctx context.Context, filter TripFilter, opts *options.FindOptions) ([]*Trip, error)
	Count(ctx context.Context, filter TripFilter) (int64, error)
	AddCollaborator(ctx context.Context, tripID primitive.ObjectID, collaborator *Collaborator) error
	RemoveCollaborator(ctx context.Context, tripID, userID primitive.ObjectID) error
	UpdateCollaboratorRole(ctx context.Context, tripID, userID primitive.ObjectID, role string) error
	IncrementPlaceCount(ctx context.Context, tripID primitive.ObjectID, delta int) error
	IncrementViewCount(ctx context.Context, tripID primitive.ObjectID) error
}

type mongoRepository struct {
	collection *mongo.Collection
}

func NewRepository(db *mongo.Database) Repository {
	return &mongoRepository{
		collection: db.Collection("trips"),
	}
}

func (r *mongoRepository) Create(ctx context.Context, trip *Trip) error {
	trip.ID = primitive.NewObjectID()
	trip.CreatedAt = time.Now()
	trip.UpdatedAt = time.Now()
	trip.Status = StatusPlanning
	trip.PlaceCount = 0
	trip.ViewCount = 0
	
	if trip.Collaborators == nil {
		trip.Collaborators = []Collaborator{}
	}
	
	_, err := r.collection.InsertOne(ctx, trip)
	return err
}

func (r *mongoRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*Trip, error) {
	var trip Trip
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&trip)
	if err == mongo.ErrNoDocuments {
		return nil, ErrTripNotFound
	}
	return &trip, err
}

func (r *mongoRepository) Update(ctx context.Context, id primitive.ObjectID, update bson.M) error {
	update["updated_at"] = time.Now()
	
	result, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{"$set": update},
	)
	
	if err != nil {
		return err
	}
	
	if result.MatchedCount == 0 {
		return ErrTripNotFound
	}
	
	return nil
}

func (r *mongoRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}
	
	if result.DeletedCount == 0 {
		return ErrTripNotFound
	}
	
	return nil
}

func (r *mongoRepository) List(ctx context.Context, filter TripFilter, opts *options.FindOptions) ([]*Trip, error) {
	query := r.buildFilterQuery(filter)
	
	cursor, err := r.collection.Find(ctx, query, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	
	var trips []*Trip
	if err := cursor.All(ctx, &trips); err != nil {
		return nil, err
	}
	
	return trips, nil
}

func (r *mongoRepository) Count(ctx context.Context, filter TripFilter) (int64, error) {
	query := r.buildFilterQuery(filter)
	return r.collection.CountDocuments(ctx, query)
}

func (r *mongoRepository) AddCollaborator(ctx context.Context, tripID primitive.ObjectID, collaborator *Collaborator) error {
	collaborator.JoinedAt = time.Now()
	
	result, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": tripID},
		bson.M{
			"$push": bson.M{"collaborators": collaborator},
			"$set":  bson.M{"updated_at": time.Now()},
		},
	)
	
	if err != nil {
		return err
	}
	
	if result.MatchedCount == 0 {
		return ErrTripNotFound
	}
	
	return nil
}

func (r *mongoRepository) RemoveCollaborator(ctx context.Context, tripID, userID primitive.ObjectID) error {
	result, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": tripID},
		bson.M{
			"$pull": bson.M{"collaborators": bson.M{"user_id": userID}},
			"$set":  bson.M{"updated_at": time.Now()},
		},
	)
	
	if err != nil {
		return err
	}
	
	if result.MatchedCount == 0 {
		return ErrTripNotFound
	}
	
	return nil
}

func (r *mongoRepository) UpdateCollaboratorRole(ctx context.Context, tripID, userID primitive.ObjectID, role string) error {
	result, err := r.collection.UpdateOne(
		ctx,
		bson.M{
			"_id": tripID,
			"collaborators.user_id": userID,
		},
		bson.M{
			"$set": bson.M{
				"collaborators.$.role": role,
				"updated_at": time.Now(),
			},
		},
	)
	
	if err != nil {
		return err
	}
	
	if result.MatchedCount == 0 {
		return ErrTripNotFound
	}
	
	return nil
}

func (r *mongoRepository) IncrementPlaceCount(ctx context.Context, tripID primitive.ObjectID, delta int) error {
	result, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": tripID},
		bson.M{
			"$inc": bson.M{"place_count": delta},
			"$set": bson.M{"updated_at": time.Now()},
		},
	)
	
	if err != nil {
		return err
	}
	
	if result.MatchedCount == 0 {
		return ErrTripNotFound
	}
	
	return nil
}

func (r *mongoRepository) IncrementViewCount(ctx context.Context, tripID primitive.ObjectID) error {
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": tripID},
		bson.M{"$inc": bson.M{"view_count": 1}},
	)
	
	return err
}

func (r *mongoRepository) buildFilterQuery(filter TripFilter) bson.M {
	query := bson.M{}
	
	if filter.OwnerID != nil {
		query["owner_id"] = filter.OwnerID
	}
	
	if filter.CollaboratorID != nil {
		query["$or"] = []bson.M{
			{"owner_id": filter.CollaboratorID},
			{"collaborators.user_id": filter.CollaboratorID},
		}
	}
	
	if filter.Status != nil {
		query["status"] = filter.Status
	}
	
	if filter.IsPublic != nil {
		query["is_public"] = filter.IsPublic
	}
	
	if len(filter.Tags) > 0 {
		query["tags"] = bson.M{"$in": filter.Tags}
	}
	
	if filter.StartDateFrom != nil || filter.StartDateTo != nil {
		dateQuery := bson.M{}
		if filter.StartDateFrom != nil {
			dateQuery["$gte"] = filter.StartDateFrom
		}
		if filter.StartDateTo != nil {
			dateQuery["$lte"] = filter.StartDateTo
		}
		query["start_date"] = dateQuery
	}
	
	if filter.SearchQuery != "" {
		query["$text"] = bson.M{"$search": filter.SearchQuery}
	}
	
	return query
}