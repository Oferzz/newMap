package places

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
	ErrPlaceNotFound = errors.New("place not found")
	ErrUnauthorized  = errors.New("unauthorized")
)

type Repository interface {
	Create(ctx context.Context, place *Place) error
	GetByID(ctx context.Context, id primitive.ObjectID) (*Place, error)
	Update(ctx context.Context, id primitive.ObjectID, update bson.M) error
	Delete(ctx context.Context, id primitive.ObjectID) error
	List(ctx context.Context, filter PlaceFilter, opts *options.FindOptions) ([]*Place, error)
	Count(ctx context.Context, filter PlaceFilter) (int64, error)
	GetByTripID(ctx context.Context, tripID primitive.ObjectID) ([]*Place, error)
	GetChildren(ctx context.Context, parentID primitive.ObjectID) ([]*Place, error)
	DeleteByTripID(ctx context.Context, tripID primitive.ObjectID) error
	MarkAsVisited(ctx context.Context, id primitive.ObjectID, visited bool) error
}

type mongoRepository struct {
	collection *mongo.Collection
}

func NewRepository(db *mongo.Database) Repository {
	return &mongoRepository{
		collection: db.Collection("places"),
	}
}

func (r *mongoRepository) Create(ctx context.Context, place *Place) error {
	place.ID = primitive.NewObjectID()
	place.CreatedAt = time.Now()
	place.UpdatedAt = time.Now()
	
	if place.Images == nil {
		place.Images = []string{}
	}
	if place.Tags == nil {
		place.Tags = []string{}
	}
	
	_, err := r.collection.InsertOne(ctx, place)
	return err
}

func (r *mongoRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*Place, error) {
	var place Place
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&place)
	if err == mongo.ErrNoDocuments {
		return nil, ErrPlaceNotFound
	}
	return &place, err
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
		return ErrPlaceNotFound
	}
	
	return nil
}

func (r *mongoRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}
	
	if result.DeletedCount == 0 {
		return ErrPlaceNotFound
	}
	
	return nil
}

func (r *mongoRepository) List(ctx context.Context, filter PlaceFilter, opts *options.FindOptions) ([]*Place, error) {
	query := r.buildFilterQuery(filter)
	
	cursor, err := r.collection.Find(ctx, query, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	
	var places []*Place
	if err := cursor.All(ctx, &places); err != nil {
		return nil, err
	}
	
	return places, nil
}

func (r *mongoRepository) Count(ctx context.Context, filter PlaceFilter) (int64, error) {
	query := r.buildFilterQuery(filter)
	return r.collection.CountDocuments(ctx, query)
}

func (r *mongoRepository) GetByTripID(ctx context.Context, tripID primitive.ObjectID) ([]*Place, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"trip_id": tripID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	
	var places []*Place
	if err := cursor.All(ctx, &places); err != nil {
		return nil, err
	}
	
	return places, nil
}

func (r *mongoRepository) GetChildren(ctx context.Context, parentID primitive.ObjectID) ([]*Place, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"parent_id": parentID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	
	var places []*Place
	if err := cursor.All(ctx, &places); err != nil {
		return nil, err
	}
	
	return places, nil
}

func (r *mongoRepository) DeleteByTripID(ctx context.Context, tripID primitive.ObjectID) error {
	_, err := r.collection.DeleteMany(ctx, bson.M{"trip_id": tripID})
	return err
}

func (r *mongoRepository) MarkAsVisited(ctx context.Context, id primitive.ObjectID, visited bool) error {
	result, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{
			"$set": bson.M{
				"is_visited": visited,
				"updated_at": time.Now(),
			},
		},
	)
	
	if err != nil {
		return err
	}
	
	if result.MatchedCount == 0 {
		return ErrPlaceNotFound
	}
	
	return nil
}

func (r *mongoRepository) buildFilterQuery(filter PlaceFilter) bson.M {
	query := bson.M{}
	
	if filter.TripID != nil {
		query["trip_id"] = filter.TripID
	}
	
	if filter.ParentID != nil {
		query["parent_id"] = filter.ParentID
	}
	
	if filter.Category != nil {
		query["category"] = filter.Category
	}
	
	if filter.IsVisited != nil {
		query["is_visited"] = filter.IsVisited
	}
	
	if len(filter.Tags) > 0 {
		query["tags"] = bson.M{"$in": filter.Tags}
	}
	
	if filter.MinRating != nil {
		query["rating"] = bson.M{"$gte": filter.MinRating}
	}
	
	if filter.MaxCost != nil {
		query["cost.amount"] = bson.M{"$lte": filter.MaxCost}
	}
	
	if filter.DateFrom != nil || filter.DateTo != nil {
		dateQuery := bson.M{}
		if filter.DateFrom != nil {
			dateQuery["$gte"] = filter.DateFrom
		}
		if filter.DateTo != nil {
			dateQuery["$lte"] = filter.DateTo
		}
		query["visit_date"] = dateQuery
	}
	
	if filter.Bounds != nil {
		query["location.coordinates"] = bson.M{
			"$geoWithin": bson.M{
				"$box": [][]float64{
					{filter.Bounds.MinLng, filter.Bounds.MinLat},
					{filter.Bounds.MaxLng, filter.Bounds.MaxLat},
				},
			},
		}
	}
	
	if filter.SearchQuery != "" {
		query["$text"] = bson.M{"$search": filter.SearchQuery}
	}
	
	return query
}