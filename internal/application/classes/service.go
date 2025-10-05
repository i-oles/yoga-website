package classes

import (
	"context"
	"fmt"
	"main/internal/domain/errs"
	"main/internal/domain/models"
	"main/internal/domain/repositories"
	"time"

	"github.com/google/uuid"
)

const classesLimit = 6

type Service struct {
	classesRepo  repositories.IClasses
	bookingsRepo repositories.IBookings
}

func NewService(
	classesRepo repositories.IClasses,
	bookingsRepo repositories.IBookings,
) *Service {
	return &Service{
		classesRepo:  classesRepo,
		bookingsRepo: bookingsRepo,
	}
}

func (s *Service) GetAllClasses(ctx context.Context) ([]models.Class, error) {
	filteredClasses := make([]models.Class, 0)

	classes, err := s.classesRepo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not get all classes: %w", err)
	}

	counter := 0

	for _, class := range classes {
		if counter >= classesLimit {
			break
		}

		if class.StartTime.Before(time.Now()) {
			continue
		}

		filteredClasses = append(filteredClasses, class)
		counter++
	}

	return filteredClasses, nil
}

func (s *Service) CreateClasses(ctx context.Context, classes []models.Class) ([]models.Class, error) {
	err := s.validateClasses(classes)
	if err != nil {
		return nil, errs.ErrClassValidation(err)
	}

	insertedClasses, err := s.classesRepo.Insert(ctx, classes)
	if err != nil {
		return nil, fmt.Errorf("could not insert classes: %w", err)
	}

	return insertedClasses, nil
}

func (s *Service) DeleteClass(ctx context.Context, id uuid.UUID) error {
	bookings, err := s.bookingsRepo.GetAllByClassID(ctx, id)
	if err != nil {
		return fmt.Errorf("could get classes for classID %v: %w", id, err)
	}

	if len(bookings) != 0 {
		return errs.ErrClassNotEmpty(
			fmt.Errorf("could not delete class: %v - class not empty", id),
		)
	}

	err = s.classesRepo.Delete(ctx, id)
	if err != nil {
		return fmt.Errorf("could not delete class: %w", err)
	}

	return nil
}

func (s *Service) validateClasses(classes []models.Class) error {
	for _, class := range classes {
		if class.StartTime.Before(time.Now()) {
			return fmt.Errorf("class start_time: %v expired", class.StartTime)
		}

		// TODO you do not need to pass currentCapacity in request, just set it as max capacity during creation
		if class.CurrentCapacity != class.MaxCapacity {
			return fmt.Errorf("%d != %d: current_capacity should be equal to max_capacity when creating class",
				class.CurrentCapacity, class.MaxCapacity,
			)
		}
	}

	return nil
}
