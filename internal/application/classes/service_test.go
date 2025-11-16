package classes

import (
	"context"
	"errors"
	"main/internal/domain/errs"
	"main/internal/domain/models"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

func intPtr(v int) *int {
	return &v
}

type mockClassesRepo struct {
	classes []models.Class
	err     error
}

func (m *mockClassesRepo) Get(ctx context.Context, id uuid.UUID) (models.Class, error) {
	return models.Class{}, nil
}

func (m *mockClassesRepo) GetAll(ctx context.Context) ([]models.Class, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.classes, nil
}

func (m *mockClassesRepo) Insert(ctx context.Context, classes []models.Class) ([]models.Class, error) {
	return nil, nil
}

func (m *mockClassesRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return nil
}

func (m *mockClassesRepo) Update(ctx context.Context, id uuid.UUID, update map[string]any) error {
	return nil
}

type mockSender struct{}

func (m *mockSender) SendLinkToConfirmation(recipientEmail, recipientFirstName, linkToConfirmation string) error {
	return nil
}

func (m *mockSender) SendConfirmations(msg models.ConfirmationMsg) error {
	return nil
}

func (m *mockSender) SendInfoAboutCancellationToOwner(recipientFirstName, recipientLastName string, startTime time.Time) error {
	return nil
}

func (m *mockSender) SendInfoAboutClassCancellation(recipientEmail, recipientFirstName string, class models.Class) error {
	return nil
}

func (m *mockSender) SendInfoAboutBookingCancellation(recipientEmail, recipientFirstName string, class models.Class) error {
	return nil
}

func (m *mockSender) SendInfoAboutUpdate(recipientEmail, recipientFirstName, message string, class models.Class) error {
	return nil
}

type mockBookingsRepo struct{}

func (m *mockBookingsRepo) GetByEmailAndClassID(ctx context.Context, classID uuid.UUID, email string) (models.Booking, error) {
	return models.Booking{}, nil
}

func (m *mockBookingsRepo) List(ctx context.Context) ([]models.Booking, error) {
	return nil, nil
}

func (m *mockBookingsRepo) ListByClassID(ctx context.Context, classID uuid.UUID) ([]models.Booking, error) {
	return nil, nil
}

func (m *mockBookingsRepo) CountForClassID(ctx context.Context, classID uuid.UUID) (int, error) {
	return 0, nil
}

func (m *mockBookingsRepo) Get(ctx context.Context, id uuid.UUID) (models.Booking, error) {
	return models.Booking{}, nil
}

func (m *mockBookingsRepo) Insert(ctx context.Context, confirmedBooking models.Booking) (uuid.UUID, error) {
	return uuid.Nil, nil
}

func (m *mockBookingsRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return nil
}

func TestService_GetClasses(t *testing.T) {
	now := time.Now()
	pastTime := now.Add(-2 * time.Hour)
	futureTime1 := now.Add(1 * time.Hour)
	futureTime2 := now.Add(2 * time.Hour)
	futureTime3 := now.Add(3 * time.Hour)

	testClassesWithCurrentCap := []models.ClassWithCurrentCapacity{
		{
			ID:              uuid.New(),
			StartTime:       pastTime,
			ClassLevel:      "Beginner",
			ClassName:       "Morning Yoga",
			CurrentCapacity: 5,
			MaxCapacity:     10,
			Location:        "Studio A",
		},
		{
			ID:              uuid.New(),
			StartTime:       futureTime1,
			ClassLevel:      "Intermediate",
			ClassName:       "Afternoon Yoga",
			CurrentCapacity: 8,
			MaxCapacity:     15,
			Location:        "Studio B",
		},
		{
			ID:              uuid.New(),
			StartTime:       futureTime2,
			ClassLevel:      "Advanced",
			ClassName:       "Evening Yoga",
			CurrentCapacity: 10,
			MaxCapacity:     12,
			Location:        "Studio C",
		},
		{
			ID:              uuid.New(),
			StartTime:       futureTime3,
			ClassLevel:      "Beginner",
			ClassName:       "Night Yoga",
			CurrentCapacity: 3,
			MaxCapacity:     20,
			Location:        "Studio D",
		},
	}

	testClasses := []models.Class{
		{
			ID:          uuid.New(),
			StartTime:   pastTime,
			ClassLevel:  "Beginner",
			ClassName:   "Morning Yoga",
			MaxCapacity: 10,
			Location:    "Studio A",
		},
		{
			ID:          uuid.New(),
			StartTime:   futureTime1,
			ClassLevel:  "Intermediate",
			ClassName:   "Afternoon Yoga",
			MaxCapacity: 15,
			Location:    "Studio B",
		},
		{
			ID:          uuid.New(),
			StartTime:   futureTime2,
			ClassLevel:  "Advanced",
			ClassName:   "Evening Yoga",
			MaxCapacity: 12,
			Location:    "Studio C",
		},
		{
			ID:          uuid.New(),
			StartTime:   futureTime3,
			ClassLevel:  "Beginner",
			ClassName:   "Night Yoga",
			MaxCapacity: 20,
			Location:    "Studio D",
		},
	}

	tests := []struct {
		name                string
		classesWithCurrCap  []models.ClassWithCurrentCapacity
		classes             []models.Class
		repoError           error
		onlyUpcomingClasses bool
		classesLimit        *int
		wantCount           int
		wantError           bool
		validateResult      func(t *testing.T, classes []models.ClassWithCurrentCapacity)
	}{
		{
			name:                "Get all classes without filters",
			classesWithCurrCap:  testClassesWithCurrentCap,
			classes:             testClasses,
			onlyUpcomingClasses: false,
			classesLimit:        nil,
			wantCount:           4,
			wantError:           false,
			validateResult: func(t *testing.T, classes []models.ClassWithCurrentCapacity) {
				if len(classes) != 4 {
					t.Errorf("Expected 4 classes, got %d", len(classes))
				}
			},
		},
		{
			name:                "Get only upcoming classes",
			classesWithCurrCap:  testClassesWithCurrentCap,
			classes:             testClasses,
			onlyUpcomingClasses: true,
			classesLimit:        nil,
			wantCount:           3,
			wantError:           false,
			validateResult: func(t *testing.T, classes []models.ClassWithCurrentCapacity) {
				if len(classes) != 3 {
					t.Errorf("Expected 3 upcoming classes, got %d", len(classes))
				}
				for _, class := range classes {
					if class.StartTime.Before(now) {
						t.Errorf("Expected only upcoming classes, got class with start time: %v", class.StartTime)
					}
				}
			},
		},
		{
			name:                "Get all classes with limit",
			classesWithCurrCap:  testClassesWithCurrentCap,
			classes:             testClasses,
			onlyUpcomingClasses: false,
			classesLimit:        intPtr(2),
			wantCount:           2,
			wantError:           false,
			validateResult: func(t *testing.T, classes []models.ClassWithCurrentCapacity) {
				if len(classes) != 2 {
					t.Errorf("Expected 2 classes, got %d", len(classes))
				}

				hasPastClass := false
				hasFutureClass := false
				currentTime := time.Now()

				for _, class := range classes {
					if class.StartTime.Before(currentTime) {
						hasPastClass = true
					} else {
						hasFutureClass = true
					}
				}

				if !hasPastClass {
					t.Errorf("Expected at least one past class in the result, but none found")
				}

				if !hasFutureClass {
					t.Errorf("Expected at least one future class in the result, but none found")
				}
			},
		},
		{
			name:               "Get upcoming classes with limit",
			classesWithCurrCap: testClassesWithCurrentCap,

			onlyUpcomingClasses: true,
			classesLimit:        intPtr(2),
			wantCount:           2,
			wantError:           false,
			validateResult: func(t *testing.T, classes []models.ClassWithCurrentCapacity) {
				if len(classes) != 2 {
					t.Errorf("Expected 2 classes, got %d", len(classes))
				}
				for _, class := range classes {
					if class.StartTime.Before(now) {
						t.Errorf("Expected only upcoming classes, got class with start time: %v", class.StartTime)
					}
				}
			},
		},
		{
			name:                "Get upcoming classes with limit larger than available",
			classesWithCurrCap:  testClassesWithCurrentCap,
			onlyUpcomingClasses: true,
			classesLimit:        intPtr(10),
			wantCount:           3,
			wantError:           false,
			validateResult: func(t *testing.T, classes []models.ClassWithCurrentCapacity) {
				if len(classes) != 3 {
					t.Errorf("Expected 3 classes (limit larger than available), got %d", len(classes))
				}
			},
		},
		{
			name:                "Get all classes with limit larger than available",
			classesWithCurrCap:  testClassesWithCurrentCap,
			onlyUpcomingClasses: false,
			classesLimit:        intPtr(10),
			wantCount:           4,
			wantError:           false,
			validateResult: func(t *testing.T, classes []models.ClassWithCurrentCapacity) {
				if len(classes) != 4 {
					t.Errorf("Expected 4 classes (limit larger than available), got %d", len(classes))
				}
			},
		},
		{
			name:                "Get upcoming classes with zero limit",
			classesWithCurrCap:  testClassesWithCurrentCap,
			onlyUpcomingClasses: true,
			classesLimit:        intPtr(0),
			wantCount:           0,
			wantError:           false,
			validateResult: func(t *testing.T, classes []models.ClassWithCurrentCapacity) {
				if len(classes) != 0 {
					t.Errorf("Expected 0 classes (limit 0), got %d", len(classes))
				}
			},
		},
		{
			name:                "Get all classes with zero limit",
			classesWithCurrCap:  testClassesWithCurrentCap,
			onlyUpcomingClasses: false,
			classesLimit:        intPtr(0),
			wantCount:           0,
			wantError:           false,
			validateResult: func(t *testing.T, classes []models.ClassWithCurrentCapacity) {
				if len(classes) != 0 {
					t.Errorf("Expected 0 classes (limit 0), got %d", len(classes))
				}
			},
		},
		{
			name:                "Get classes from empty repository",
			classesWithCurrCap:  []models.ClassWithCurrentCapacity{},
			onlyUpcomingClasses: false,
			classesLimit:        nil,
			wantCount:           0,
			wantError:           false,
			validateResult: func(t *testing.T, classes []models.ClassWithCurrentCapacity) {
				if len(classes) != 0 {
					t.Errorf("Expected 0 classes, got %d", len(classes))
				}
			},
		},
		{
			name:                "Get upcoming classes from empty repository",
			classesWithCurrCap:  []models.ClassWithCurrentCapacity{},
			onlyUpcomingClasses: true,
			classesLimit:        nil,
			wantCount:           0,
			wantError:           false,
			validateResult: func(t *testing.T, classes []models.ClassWithCurrentCapacity) {
				if len(classes) != 0 {
					t.Errorf("Expected 0 classes, got %d", len(classes))
				}
			},
		},
		{
			name:                "Repository error",
			classesWithCurrCap:  testClassesWithCurrentCap,
			repoError:           errors.New("database error"),
			onlyUpcomingClasses: false,
			classesLimit:        nil,
			wantCount:           0,
			wantError:           true,
			validateResult: func(t *testing.T, classes []models.ClassWithCurrentCapacity) {
				if classes != nil {
					t.Errorf("Expected nil classes on error, got %v", classes)
				}
			},
		},
		{
			name:                "Get only past classes with upcoming filter",
			classesWithCurrCap:  []models.ClassWithCurrentCapacity{testClassesWithCurrentCap[0]},
			onlyUpcomingClasses: true,
			classesLimit:        nil,
			wantCount:           0,
			wantError:           false,
			validateResult: func(t *testing.T, classes []models.ClassWithCurrentCapacity) {
				if len(classes) != 0 {
					t.Errorf("Expected 0 classes (all are past), got %d", len(classes))
				}
			},
		},
		{
			name:                "Get upcoming classes with limit of one",
			classesWithCurrCap:  testClassesWithCurrentCap,
			onlyUpcomingClasses: true,
			classesLimit:        intPtr(1),
			wantCount:           1,
			wantError:           false,
			validateResult: func(t *testing.T, classes []models.ClassWithCurrentCapacity) {
				if len(classes) != 1 {
					t.Errorf("Expected 1 class, got %d", len(classes))
				}
				if classes[0].StartTime.Before(now) {
					t.Errorf("Expected upcoming class, got class with start time: %v", classes[0].StartTime)
				}
			},
		},
		{
			name:                "Get classes with negative limit - should return error",
			classesWithCurrCap:  testClassesWithCurrentCap,
			onlyUpcomingClasses: false,
			classesLimit:        intPtr(-1),
			wantCount:           0,
			wantError:           true,
			validateResult: func(t *testing.T, classes []models.ClassWithCurrentCapacity) {
				if classes != nil {
					t.Errorf("Expected nil classes on error, got %v", classes)
				}
			},
		},
		{
			name:                "Get upcoming classes with negative limit - should return error",
			classesWithCurrCap:  testClassesWithCurrentCap,
			onlyUpcomingClasses: true,
			classesLimit:        intPtr(-5),
			wantCount:           0,
			wantError:           true,
			validateResult: func(t *testing.T, classes []models.ClassWithCurrentCapacity) {
				if classes != nil {
					t.Errorf("Expected nil classes on error, got %v", classes)
				}
			},
		},
		{
			name:                "Get classes with large negative limit - should return error",
			classesWithCurrCap:  testClassesWithCurrentCap,
			onlyUpcomingClasses: false,
			classesLimit:        intPtr(-100),
			wantCount:           0,
			wantError:           true,
			validateResult: func(t *testing.T, classes []models.ClassWithCurrentCapacity) {
				if classes != nil {
					t.Errorf("Expected nil classes on error, got %v", classes)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			classesRepo := &mockClassesRepo{
				classes: tt.classes,
				err:     tt.repoError,
			}
			bookingsRepo := &mockBookingsRepo{}

			sender := &mockSender{}

			service := NewService(classesRepo, bookingsRepo, sender)
			ctx := context.Background()

			classes, err := service.ListClasses(ctx, tt.onlyUpcomingClasses, tt.classesLimit)

			if tt.wantError {
				if err == nil {
					t.Errorf("Expected error, got nil")
				} else {
					if tt.classesLimit != nil && *tt.classesLimit < 0 {
						errorMsg := err.Error()
						if !strings.Contains(errorMsg, "classes_limit") && !strings.Contains(errorMsg, "greater than or equal to 0") {
							t.Errorf("Expected error message about classes_limit validation, got: %v", err)
						}

						var classError *errs.ClassError
						if !errors.As(err, &classError) {
							t.Errorf("Expected ClassError, got: %T", err)
						} else if classError.Code != errs.BadRequestCode {
							t.Errorf("Expected BadRequestCode, got: %d", classError.Code)
						}
					}
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if len(classes) != tt.wantCount {
					t.Errorf("Expected %d classes, got %d", tt.wantCount, len(classes))
				}
			}

			if tt.validateResult != nil {
				tt.validateResult(t, classes)
			}
		})
	}
}
