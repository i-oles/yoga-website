package classes

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"main/internal/domain/errs"
	"main/internal/domain/models"
	"main/internal/domain/repositories"

	"github.com/google/uuid"
)

var (
	now         = time.Now()
	pastTime    = now.Add(-2 * time.Hour)
	futureTime1 = now.Add(1 * time.Hour)
	futureTime2 = now.Add(2 * time.Hour)
	futureTime3 = now.Add(3 * time.Hour)
	testID1     = uuid.New()
	testID2     = uuid.New()
	testID3     = uuid.New()
	testID4     = uuid.New()
)

var expiredAndFutureClassesWithCurrentCap = []models.ClassWithCurrentCapacity{
	{
		ID:              testID1,
		StartTime:       pastTime,
		ClassLevel:      "Beginner",
		ClassName:       "Morning Yoga",
		CurrentCapacity: 9,
		MaxCapacity:     10,
		Location:        "Studio A",
	},
	{
		ID:              testID2,
		StartTime:       futureTime1,
		ClassLevel:      "Intermediate",
		ClassName:       "Afternoon Yoga",
		CurrentCapacity: 14,
		MaxCapacity:     15,
		Location:        "Studio B",
	},
	{
		ID:              testID3,
		StartTime:       futureTime2,
		ClassLevel:      "Advanced",
		ClassName:       "Evening Yoga",
		CurrentCapacity: 11,
		MaxCapacity:     12,
		Location:        "Studio C",
	},
	{
		ID:              testID4,
		StartTime:       futureTime3,
		ClassLevel:      "Beginner",
		ClassName:       "Night Yoga",
		CurrentCapacity: 19,
		MaxCapacity:     20,
		Location:        "Studio D",
	},
}

var expiredAndFutureClasses = []models.Class{
	{
		ID:          testID1,
		StartTime:   pastTime,
		ClassLevel:  "Beginner",
		ClassName:   "Morning Yoga",
		MaxCapacity: 10,
		Location:    "Studio A",
	},
	{
		ID:          testID2,
		StartTime:   futureTime1,
		ClassLevel:  "Intermediate",
		ClassName:   "Afternoon Yoga",
		MaxCapacity: 15,
		Location:    "Studio B",
	},
	{
		ID:          testID3,
		StartTime:   futureTime2,
		ClassLevel:  "Advanced",
		ClassName:   "Evening Yoga",
		MaxCapacity: 12,
		Location:    "Studio C",
	},
	{
		ID:          testID4,
		StartTime:   futureTime3,
		ClassLevel:  "Beginner",
		ClassName:   "Night Yoga",
		MaxCapacity: 20,
		Location:    "Studio D",
	},
}

var validClass = models.Class{
	ID:          testID1,
	StartTime:   futureTime1,
	ClassLevel:  "Beginner",
	ClassName:   "Vinyasa",
	MaxCapacity: 5,
	Location:    "Studio A",
}

var futureClasses = []models.Class{
	{
		ID:          testID2,
		StartTime:   futureTime1,
		ClassLevel:  "Intermediate",
		ClassName:   "Ashtanga",
		MaxCapacity: 15,
		Location:    "Studio B",
	},
	{
		ID:          testID3,
		StartTime:   futureTime2,
		ClassLevel:  "Advanced",
		ClassName:   "Vinyasa",
		MaxCapacity: 12,
		Location:    "Studio C",
	},
}

var expiredClass = models.Class{
	ID:          testID1,
	StartTime:   pastTime,
	ClassLevel:  "Beginner",
	ClassName:   "Vinyasa",
	MaxCapacity: 5,
	Location:    "Studio A",
}

func intPtr(v int) *int {
	return &v
}

type mockClassesRepo struct {
	classes []models.Class
}

func newMockClassesRepo(classes []models.Class) *mockClassesRepo {
	return &mockClassesRepo{
		classes: classes,
	}
}

func (m *mockClassesRepo) Get(_ context.Context, _ uuid.UUID) (models.Class, error) {
	if len(m.classes) != 0 {
		return m.classes[0], nil
	}

	return models.Class{}, nil
}

func (m *mockClassesRepo) List(_ context.Context) ([]models.Class, error) {
	return m.classes, nil
}

func (m *mockClassesRepo) Insert(_ context.Context, classes []models.Class) ([]models.Class, error) {
	return classes, nil
}

func (m *mockClassesRepo) Delete(_ context.Context, _ uuid.UUID) error {
	return nil
}

func (m *mockClassesRepo) Update(_ context.Context, _ uuid.UUID, _ map[string]any) error {
	return nil
}

type mockSender struct{}

func (m *mockSender) SendLinkToConfirmation(_, _, _ string) error {
	return nil
}

func (m *mockSender) SendConfirmations(_ models.ConfirmationMsg) error {
	return nil
}

func (m *mockSender) SendInfoAboutCancellationToOwner(_, _ string, _ time.Time) error {
	return nil
}

func (m *mockSender) SendInfoAboutClassCancellation(_, _, _ string, _ models.Class) error {
	return nil
}

func (m *mockSender) SendInfoAboutBookingCancellation(_, _ string, _ models.Class) error {
	return nil
}

func (m *mockSender) SendInfoAboutUpdate(_, _, _ string, _ models.Class) error {
	return nil
}

type mockBookingsRepo struct {
	count int
}

func newMockBookingsRepoWithOneBooking() *mockBookingsRepo {
	return &mockBookingsRepo{count: 1}
}

func (m *mockBookingsRepo) GetByID(_ context.Context, _ uuid.UUID) (models.Booking, error) {
	return models.Booking{}, nil
}

func (m *mockBookingsRepo) GetByEmailAndClassID(_ context.Context, _ uuid.UUID, _ string) (models.Booking, error) {
	return models.Booking{}, nil
}

func (m *mockBookingsRepo) List(_ context.Context) ([]models.Booking, error) {
	return []models.Booking{}, nil
}

func (m *mockBookingsRepo) ListByClassID(_ context.Context, _ uuid.UUID) ([]models.Booking, error) {
	return []models.Booking{}, nil
}

func (m *mockBookingsRepo) CountForClassID(_ context.Context, _ uuid.UUID) (int, error) {
	return m.count, nil
}

func (m *mockBookingsRepo) Insert(_ context.Context, _ models.Booking) (uuid.UUID, error) {
	return uuid.Nil, nil
}

func (m *mockBookingsRepo) Delete(_ context.Context, _ uuid.UUID) error {
	return nil
}

type mockClassesRepoError struct{}

func (m *mockClassesRepoError) Get(_ context.Context, _ uuid.UUID) (models.Class, error) {
	return models.Class{}, errors.New("db error")
}

func (m *mockClassesRepoError) List(_ context.Context) ([]models.Class, error) {
	return nil, errors.New("db error")
}

func (m *mockClassesRepoError) Insert(_ context.Context, _ []models.Class) ([]models.Class, error) {
	return nil, errors.New("db error")
}

func (m *mockClassesRepoError) Delete(_ context.Context, _ uuid.UUID) error {
	return errors.New("db error")
}

func (m *mockClassesRepoError) Update(_ context.Context, _ uuid.UUID, _ map[string]any) error {
	return errors.New("db error")
}

func TestService_ListClasses(t *testing.T) {
	tests := []struct {
		name                string
		onlyUpcomingClasses bool
		classesLimit        *int
		classesRepo         repositories.IClasses
		bookingsRepo        repositories.IBookings
		wantClasses         []models.ClassWithCurrentCapacity
		wantError           bool
		error               error
	}{
		{
			name:                "List classes without filters",
			onlyUpcomingClasses: false,
			classesLimit:        nil,
			classesRepo:         newMockClassesRepo(expiredAndFutureClasses),
			bookingsRepo:        newMockBookingsRepoWithOneBooking(),
			wantClasses:         expiredAndFutureClassesWithCurrentCap,
		},
		{
			name:                "List only upcoming classes",
			onlyUpcomingClasses: true,
			classesLimit:        nil,
			classesRepo:         newMockClassesRepo(expiredAndFutureClasses),
			bookingsRepo:        newMockBookingsRepoWithOneBooking(),
			wantClasses: []models.ClassWithCurrentCapacity{
				expiredAndFutureClassesWithCurrentCap[1],
				expiredAndFutureClassesWithCurrentCap[2],
				expiredAndFutureClassesWithCurrentCap[3],
			},
		},
		{
			name:                "List all classes with limit",
			onlyUpcomingClasses: false,
			classesLimit:        intPtr(2),
			classesRepo:         newMockClassesRepo(expiredAndFutureClasses),
			bookingsRepo:        newMockBookingsRepoWithOneBooking(),
			wantClasses: []models.ClassWithCurrentCapacity{
				expiredAndFutureClassesWithCurrentCap[0],
				expiredAndFutureClassesWithCurrentCap[1],
			},
		},
		{
			name:                "List upcoming classes with limit",
			onlyUpcomingClasses: true,
			classesLimit:        intPtr(2),
			classesRepo:         newMockClassesRepo(expiredAndFutureClasses),
			bookingsRepo:        newMockBookingsRepoWithOneBooking(),
			wantClasses: []models.ClassWithCurrentCapacity{
				expiredAndFutureClassesWithCurrentCap[1],
				expiredAndFutureClassesWithCurrentCap[2],
			},
		},
		{
			name:                "List upcoming classes with limit larger than available",
			onlyUpcomingClasses: true,
			classesLimit:        intPtr(10),
			classesRepo:         newMockClassesRepo(expiredAndFutureClasses),
			bookingsRepo:        newMockBookingsRepoWithOneBooking(),
			wantClasses: []models.ClassWithCurrentCapacity{
				expiredAndFutureClassesWithCurrentCap[1],
				expiredAndFutureClassesWithCurrentCap[2],
				expiredAndFutureClassesWithCurrentCap[3],
			},
		},
		{
			name:                "List classes with limit larger than available",
			onlyUpcomingClasses: false,
			classesLimit:        intPtr(10),
			classesRepo:         newMockClassesRepo(expiredAndFutureClasses),
			bookingsRepo:        newMockBookingsRepoWithOneBooking(),
			wantClasses:         expiredAndFutureClassesWithCurrentCap,
		},
		{
			name:                "List upcoming classes with zero limit",
			onlyUpcomingClasses: true,
			classesLimit:        intPtr(0),
			classesRepo:         newMockClassesRepo(expiredAndFutureClasses),
			bookingsRepo:        newMockBookingsRepoWithOneBooking(),
			wantClasses:         []models.ClassWithCurrentCapacity{},
		},
		{
			name:                "List classes with zero limit",
			onlyUpcomingClasses: false,
			classesLimit:        intPtr(0),
			classesRepo:         newMockClassesRepo(expiredAndFutureClasses),
			bookingsRepo:        newMockBookingsRepoWithOneBooking(),
			wantClasses:         []models.ClassWithCurrentCapacity{},
		},
		{
			name:                "List classes from empty repository",
			onlyUpcomingClasses: false,
			classesLimit:        nil,
			classesRepo:         newMockClassesRepo([]models.Class{}),
			bookingsRepo:        newMockBookingsRepoWithOneBooking(),
			wantClasses:         []models.ClassWithCurrentCapacity{},
		},
		{
			name:                "List upcoming classes from empty repository",
			onlyUpcomingClasses: true,
			classesLimit:        nil,
			classesRepo:         newMockClassesRepo([]models.Class{}),
			bookingsRepo:        newMockBookingsRepoWithOneBooking(),
			wantClasses:         []models.ClassWithCurrentCapacity{},
		},
		{
			name:                "try to list past classes with upcoming filter",
			onlyUpcomingClasses: true,
			classesLimit:        nil,
			classesRepo:         newMockClassesRepo([]models.Class{expiredAndFutureClasses[0]}),
			bookingsRepo:        newMockBookingsRepoWithOneBooking(),
			wantClasses:         []models.ClassWithCurrentCapacity{},
		},
		{
			name:                "List upcoming classes with limit of one",
			onlyUpcomingClasses: true,
			classesLimit:        intPtr(1),
			classesRepo:         newMockClassesRepo(expiredAndFutureClasses),
			bookingsRepo:        newMockBookingsRepoWithOneBooking(),
			wantClasses: []models.ClassWithCurrentCapacity{
				expiredAndFutureClassesWithCurrentCap[1],
			},
		},
		{
			name:                "List classes with negative limit - should return error",
			onlyUpcomingClasses: false,
			classesLimit:        intPtr(-1),
			wantError:           true,
			error: errs.ErrClassValidation(
				fmt.Errorf("classes_limit must be greater than or equal to 0, got: %d", -1),
			),
		},
		{
			name:                "List upcoming classes with negative limit - should return error",
			onlyUpcomingClasses: true,
			classesLimit:        intPtr(-5),
			wantError:           true,
			error: errs.ErrClassValidation(
				fmt.Errorf("classes_limit must be greater than or equal to 0, got: %d", -5),
			),
		},
		{
			name:                "Repository error",
			onlyUpcomingClasses: false,
			classesLimit:        nil,
			classesRepo:         &mockClassesRepoError{},
			wantError:           true,
			error:               errors.New("db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sender := &mockSender{}

			service := NewService(tt.classesRepo, tt.bookingsRepo, sender)
			ctx := context.Background()

			classes, err := service.ListClasses(ctx, tt.onlyUpcomingClasses, tt.classesLimit)
			if tt.wantError {
				if err == nil {
					t.Fatalf("expected error %v, but got nil", tt.error)
				}

				if !strings.Contains(err.Error(), tt.error.Error()) {
					t.Fatalf("expected error to contain %q, got %v", tt.error.Error(), err)
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !reflect.DeepEqual(classes, tt.wantClasses) {
				t.Errorf("Expected: %v, got %v", tt.wantClasses, len(classes))
			}
		})
	}
}

func TestService_CreateClasses(t *testing.T) {
	tests := []struct {
		name        string
		classes     []models.Class
		classesRepo repositories.IClasses
		want        []models.Class
		wantError   bool
		error       error
	}{
		{
			name:        "Create one valid class",
			classes:     []models.Class{validClass},
			classesRepo: newMockClassesRepo([]models.Class{validClass}),
			want:        []models.Class{validClass},
		},
		{
			name:        "Create valid classes",
			classes:     futureClasses,
			classesRepo: newMockClassesRepo(futureClasses),
			want:        futureClasses,
		},
		{
			name:      "Validation error - expired class",
			classes:   []models.Class{expiredClass},
			wantError: true,
			error: errs.ErrClassValidation(
				fmt.Errorf("class start_time: %v expired", expiredClass.StartTime),
			),
		},
		{
			name:      "Validation error - all class should start in future",
			classes:   expiredAndFutureClasses,
			wantError: true,
			error: errs.ErrClassValidation(
				fmt.Errorf("class start_time: %v expired", expiredAndFutureClasses[0].StartTime),
			),
		},
		{
			name:        "Repository insert error",
			classes:     []models.Class{validClass},
			classesRepo: &mockClassesRepoError{},
			wantError:   true,
			error:       fmt.Errorf("could not insert classes: %w", errors.New("db error")),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sender := &mockSender{}

			classesRepo := tt.classesRepo
			if classesRepo == nil {
				classesRepo = newMockClassesRepo([]models.Class{})
			}

			bookingsRepo := newMockBookingsRepoWithOneBooking()

			service := NewService(classesRepo, bookingsRepo, sender)
			ctx := context.Background()

			result, err := service.CreateClasses(ctx, tt.classes)

			if tt.wantError {
				if err == nil {
					t.Fatalf("expected error %v, got nil", tt.error)
				}

				if !strings.Contains(err.Error(), tt.error.Error()) {
					t.Fatalf("expected error to contain %q, got %v", tt.error.Error(), err)
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !reflect.DeepEqual(result, tt.want) {
				t.Errorf("expected: %v, got %v", tt.want, result)
			}
		})
	}
}
