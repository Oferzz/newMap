package trips

import (
	"context"
	"testing"
	"time"

	"github.com/Oferzz/newMap/apps/api/internal/domain/users"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type mockTripRepository struct {
	trips map[string]*Trip
}

func newMockTripRepository() *mockTripRepository {
	return &mockTripRepository{
		trips: make(map[string]*Trip),
	}
}

func (m *mockTripRepository) Create(ctx context.Context, trip *Trip) error {
	trip.ID = primitive.NewObjectID()
	trip.CreatedAt = time.Now()
	trip.UpdatedAt = time.Now()
	m.trips[trip.ID.Hex()] = trip
	return nil
}

func (m *mockTripRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*Trip, error) {
	if trip, exists := m.trips[id.Hex()]; exists {
		return trip, nil
	}
	return nil, ErrTripNotFound
}

func (m *mockTripRepository) Update(ctx context.Context, id primitive.ObjectID, update bson.M) error {
	if trip, exists := m.trips[id.Hex()]; exists {
		// Simple update simulation
		if name, ok := update["name"].(string); ok {
			trip.Name = name
		}
		if desc, ok := update["description"].(string); ok {
			trip.Description = desc
		}
		trip.UpdatedAt = time.Now()
		return nil
	}
	return ErrTripNotFound
}

func (m *mockTripRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	if _, exists := m.trips[id.Hex()]; exists {
		delete(m.trips, id.Hex())
		return nil
	}
	return ErrTripNotFound
}

func (m *mockTripRepository) List(ctx context.Context, filter TripFilter, opts *options.FindOptions) ([]*Trip, error) {
	var trips []*Trip
	for _, trip := range m.trips {
		// Simple filter implementation
		if filter.OwnerID != nil && trip.OwnerID != *filter.OwnerID {
			continue
		}
		if filter.Status != nil && trip.Status != *filter.Status {
			continue
		}
		trips = append(trips, trip)
	}
	return trips, nil
}

func (m *mockTripRepository) Count(ctx context.Context, filter TripFilter) (int64, error) {
	trips, _ := m.List(ctx, filter, nil)
	return int64(len(trips)), nil
}

func (m *mockTripRepository) AddCollaborator(ctx context.Context, tripID primitive.ObjectID, collaborator *Collaborator) error {
	if trip, exists := m.trips[tripID.Hex()]; exists {
		trip.Collaborators = append(trip.Collaborators, *collaborator)
		trip.UpdatedAt = time.Now()
		return nil
	}
	return ErrTripNotFound
}

func (m *mockTripRepository) RemoveCollaborator(ctx context.Context, tripID, userID primitive.ObjectID) error {
	if trip, exists := m.trips[tripID.Hex()]; exists {
		for i, collab := range trip.Collaborators {
			if collab.UserID == userID {
				trip.Collaborators = append(trip.Collaborators[:i], trip.Collaborators[i+1:]...)
				trip.UpdatedAt = time.Now()
				return nil
			}
		}
		return nil
	}
	return ErrTripNotFound
}

func (m *mockTripRepository) UpdateCollaboratorRole(ctx context.Context, tripID, userID primitive.ObjectID, role string) error {
	if trip, exists := m.trips[tripID.Hex()]; exists {
		for i, collab := range trip.Collaborators {
			if collab.UserID == userID {
				trip.Collaborators[i].Role = users.Role(role)
				trip.UpdatedAt = time.Now()
				return nil
			}
		}
		return nil
	}
	return ErrTripNotFound
}

func (m *mockTripRepository) IncrementPlaceCount(ctx context.Context, tripID primitive.ObjectID, delta int) error {
	if trip, exists := m.trips[tripID.Hex()]; exists {
		trip.PlaceCount += delta
		return nil
	}
	return ErrTripNotFound
}

func (m *mockTripRepository) IncrementViewCount(ctx context.Context, tripID primitive.ObjectID) error {
	if trip, exists := m.trips[tripID.Hex()]; exists {
		trip.ViewCount++
		return nil
	}
	return ErrTripNotFound
}

type mockUserRepository struct {
	users map[string]*users.User
}

func newMockUserRepository() *mockUserRepository {
	return &mockUserRepository{
		users: make(map[string]*users.User),
	}
}

func (m *mockUserRepository) Create(ctx context.Context, user *users.User) error {
	user.ID = primitive.NewObjectID()
	m.users[user.Email] = user
	m.users[user.ID.Hex()] = user
	return nil
}

func (m *mockUserRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*users.User, error) {
	if user, exists := m.users[id.Hex()]; exists {
		return user, nil
	}
	return nil, users.ErrUserNotFound
}

func (m *mockUserRepository) GetByEmail(ctx context.Context, email string) (*users.User, error) {
	if user, exists := m.users[email]; exists {
		return user, nil
	}
	return nil, users.ErrUserNotFound
}

func (m *mockUserRepository) GetByUsername(ctx context.Context, username string) (*users.User, error) {
	return nil, users.ErrUserNotFound
}

func (m *mockUserRepository) Update(ctx context.Context, id primitive.ObjectID, update *users.UpdateUserInput) error {
	return nil
}

func (m *mockUserRepository) UpdatePassword(ctx context.Context, id primitive.ObjectID, passwordHash string) error {
	return nil
}

func (m *mockUserRepository) UpdateLastLogin(ctx context.Context, id primitive.ObjectID) error {
	return nil
}

func (m *mockUserRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	return nil
}

func (m *mockUserRepository) List(ctx context.Context, filter bson.M, opts *options.FindOptions) ([]*users.User, error) {
	return nil, nil
}

func (m *mockUserRepository) Count(ctx context.Context, filter bson.M) (int64, error) {
	return 0, nil
}

func TestTripService_Create(t *testing.T) {
	tripRepo := newMockTripRepository()
	userRepo := newMockUserRepository()
	service := NewService(tripRepo, userRepo)

	// Create test user
	testUser := &users.User{
		ID:       primitive.NewObjectID(),
		Email:    "test@example.com",
		Username: "testuser",
		FullName: "Test User",
	}
	userRepo.users[testUser.ID.Hex()] = testUser

	ctx := context.Background()

	// Test successful trip creation
	input := &CreateTripInput{
		Name:        "Test Trip",
		Description: "A test trip",
		StartDate:   time.Now().Add(24 * time.Hour),
		EndDate:     time.Now().Add(48 * time.Hour),
		IsPublic:    true,
		Tags:        []string{"test", "vacation"},
	}

	trip, err := service.Create(ctx, testUser.ID, input)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if trip.Name != input.Name {
		t.Errorf("Expected name %s, got %s", input.Name, trip.Name)
	}

	if trip.OwnerID != testUser.ID {
		t.Errorf("Expected owner ID %s, got %s", testUser.ID, trip.OwnerID)
	}

	if trip.Status != StatusUpcoming {
		t.Errorf("Expected status %s, got %s", StatusUpcoming, trip.Status)
	}
}

func TestTripService_InviteCollaborator(t *testing.T) {
	tripRepo := newMockTripRepository()
	userRepo := newMockUserRepository()
	service := NewService(tripRepo, userRepo)

	// Create test users
	owner := &users.User{
		ID:       primitive.NewObjectID(),
		Email:    "owner@example.com",
		Username: "owner",
	}
	collaborator := &users.User{
		ID:       primitive.NewObjectID(),
		Email:    "collaborator@example.com",
		Username: "collaborator",
	}
	userRepo.users[owner.ID.Hex()] = owner
	userRepo.users[owner.Email] = owner
	userRepo.users[collaborator.ID.Hex()] = collaborator
	userRepo.users[collaborator.Email] = collaborator

	// Create test trip
	trip := &Trip{
		ID:            primitive.NewObjectID(),
		Name:          "Test Trip",
		OwnerID:       owner.ID,
		Collaborators: []Collaborator{},
	}
	tripRepo.trips[trip.ID.Hex()] = trip

	ctx := context.Background()

	// Test successful invitation
	input := &InviteCollaboratorInput{
		Email: collaborator.Email,
		Role:  users.RoleEditor,
	}

	err := service.InviteCollaborator(ctx, trip.ID, owner.ID, input)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify collaborator was added
	updatedTrip := tripRepo.trips[trip.ID.Hex()]
	if len(updatedTrip.Collaborators) != 1 {
		t.Errorf("Expected 1 collaborator, got %d", len(updatedTrip.Collaborators))
	}

	if updatedTrip.Collaborators[0].UserID != collaborator.ID {
		t.Errorf("Expected collaborator ID %s, got %s", collaborator.ID, updatedTrip.Collaborators[0].UserID)
	}

	// Test inviting same user again
	err = service.InviteCollaborator(ctx, trip.ID, owner.ID, input)
	if err == nil {
		t.Error("Expected error when inviting existing collaborator")
	}
}

func TestTripService_Permissions(t *testing.T) {
	tripRepo := newMockTripRepository()
	userRepo := newMockUserRepository()
	service := NewService(tripRepo, userRepo)

	// Create test users
	owner := &users.User{
		ID: primitive.NewObjectID(),
	}
	editor := &users.User{
		ID: primitive.NewObjectID(),
	}
	viewer := &users.User{
		ID: primitive.NewObjectID(),
	}
	userRepo.users[owner.ID.Hex()] = owner
	userRepo.users[editor.ID.Hex()] = editor
	userRepo.users[viewer.ID.Hex()] = viewer

	// Create test trip with collaborators
	trip := &Trip{
		ID:      primitive.NewObjectID(),
		Name:    "Test Trip",
		OwnerID: owner.ID,
		Collaborators: []Collaborator{
			{UserID: editor.ID, Role: users.RoleEditor},
			{UserID: viewer.ID, Role: users.RoleViewer},
		},
	}
	tripRepo.trips[trip.ID.Hex()] = trip

	ctx := context.Background()

	// Test owner can update
	updateInput := &UpdateTripInput{
		Name: strPtr("Updated Trip"),
	}
	_, err := service.Update(ctx, trip.ID, owner.ID, updateInput)
	if err != nil {
		t.Errorf("Owner should be able to update trip: %v", err)
	}

	// Test editor can update
	_, err = service.Update(ctx, trip.ID, editor.ID, updateInput)
	if err != nil {
		t.Errorf("Editor should be able to update trip: %v", err)
	}

	// Test viewer cannot update
	_, err = service.Update(ctx, trip.ID, viewer.ID, updateInput)
	if err != ErrUnauthorized {
		t.Errorf("Viewer should not be able to update trip")
	}

	// Test only owner can delete
	err = service.Delete(ctx, trip.ID, editor.ID)
	if err != ErrUnauthorized {
		t.Errorf("Non-owner should not be able to delete trip")
	}

	err = service.Delete(ctx, trip.ID, owner.ID)
	if err != nil {
		t.Errorf("Owner should be able to delete trip: %v", err)
	}
}

func strPtr(s string) *string {
	return &s
}