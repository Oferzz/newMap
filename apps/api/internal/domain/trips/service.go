package trips

import (
	"context"
	"errors"
	"time"

	"github.com/Oferzz/newMap/apps/api/internal/domain/users"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Service interface {
	Create(ctx context.Context, userID primitive.ObjectID, input *CreateTripInput) (*Trip, error)
	GetByID(ctx context.Context, tripID, userID primitive.ObjectID) (*Trip, error)
	Update(ctx context.Context, tripID, userID primitive.ObjectID, input *UpdateTripInput) (*Trip, error)
	Delete(ctx context.Context, tripID, userID primitive.ObjectID) error
	List(ctx context.Context, opts TripListOptions, userID *primitive.ObjectID) ([]*Trip, int64, error)
	InviteCollaborator(ctx context.Context, tripID, inviterID primitive.ObjectID, input *InviteCollaboratorInput) error
	RemoveCollaborator(ctx context.Context, tripID, removerID, userID primitive.ObjectID) error
	UpdateCollaboratorRole(ctx context.Context, tripID, updaterID primitive.ObjectID, input *UpdateCollaboratorRoleInput) error
	LeaveTrip(ctx context.Context, tripID, userID primitive.ObjectID) error
}

type service struct {
	repo     Repository
	userRepo users.Repository
}

func NewService(repo Repository, userRepo users.Repository) Service {
	return &service{
		repo:     repo,
		userRepo: userRepo,
	}
}

func (s *service) Create(ctx context.Context, userID primitive.ObjectID, input *CreateTripInput) (*Trip, error) {
	// Verify user exists
	_, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Validate dates
	if input.EndDate.Before(input.StartDate) {
		return nil, errors.New("end date must be after start date")
	}

	// Create trip
	trip := &Trip{
		Name:          input.Name,
		Description:   input.Description,
		OwnerID:       userID,
		StartDate:     input.StartDate,
		EndDate:       input.EndDate,
		IsPublic:      input.IsPublic,
		Tags:          input.Tags,
		Budget:        input.Budget,
		Collaborators: []Collaborator{},
	}

	// Set status based on dates
	now := time.Now()
	if input.StartDate.After(now) {
		trip.Status = StatusUpcoming
	} else if input.EndDate.Before(now) {
		trip.Status = StatusCompleted
	} else {
		trip.Status = StatusOngoing
	}

	if err := s.repo.Create(ctx, trip); err != nil {
		return nil, err
	}

	return trip, nil
}

func (s *service) GetByID(ctx context.Context, tripID, userID primitive.ObjectID) (*Trip, error) {
	trip, err := s.repo.GetByID(ctx, tripID)
	if err != nil {
		return nil, err
	}

	// Check if user has access to view the trip
	if !trip.IsPublic && !trip.HasCollaborator(userID) {
		return nil, ErrUnauthorized
	}

	// Increment view count if not owner or collaborator
	if !trip.HasCollaborator(userID) {
		go s.repo.IncrementViewCount(context.Background(), tripID)
	}

	return trip, nil
}

func (s *service) Update(ctx context.Context, tripID, userID primitive.ObjectID, input *UpdateTripInput) (*Trip, error) {
	// Get trip to check permissions
	trip, err := s.repo.GetByID(ctx, tripID)
	if err != nil {
		return nil, err
	}

	// Check if user can update
	if !trip.CanUserPerform(userID, users.PermissionTripUpdate) {
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
	if input.CoverImage != nil {
		update["cover_image"] = *input.CoverImage
	}
	if input.StartDate != nil {
		update["start_date"] = *input.StartDate
	}
	if input.EndDate != nil {
		update["end_date"] = *input.EndDate
	}
	if input.Status != nil && input.Status.IsValid() {
		update["status"] = *input.Status
	}
	if input.IsPublic != nil {
		update["is_public"] = *input.IsPublic
	}
	if input.Tags != nil {
		update["tags"] = input.Tags
	}
	if input.Budget != nil {
		update["budget"] = input.Budget
	}

	// Validate dates if both are being updated
	if input.StartDate != nil && input.EndDate != nil {
		if input.EndDate.Before(*input.StartDate) {
			return nil, errors.New("end date must be after start date")
		}
	}

	if err := s.repo.Update(ctx, tripID, update); err != nil {
		return nil, err
	}

	// Return updated trip
	return s.repo.GetByID(ctx, tripID)
}

func (s *service) Delete(ctx context.Context, tripID, userID primitive.ObjectID) error {
	// Get trip to check permissions
	trip, err := s.repo.GetByID(ctx, tripID)
	if err != nil {
		return err
	}

	// Only owner can delete
	if trip.OwnerID != userID {
		return ErrUnauthorized
	}

	return s.repo.Delete(ctx, tripID)
}

func (s *service) List(ctx context.Context, opts TripListOptions, userID *primitive.ObjectID) ([]*Trip, int64, error) {
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
	case "start_date":
		findOpts.SetSort(bson.D{{Key: "start_date", Value: 1}})
	case "-start_date":
		findOpts.SetSort(bson.D{{Key: "start_date", Value: -1}})
	default:
		findOpts.SetSort(bson.D{{Key: "created_at", Value: -1}})
	}

	// Get trips
	trips, err := s.repo.List(ctx, opts.Filter, findOpts)
	if err != nil {
		return nil, 0, err
	}

	// Get total count
	count, err := s.repo.Count(ctx, opts.Filter)
	if err != nil {
		return nil, 0, err
	}

	return trips, count, nil
}

func (s *service) InviteCollaborator(ctx context.Context, tripID, inviterID primitive.ObjectID, input *InviteCollaboratorInput) error {
	// Get trip to check permissions
	trip, err := s.repo.GetByID(ctx, tripID)
	if err != nil {
		return err
	}

	// Check if inviter has permission
	if !trip.CanUserPerform(inviterID, users.PermissionTripUpdate) {
		return ErrUnauthorized
	}

	// Get user by email
	user, err := s.userRepo.GetByEmail(ctx, input.Email)
	if err != nil {
		if err == users.ErrUserNotFound {
			return errors.New("user not found with this email")
		}
		return err
	}

	// Check if user is already a collaborator
	if trip.HasCollaborator(user.ID) {
		return errors.New("user is already a collaborator")
	}

	// Validate role
	if !input.Role.IsValid() {
		return errors.New("invalid role")
	}

	// Add collaborator
	collaborator := &Collaborator{
		UserID:    user.ID,
		Username:  user.Username,
		Email:     user.Email,
		Role:      input.Role,
		InvitedBy: inviterID,
	}

	return s.repo.AddCollaborator(ctx, tripID, collaborator)
}

func (s *service) RemoveCollaborator(ctx context.Context, tripID, removerID, userID primitive.ObjectID) error {
	// Get trip to check permissions
	trip, err := s.repo.GetByID(ctx, tripID)
	if err != nil {
		return err
	}

	// Only owner can remove collaborators
	if trip.OwnerID != removerID {
		return ErrUnauthorized
	}

	// Can't remove the owner
	if userID == trip.OwnerID {
		return errors.New("cannot remove trip owner")
	}

	return s.repo.RemoveCollaborator(ctx, tripID, userID)
}

func (s *service) UpdateCollaboratorRole(ctx context.Context, tripID, updaterID primitive.ObjectID, input *UpdateCollaboratorRoleInput) error {
	// Get trip to check permissions
	trip, err := s.repo.GetByID(ctx, tripID)
	if err != nil {
		return err
	}

	// Only owner can update roles
	if trip.OwnerID != updaterID {
		return ErrUnauthorized
	}

	// Parse user ID
	userID, err := primitive.ObjectIDFromHex(input.UserID)
	if err != nil {
		return errors.New("invalid user ID")
	}

	// Can't change owner's role
	if userID == trip.OwnerID {
		return errors.New("cannot change trip owner's role")
	}

	// Validate role
	if !input.Role.IsValid() {
		return errors.New("invalid role")
	}

	// Check if user is a collaborator
	if !trip.HasCollaborator(userID) {
		return errors.New("user is not a collaborator")
	}

	return s.repo.UpdateCollaboratorRole(ctx, tripID, userID, string(input.Role))
}

func (s *service) LeaveTrip(ctx context.Context, tripID, userID primitive.ObjectID) error {
	// Get trip
	trip, err := s.repo.GetByID(ctx, tripID)
	if err != nil {
		return err
	}

	// Owner can't leave their own trip
	if trip.OwnerID == userID {
		return errors.New("trip owner cannot leave the trip")
	}

	// Check if user is a collaborator
	if !trip.HasCollaborator(userID) {
		return errors.New("you are not a collaborator on this trip")
	}

	return s.repo.RemoveCollaborator(ctx, tripID, userID)
}