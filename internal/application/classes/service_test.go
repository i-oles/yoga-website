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
	"main/internal/domain/sender"

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

func anyValuePtr[T any](v T) *T {
	return &v
}

type mockClassesRepo struct {
	classes []models.Class
	error   error
}

func newMockClassesRepo(classes []models.Class, err error) *mockClassesRepo {
	return &mockClassesRepo{
		classes: classes,
		error:   err,
	}
}

func (m *mockClassesRepo) Get(_ context.Context, _ uuid.UUID) (models.Class, error) {
	// TODO: can I do in better?
	if len(m.classes) != 0 {
		return m.classes[0], nil
	}

	return models.Class{}, m.error
}

func (m *mockClassesRepo) List(_ context.Context) ([]models.Class, error) {
	return m.classes, m.error
}

func (m *mockClassesRepo) Insert(_ context.Context, classes []models.Class) ([]models.Class, error) {
	return classes, m.error
}

func (m *mockClassesRepo) Delete(_ context.Context, _ uuid.UUID) error {
	return m.error
}

func (m *mockClassesRepo) Update(_ context.Context, _ uuid.UUID, _ map[string]any) error {
	return m.error
}

type mockSender struct {
	err error
}

func newMockErrorSender(err error) *mockSender {
	return &mockSender{err: err}
}

func (m *mockSender) SendLinkToConfirmation(_, _, _ string) error {
	return m.err
}

func (m *mockSender) SendConfirmations(_ models.ConfirmationMsg) error {
	return m.err
}

func (m *mockSender) SendInfoAboutCancellationToOwner(_, _ string, _ time.Time) error {
	return m.err
}

func (m *mockSender) SendInfoAboutClassCancellation(_, _, _ string, _ models.Class) error {
	return m.err
}

func (m *mockSender) SendInfoAboutBookingCancellation(_, _ string, _ models.Class) error {
	return m.err
}

func (m *mockSender) SendInfoAboutUpdate(_, _, _ string, _ models.Class) error {
	return m.err
}

var testBooking = models.Booking{
	ID:                uuid.MustParse("7c9b4c3e-2a6f-4b9d-9c8f-6f1a3e0b5d42"),
	ClassID:           testID1,
	FirstName:         "Jan",
	LastName:          "Kowalski",
	Email:             "jan.kowalski@example.com",
	CreatedAt:         time.Date(2025, 1, 10, 12, 0, 0, 0, time.UTC),
	ConfirmationToken: "confirm_abc123xyz",
	Class: &models.Class{
		ID: testID1,
	},
}

var testBookingWithoutClass = models.Booking{
	ID:                uuid.MustParse("7c9b4c3e-2a6f-4b9d-9c8f-6f1a3e0b5d42"),
	ClassID:           testID1,
	FirstName:         "Adam",
	LastName:          "Kowalski",
	Email:             "adam.kowalski@example.com",
	CreatedAt:         time.Date(2024, 1, 10, 12, 0, 0, 0, time.UTC),
	ConfirmationToken: "confirm_abcxxxxxx",
}

type mockBookingsRepo struct {
	count       int
	testBooking models.Booking
	error       error
}

func newMockBookingsRepo(booking models.Booking, err error) *mockBookingsRepo {
	return &mockBookingsRepo{count: 1, testBooking: booking, error: err}
}

func (m *mockBookingsRepo) GetByID(_ context.Context, _ uuid.UUID) (models.Booking, error) {
	return models.Booking{}, m.error
}

func (m *mockBookingsRepo) GetByEmailAndClassID(_ context.Context, _ uuid.UUID, _ string) (models.Booking, error) {
	return models.Booking{}, m.error
}

func (m *mockBookingsRepo) List(_ context.Context) ([]models.Booking, error) {
	return []models.Booking{}, m.error
}

func (m *mockBookingsRepo) ListByClassID(_ context.Context, classID uuid.UUID) ([]models.Booking, error) {
	if m.testBooking.ClassID != classID {
		return []models.Booking{}, nil
	}

	return []models.Booking{m.testBooking}, m.error
}

func (m *mockBookingsRepo) CountForClassID(_ context.Context, _ uuid.UUID) (int, error) {
	return m.count, m.error
}

func (m *mockBookingsRepo) Insert(_ context.Context, _ models.Booking) (uuid.UUID, error) {
	return uuid.Nil, m.error
}

func (m *mockBookingsRepo) Delete(_ context.Context, _ uuid.UUID) error {
	return m.error
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
			classesRepo:         newMockClassesRepo(expiredAndFutureClasses, nil),
			bookingsRepo:        newMockBookingsRepo(testBooking, nil),
			wantClasses:         expiredAndFutureClassesWithCurrentCap,
		},
		{
			name:                "List only upcoming classes",
			onlyUpcomingClasses: true,
			classesLimit:        nil,
			classesRepo:         newMockClassesRepo(expiredAndFutureClasses, nil),
			bookingsRepo:        newMockBookingsRepo(testBooking, nil),
			wantClasses: []models.ClassWithCurrentCapacity{
				expiredAndFutureClassesWithCurrentCap[1],
				expiredAndFutureClassesWithCurrentCap[2],
				expiredAndFutureClassesWithCurrentCap[3],
			},
		},
		{
			name:                "List all classes with limit",
			onlyUpcomingClasses: false,
			classesLimit:        anyValuePtr(2),
			classesRepo:         newMockClassesRepo(expiredAndFutureClasses, nil),
			bookingsRepo:        newMockBookingsRepo(testBooking, nil),
			wantClasses: []models.ClassWithCurrentCapacity{
				expiredAndFutureClassesWithCurrentCap[0],
				expiredAndFutureClassesWithCurrentCap[1],
			},
		},
		{
			name:                "List upcoming classes with limit",
			onlyUpcomingClasses: true,
			classesLimit:        anyValuePtr(2),
			classesRepo:         newMockClassesRepo(expiredAndFutureClasses, nil),
			bookingsRepo:        newMockBookingsRepo(testBooking, nil),
			wantClasses: []models.ClassWithCurrentCapacity{
				expiredAndFutureClassesWithCurrentCap[1],
				expiredAndFutureClassesWithCurrentCap[2],
			},
		},
		{
			name:                "List upcoming classes with limit larger than available",
			onlyUpcomingClasses: true,
			classesLimit:        anyValuePtr(10),
			classesRepo:         newMockClassesRepo(expiredAndFutureClasses, nil),
			bookingsRepo:        newMockBookingsRepo(testBooking, nil),
			wantClasses: []models.ClassWithCurrentCapacity{
				expiredAndFutureClassesWithCurrentCap[1],
				expiredAndFutureClassesWithCurrentCap[2],
				expiredAndFutureClassesWithCurrentCap[3],
			},
		},
		{
			name:                "List classes with limit larger than available",
			onlyUpcomingClasses: false,
			classesLimit:        anyValuePtr(10),
			classesRepo:         newMockClassesRepo(expiredAndFutureClasses, nil),
			bookingsRepo:        newMockBookingsRepo(testBooking, nil),
			wantClasses:         expiredAndFutureClassesWithCurrentCap,
		},
		{
			name:                "List upcoming classes with zero limit",
			onlyUpcomingClasses: true,
			classesLimit:        anyValuePtr(0),
			classesRepo:         newMockClassesRepo(expiredAndFutureClasses, nil),
			bookingsRepo:        newMockBookingsRepo(testBooking, nil),
			wantClasses:         []models.ClassWithCurrentCapacity{},
		},
		{
			name:                "List classes with zero limit",
			onlyUpcomingClasses: false,
			classesLimit:        anyValuePtr(0),
			classesRepo:         newMockClassesRepo(expiredAndFutureClasses, nil),
			bookingsRepo:        newMockBookingsRepo(testBooking, nil),
			wantClasses:         []models.ClassWithCurrentCapacity{},
		},
		{
			name:                "List classes from empty repository",
			onlyUpcomingClasses: false,
			classesLimit:        nil,
			classesRepo:         newMockClassesRepo([]models.Class{}, nil),
			bookingsRepo:        newMockBookingsRepo(testBooking, nil),
			wantClasses:         []models.ClassWithCurrentCapacity{},
		},
		{
			name:                "List upcoming classes from empty repository",
			onlyUpcomingClasses: true,
			classesLimit:        nil,
			classesRepo:         newMockClassesRepo([]models.Class{}, nil),
			bookingsRepo:        newMockBookingsRepo(testBooking, nil),
			wantClasses:         []models.ClassWithCurrentCapacity{},
		},
		{
			name:                "try to list past classes with upcoming filter",
			onlyUpcomingClasses: true,
			classesLimit:        nil,
			classesRepo:         newMockClassesRepo([]models.Class{expiredAndFutureClasses[0]}, nil),
			bookingsRepo:        newMockBookingsRepo(testBooking, nil),
			wantClasses:         []models.ClassWithCurrentCapacity{},
		},
		{
			name:                "List upcoming classes with limit of one",
			onlyUpcomingClasses: true,
			classesLimit:        anyValuePtr(1),
			classesRepo:         newMockClassesRepo(expiredAndFutureClasses, nil),
			bookingsRepo:        newMockBookingsRepo(testBooking, nil),
			wantClasses: []models.ClassWithCurrentCapacity{
				expiredAndFutureClassesWithCurrentCap[1],
			},
		},
		{
			name:                "List classes with negative limit - should return error",
			onlyUpcomingClasses: false,
			classesLimit:        anyValuePtr(-1),
			wantError:           true,
			error: errs.ErrClassValidation(
				fmt.Errorf("classes_limit must be greater than or equal to 0, got: %d", -1),
			),
		},
		{
			name:                "List upcoming classes with negative limit - should return error",
			onlyUpcomingClasses: true,
			classesLimit:        anyValuePtr(-5),
			wantError:           true,
			error: errs.ErrClassValidation(
				fmt.Errorf("classes_limit must be greater than or equal to 0, got: %d", -5),
			),
		},
		{
			name:                "Repository error",
			onlyUpcomingClasses: false,
			classesLimit:        nil,
			classesRepo:         newMockClassesRepo(futureClasses, errors.New("db error")),
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
			classesRepo: newMockClassesRepo([]models.Class{validClass}, nil),
			want:        []models.Class{validClass},
		},
		{
			name:        "Create valid classes",
			classes:     futureClasses,
			classesRepo: newMockClassesRepo(futureClasses, nil),
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
			classesRepo: newMockClassesRepo([]models.Class{}, errors.New("db error")),
			wantError:   true,
			error:       fmt.Errorf("could not insert classes: %w", errors.New("db error")),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sender := &mockSender{}

			bookingsRepo := newMockBookingsRepo(testBooking, nil)

			service := NewService(tt.classesRepo, bookingsRepo, sender)
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

func TestService_DeleteClass(t *testing.T) {
	tests := []struct {
		name          string
		classID       uuid.UUID
		reasonMsg     *string
		classesRepo   repositories.IClasses
		bookingsRepo  repositories.IBookings
		messageSender sender.ISender
		wantError     bool
		error         error
	}{
		{
			name:          "delete class: success",
			classID:       testID1,
			reasonMsg:     anyValuePtr("testReason"),
			classesRepo:   newMockClassesRepo(futureClasses, nil),
			bookingsRepo:  newMockBookingsRepo(testBooking, nil),
			messageSender: &mockSender{},
		},
		{
			name:          "delete class: success with no bookings and no reason msg",
			classID:       testID2,
			classesRepo:   newMockClassesRepo(futureClasses, nil),
			bookingsRepo:  newMockBookingsRepo(testBooking, nil),
			messageSender: &mockSender{},
		},
		{
			name:          "delete class: error reasonMsg empty",
			classID:       testID1,
			classesRepo:   newMockClassesRepo(futureClasses, nil),
			bookingsRepo:  newMockBookingsRepo(testBooking, nil),
			messageSender: &mockSender{},
			wantError:     true,
			error: errs.ErrClassValidation(
				errors.New("reason msg can not be empty, when classes has bookings"),
			),
		},
		{
			name:          "delete class: error class not empty",
			classID:       testID1,
			classesRepo:   newMockClassesRepo(futureClasses, nil),
			bookingsRepo:  newMockBookingsRepo(testBookingWithoutClass, nil),
			reasonMsg:     anyValuePtr("testReason"),
			messageSender: &mockSender{},
			wantError:     true,
			error:         errors.New("class field should not be empty"),
		},
		{
			name:          "delete class: messageSender error",
			classID:       testID1,
			classesRepo:   newMockClassesRepo(futureClasses, nil),
			bookingsRepo:  newMockBookingsRepo(testBooking, nil),
			reasonMsg:     anyValuePtr("testReason"),
			messageSender: newMockErrorSender(errors.New("msgSender error")),
			wantError:     true,
			error:         errors.New("msgSender error"),
		},
		{
			name:          "delete class: no bookings and no reason msg, classRepo.Delete() error",
			classID:       testID2,
			classesRepo:   newMockClassesRepo(futureClasses, errors.New("db error")),
			bookingsRepo:  newMockBookingsRepo(testBooking, nil),
			messageSender: &mockSender{},
			wantError:     true,
			error:         errors.New("db error"),
		},
		{
			name:          "delete class: classRepo.Delete() error",
			classID:       testID1,
			classesRepo:   newMockClassesRepo(futureClasses, nil),
			bookingsRepo:  newMockBookingsRepo(testBooking, errors.New("db error")),
			reasonMsg:     anyValuePtr("testReason"),
			messageSender: &mockSender{},
			wantError:     true,
			error:         errors.New("db error"),
		},
		{
			name:          "delete class: classRepo.Delete() error",
			classID:       testID1,
			classesRepo:   newMockClassesRepo(futureClasses, nil),
			bookingsRepo:  newMockBookingsRepo(testBooking, errors.New("db error")),
			reasonMsg:     anyValuePtr("testReason"),
			messageSender: &mockSender{},
			wantError:     true,
			error:         errors.New("db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewService(tt.classesRepo, tt.bookingsRepo, tt.messageSender)
			ctx := context.Background()

			err := service.DeleteClass(ctx, tt.classID, tt.reasonMsg)
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
				t.Fatalf("got error: %v, but want %v", err, tt.error)
			}
		})
	}
}
