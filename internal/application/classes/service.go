package classes

import (
	"context"
	"errors"
	"fmt"
	"main/internal/domain/errs"
	"main/internal/domain/models"
	"main/internal/domain/repositories"
	"main/internal/domain/services"
	"time"
	repositoryError "main/internal/infrastructure/errs"

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
		classesRepo:   classesRepo,
		bookingsRepo:  bookingsRepo,
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
	err := validateClasses(classes)
	if err != nil {
		return nil, errs.ErrClassValidation(err)
	}

	insertedClasses, err := s.classesRepo.Insert(ctx, classes)
	if err != nil {
		return nil, fmt.Errorf("could not insert classes: %w", err)
	}

	return insertedClasses, nil
}

func (s *Service) DeleteClass(ctx context.Context, class_id uuid.UUID) error {
	bookings, err := s.bookingsRepo.GetAllByClassID(ctx, class_id)
	if err != nil {
		return fmt.Errorf("could not get classes for classID %v: %w", class_id, err)
	}

	for _, booking := range bookings {
		if booking.Class == nil {
			return fmt.Errorf("class field should not be empty")
		}

		err := s.MessageSender.SendInfoAboutClassCancellation(
			booking.Email,
			booking.FirstName,
			*booking.Class,
		)
		if err != nil {
			return fmt.Errorf("could not send info about class cancellation: %w", err)
		}

		err = s.bookingsRepo.Delete(ctx, booking.ID)
		if err != nil {
			return fmt.Errorf("could not delete booking for id %v: %w", booking.ID, err)
		}
	}

	err = s.classesRepo.Delete(ctx, class_id)
	if err != nil {
		return fmt.Errorf("could not delete class: %w", err)
	}

	return nil
}

func (s *Service) UpdateClass(ctx context.Context, id uuid.UUID, update models.UpdateClass) (models.Class, error) {
	err := s.validateClassUpdate(ctx, id, update)
	if err != nil {
		return models.Class{}, errs.ErrClassValidation(err)
	}

	_, err = s.classesRepo.Get(ctx, id)
	if err != nil {
		if errors.Is(err, repositoryError.ErrNotFound) {
			return models.Class{}, errs.ErrClassNotFound(err) 
		}

		return models.Class{}, fmt.Errorf("could not get class for class_id %v: %w", id, err) 
	}

	updateData, err := getDataForClassUpdate(update)
	if err != nil {
		return models.Class{}, fmt.Errorf("could not get data for class update: %w", err)
	}

	err = s.classesRepo.Update(ctx, id, updateData)
	if err != nil {
		return models.Class{}, fmt.Errorf("could not update class: %w", err)
	}

	updatedClass, err := s.classesRepo.Get(ctx, id)
	if err != nil {
		return models.Class{}, fmt.Errorf("could not get class after update: %w", err)
	}

	if update.Location == nil && update.StartTime == nil {
		return updatedClass, nil
	}

	bookings, err := s.bookingsRepo.GetAllByClassID(ctx, updatedClass.ID)
	if err != nil {
		return models.Class{}, fmt.Errorf("could not get bookings for class %v: %w", updatedClass.ID, err)
	}

	msg, err := setMessageForNotification(update.StartTime, update.Location)
	if err != nil {
		return models.Class{}, fmt.Errorf("could not set msg for notification: %w", err)
	}

	for _, booking := range bookings {
		err = s.MessageSender.SendInfoAboutUpdate(
			booking.Email,
			booking.FirstName,
			msg,
			updatedClass,
		)
		if err != nil {
			return models.Class{}, fmt.Errorf("could not send info about class update to %s: %w", booking.Email, err)
		}
	}

	return updatedClass, nil
}

func getDataForClassUpdate(update models.UpdateClass) (map[string]interface{}, error) {
	updateData := map[string]interface{}{}
	if update.StartTime != nil {
		updateData["start_time"] = *update.StartTime
	}
	if update.ClassLevel != nil {
		updateData["class_level"] = *update.ClassLevel
	}
	if update.ClassName != nil {
		updateData["class_name"] = *update.ClassName
	}
	if update.CurrentCapacity != nil {
		updateData["current_capacity"] = *update.CurrentCapacity
	}
	if update.MaxCapacity != nil {
		updateData["max_capacity"] = *update.MaxCapacity
	}
	if update.Location != nil {
		updateData["location"] = *update.Location
	}

	if len(updateData) == 0 {
		return nil, fmt.Errorf("no fields to update class")
	}

	return updateData, nil
}

func setMessageForNotification(
	startTime *time.Time,
	location *string,
) (string, error) {
	if location != nil && startTime != nil {
		return "Wyjątkowo musiałem zmienić lokalizację i czas rozpoczęcia zajęć.", nil
	}

	if location != nil {
		return "Wyjątkowo musiałem zmienić lokalizację zajęć.", nil
	}

	if startTime != nil {
		return "Wyjątkowo musiałem zmienić czas rozpoczęcia zajęć.", nil
	}

	return "", errors.New("message for notification should not be empty")
}

func (s *Service) validateClassUpdate(
	ctx context.Context,
	id uuid.UUID,
	update models.UpdateClass,
) error {
	if update.StartTime != nil {
		if update.StartTime.Before(time.Now()) {
			return fmt.Errorf("class start_time: %v expired", update.StartTime)
		}
	}

	if update.CurrentCapacity != nil && update.MaxCapacity != nil {
		if *update.CurrentCapacity > *update.MaxCapacity ||
			*update.CurrentCapacity < 0 ||
			*update.MaxCapacity < 0 {
			return errors.New("current and max capacity has to be positive number, where current capacity could not be bigger then max capacity")
		}

		return nil
	}

	if update.CurrentCapacity != nil || update.MaxCapacity != nil {
		class, err := s.classesRepo.Get(ctx, id)
		if err != nil {
			return fmt.Errorf("could not get class for id: %v", id)
		}

		if update.CurrentCapacity != nil {
			if *update.CurrentCapacity < 0 {
				return fmt.Errorf("current capacity has to be positive number, got: %d", *update.CurrentCapacity)
			}

			if *update.CurrentCapacity > class.MaxCapacity {
				return fmt.Errorf("could not set currentCapacity to %d - it is bigger then maxCapacity of this class: %d",
					*update.CurrentCapacity,
					class.MaxCapacity,
				)
			}
		}

		if update.MaxCapacity != nil {
			if *update.MaxCapacity < 0 {
				return fmt.Errorf("max capacity has to be positive number, got: %d", *update.MaxCapacity)
			}

			if *update.MaxCapacity < class.CurrentCapacity {
				return fmt.Errorf("could not set maxCapacity to %d - it is lower then currentCapacity of this class: %d",
					*update.MaxCapacity,
					class.CurrentCapacity,
				)
			}
		}
	}

	return nil
}

func validateClasses(classes []models.Class) error {
	for _, class := range classes {
		err := validateClass(class)
		if err != nil {
			return fmt.Errorf("class validation failed %w", err)
		}
	}

	return nil
}

func validateClass(class models.Class) error {
	if class.StartTime.Before(time.Now()) {
		return fmt.Errorf("class start_time: %v expired", class.StartTime)
	}

	// TODO you do not need to pass currentCapacity in request, just set it as max capacity during creation
	if class.CurrentCapacity != class.MaxCapacity {
		return fmt.Errorf("%d != %d: current_capacity should be equal to max_capacity when creating class",
			class.CurrentCapacity, class.MaxCapacity,
		)
	}

	return nil
}
