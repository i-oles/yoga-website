package classes

import (
	"context"
	"fmt"
	"main/internal/domain/errs"
	"main/internal/domain/models"
	"main/internal/domain/repositories"
	"main/internal/domain/services"
	"time"

	"github.com/google/uuid"
)

type Service struct {
	classesRepo   repositories.IClasses
	bookingsRepo  repositories.IBookings
	MessageSender services.ISender
}

func NewService(
	classesRepo repositories.IClasses,
	bookingsRepo repositories.IBookings,
	messageSender services.ISender,
) *Service {
	return &Service{
		classesRepo:  classesRepo,
		bookingsRepo: bookingsRepo,
		MessageSender: messageSender,
	}
}

func (s *Service) GetClasses(
	ctx context.Context,
	onlyUpcomingClasses bool,
	classesLimit *int,
) ([]models.Class, error) {
	if classesLimit != nil && *classesLimit < 0 {
		return nil, errs.ErrClassValidation(
			fmt.Errorf("classes_limit must be greater than or equal to 0, got: %d", *classesLimit),
		)
	}

	classes, err := s.classesRepo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not get all classes: %w", err)
	}

	hasLimit := classesLimit != nil

	if !hasLimit && !onlyUpcomingClasses {
		return classes, nil
	}

	var filteredClasses []models.Class

	if onlyUpcomingClasses {
		filteredClasses = make([]models.Class, 0)
		for _, class := range classes {
			if class.StartTime.Before(time.Now()) {
				continue
			}

			filteredClasses = append(filteredClasses, class)
		}
	} else {
		filteredClasses = classes
	}

	if !hasLimit {
		return filteredClasses, nil
	}

	limit := min(*classesLimit, len(filteredClasses))

	return filteredClasses[:limit], nil

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
		return fmt.Errorf("could not get classes for classID %v: %w", id, err)
	}

	for _, booking := range bookings {
		fmt.Printf("class info: %v", booking.Class)
		err := s.MessageSender.SendInfoAboutClassCancellation(
			booking.Email,
			booking.FirstName,
			booking.Class,
		)
		if err != nil {
			return fmt.Errorf("could not send info about class cancellation: %w", err)
		}

		err = s.bookingsRepo.Delete(ctx, booking.ID)
		if err != nil {
			return fmt.Errorf("could not delete booking for id %v: %w", booking.ID, err)
		}
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
