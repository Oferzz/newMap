package places

import (
	"context"
	"errors"
	"fmt"

	"github.com/Oferzz/newMap/apps/api/internal/domain/trips"
	"github.com/Oferzz/newMap/apps/api/internal/domain/users"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// PlaceListOptions contains options for listing places
type PlaceListOptions struct {
	Filter PlaceFilter
	Page   int
	Limit  int
	Sort   string
}

type Service interface {
	Create(ctx context.Context, userID primitive.ObjectID, input *CreatePlaceInput) (*Place, error)
	GetByID(ctx context.Context, placeID, userID primitive.ObjectID) (*Place, error)
	Update(ctx context.Context, placeID, userID primitive.ObjectID, input *UpdatePlaceInput) (*Place, error)
	Delete(ctx context.Context, placeID, userID primitive.ObjectID) error
	List(ctx context.Context, opts PlaceListOptions, userID primitive.ObjectID) ([]*Place, int64, error)
	GetByTripID(ctx context.Context, tripID, userID primitive.ObjectID) ([]*Place, error)
	MarkAsVisited(ctx context.Context, placeID, userID primitive.ObjectID, visited bool) error
	GetChildren(ctx context.Context, parentID, userID primitive.ObjectID) ([]*Place, error)
}

type service struct {
	repo     Repository
	tripRepo trips.Repository
}

func NewService(repo Repository, tripRepo trips.Repository) Service {
	return &service{
		repo:     repo,
		tripRepo: tripRepo,
	}
}

func (s *service) Create(ctx context.Context, userID primitive.ObjectID, input *CreatePlaceInput) (*Place, error) {
	// Parse trip ID
	tripID, err := primitive.ObjectIDFromHex(input.TripID)
	if err != nil {
		return nil, errors.New("invalid trip ID")
	}

	// Get trip to check permissions
	trip, err := s.tripRepo.GetByID(ctx, tripID)
	if err != nil {
		if err == trips.ErrTripNotFound {
			return nil, errors.New("trip not found")
		}
		return nil, err
	}

	// Check if user can create places in this trip
	if !trip.CanUserPerform(userID, users.PermissionPlaceCreate) {
		return nil, ErrUnauthorized
	}

	// Parse parent ID if provided
	var parentID *primitive.ObjectID
	if input.ParentID != nil && *input.ParentID != "" {
		pid, err := primitive.ObjectIDFromHex(*input.ParentID)
		if err != nil {
			return nil, errors.New("invalid parent ID")
		}
		
		// Verify parent place exists and belongs to same trip
		parent, err := s.repo.GetByID(ctx, pid)
		if err != nil {
			return nil, errors.New("parent place not found")
		}
		if parent.TripID != tripID {
			return nil, errors.New("parent place belongs to different trip")
		}
		
		parentID = &pid
	}

	// Validate category
	if !input.Category.IsValid() {
		return nil, errors.New("invalid category")
	}

	// Create place
	place := &Place{
		TripID:        tripID,
		ParentID:      parentID,
		Name:          input.Name,
		Description:   input.Description,
		Category:      input.Category,
		Location:      NewLocation(input.Longitude, input.Latitude),
		Address:       input.Address,
		GooglePlaceID: getStringValue(input.GooglePlaceID),
		Images:        input.Images,
		Tags:          input.Tags,
		Notes:         input.Notes,
		Rating:        input.Rating,
		Cost:          input.Cost,
		VisitDate:     input.VisitDate,
		Duration:      input.Duration,
		IsVisited:     false,
		CreatedBy:     userID,
	}

	if err := s.repo.Create(ctx, place); err != nil {
		return nil, err
	}

	// Increment trip place count
	go s.tripRepo.IncrementPlaceCount(context.Background(), tripID, 1)

	return place, nil
}

func (s *service) GetByID(ctx context.Context, placeID, userID primitive.ObjectID) (*Place, error) {
	place, err := s.repo.GetByID(ctx, placeID)
	if err != nil {
		return nil, err
	}

	// Check if user has access to the trip
	trip, err := s.tripRepo.GetByID(ctx, place.TripID)
	if err != nil {
		return nil, err
	}

	if !trip.IsPublic && !trip.HasCollaborator(userID) {
		return nil, ErrUnauthorized
	}

	return place, nil
}

func (s *service) Update(ctx context.Context, placeID, userID primitive.ObjectID, input *UpdatePlaceInput) (*Place, error) {
	// Get place to check permissions
	place, err := s.repo.GetByID(ctx, placeID)
	if err != nil {
		return nil, err
	}

	// Get trip to check permissions
	trip, err := s.tripRepo.GetByID(ctx, place.TripID)
	if err != nil {
		return nil, err
	}

	// Check if user can update places in this trip
	if !trip.CanUserPerform(userID, users.PermissionPlaceUpdate) {
		return nil, ErrUnauthorized
	}

	// Build update document
	update := bson.M{}

	if input.Name != nil {
		update["name"] = *input.Name
	}
	if input.Description != nil {
		update["description"] = *input.Description
	}
	if input.Category != nil && input.Category.IsValid() {
		update["category"] = *input.Category
	}
	if input.Latitude != nil && input.Longitude != nil {
		update["location"] = NewLocation(*input.Longitude, *input.Latitude)
	}
	if input.Address != nil {
		update["address"] = *input.Address
	}
	if input.Images != nil {
		update["images"] = input.Images
	}
	if input.Tags != nil {
		update["tags"] = input.Tags
	}
	if input.Notes != nil {
		update["notes"] = *input.Notes
	}
	if input.Rating != nil {
		update["rating"] = *input.Rating
	}
	if input.Cost != nil {
		update["cost"] = input.Cost
	}
	if input.VisitDate != nil {
		update["visit_date"] = input.VisitDate
	}
	if input.Duration != nil {
		update["duration"] = *input.Duration
	}
	if input.IsVisited != nil {
		update["is_visited"] = *input.IsVisited
	}

	if err := s.repo.Update(ctx, placeID, update); err != nil {
		return nil, err
	}

	// Return updated place
	return s.repo.GetByID(ctx, placeID)
}

func (s *service) Delete(ctx context.Context, placeID, userID primitive.ObjectID) error {
	// Get place to check permissions
	place, err := s.repo.GetByID(ctx, placeID)
	if err != nil {
		return err
	}

	// Get trip to check permissions
	trip, err := s.tripRepo.GetByID(ctx, place.TripID)
	if err != nil {
		return err
	}

	// Check if user can delete places in this trip
	if !trip.CanUserPerform(userID, users.PermissionPlaceDelete) {
		return ErrUnauthorized
	}

	// Delete all child places first
	children, err := s.repo.GetChildren(ctx, placeID)
	if err != nil {
		return err
	}

	for _, child := range children {
		if err := s.Delete(ctx, child.ID, userID); err != nil {
			return fmt.Errorf("failed to delete child place: %w", err)
		}
	}

	// Delete the place
	if err := s.repo.Delete(ctx, placeID); err != nil {
		return err
	}

	// Decrement trip place count
	go s.tripRepo.IncrementPlaceCount(context.Background(), place.TripID, -1)

	return nil
}

func (s *service) List(ctx context.Context, opts PlaceListOptions, userID primitive.ObjectID) ([]*Place, int64, error) {
	// If filtering by trip, check permissions
	if opts.Filter.TripID != nil {
		trip, err := s.tripRepo.GetByID(ctx, *opts.Filter.TripID)
		if err != nil {
			return nil, 0, err
		}

		if !trip.IsPublic && !trip.HasCollaborator(userID) {
			return nil, 0, ErrUnauthorized
		}
	}

	// Build query options
	findOpts := options.Find().
		SetLimit(int64(opts.Limit)).
		SetSkip(int64((opts.Page - 1) * opts.Limit))

	// Set sort
	switch opts.Sort {
	case "name":
		findOpts.SetSort(bson.D{{Key: "name", Value: 1}})
	case "-name":
		findOpts.SetSort(bson.D{{Key: "name", Value: -1}})
	case "visit_date":
		findOpts.SetSort(bson.D{{Key: "visit_date", Value: 1}})
	case "-visit_date":
		findOpts.SetSort(bson.D{{Key: "visit_date", Value: -1}})
	case "rating":
		findOpts.SetSort(bson.D{{Key: "rating", Value: -1}})
	default:
		findOpts.SetSort(bson.D{{Key: "created_at", Value: -1}})
	}

	// Get places
	places, err := s.repo.List(ctx, opts.Filter, findOpts)
	if err != nil {
		return nil, 0, err
	}

	// Get total count
	count, err := s.repo.Count(ctx, opts.Filter)
	if err != nil {
		return nil, 0, err
	}

	return places, count, nil
}

func (s *service) GetByTripID(ctx context.Context, tripID, userID primitive.ObjectID) ([]*Place, error) {
	// Get trip to check permissions
	trip, err := s.tripRepo.GetByID(ctx, tripID)
	if err != nil {
		return nil, err
	}

	if !trip.IsPublic && !trip.HasCollaborator(userID) {
		return nil, ErrUnauthorized
	}

	return s.repo.GetByTripID(ctx, tripID)
}

func (s *service) MarkAsVisited(ctx context.Context, placeID, userID primitive.ObjectID, visited bool) error {
	// Get place to check permissions
	place, err := s.repo.GetByID(ctx, placeID)
	if err != nil {
		return err
	}

	// Get trip to check permissions
	trip, err := s.tripRepo.GetByID(ctx, place.TripID)
	if err != nil {
		return err
	}

	// Check if user can update places in this trip
	if !trip.CanUserPerform(userID, users.PermissionPlaceUpdate) {
		return ErrUnauthorized
	}

	return s.repo.MarkAsVisited(ctx, placeID, visited)
}

func (s *service) GetChildren(ctx context.Context, parentID, userID primitive.ObjectID) ([]*Place, error) {
	// Get parent place to check permissions
	parent, err := s.repo.GetByID(ctx, parentID)
	if err != nil {
		return nil, err
	}

	// Get trip to check permissions
	trip, err := s.tripRepo.GetByID(ctx, parent.TripID)
	if err != nil {
		return nil, err
	}

	if !trip.IsPublic && !trip.HasCollaborator(userID) {
		return nil, ErrUnauthorized
	}

	return s.repo.GetChildren(ctx, parentID)
}

func getStringValue(s *string) string {
	if s != nil {
		return *s
	}
	return ""
}