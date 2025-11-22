package classes

import (
	"context"
	"errors"
	"fmt"
	"main/internal/domain/errs"
	"main/internal/domain/models"
	"main/internal/domain/repositories"
	"main/internal/domain/services"
	repositoryError "main/internal/infrastructure/errs"
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
		classesRepo:   classesRepo,
		bookingsRepo:  bookingsRepo,
		MessageSender: messageSender,
	}
}

func (s *Service) ListClasses(
	ctx context.Context,
	onlyUpcomingClasses bool,
	classesLimit *int,
) ([]models.ClassWithCurrentCapacity, error) {
	if classesLimit != nil && *classesLimit < 0 {
		return nil, errs.ErrClassValidation(
			fmt.Errorf("classes_limit must be greater than or equal to 0, got: %d", *classesLimit),
		)
	}

	classes, err := s.classesRepo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not get all classes: %w", err)
	}

	result := make([]models.ClassWithCurrentCapacity, 0, len(classes))

	for _, class := range classes {
		bookingCount, err := s.bookingsRepo.CountForClassID(ctx, class.ID)
		if err != nil {
			return nil, fmt.Errorf("could not get bookings for class %v: %w", class.ID, err)
		}

		result = append(result, models.ClassWithCurrentCapacity{
			ID:              class.ID,
			StartTime:       class.StartTime,
			ClassLevel:      class.ClassLevel,
			ClassName:       class.ClassName,
			CurrentCapacity: class.MaxCapacity - bookingCount,
			MaxCapacity:     class.MaxCapacity,
			Location:        class.Location,
		})
	}

	if onlyUpcomingClasses {
		filtered := make([]models.ClassWithCurrentCapacity, 0, len(result))

		now := time.Now()
		for _, class := range result {
			if class.StartTime.After(now) {
				filtered = append(filtered, class)
			}
		}

		result = filtered
	}

	if classesLimit != nil {
		limit := min(*classesLimit, len(result))
		result = result[:limit]
	}

	return result, nil
}

func (s *Service) CreateClasses(
	ctx context.Context, classes []models.Class,
) ([]models.Class, error) {
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

func (s *Service) DeleteClass(ctx context.Context, classID uuid.UUID, reasonMsg string) error {
	bookings, err := s.bookingsRepo.ListByClassID(ctx, classID)
	if err != nil {
		return fmt.Errorf("could not get classes for classID %v: %w", classID, err)
	}

	for _, booking := range bookings {
		if booking.Class == nil {
			return errors.New("class field should not be empty")
		}

		err = s.bookingsRepo.Delete(ctx, booking.ID)
		if err != nil {
			return fmt.Errorf("could not delete booking for id %v: %w", booking.ID, err)
		}

		err := s.MessageSender.SendInfoAboutClassCancellation(
			booking.Email,
			booking.FirstName,
			reasonMsg,
			*booking.Class,
		)
		if err != nil {
			return fmt.Errorf("could not send info about class cancellation: %w", err)
		}
	}

	err = s.classesRepo.Delete(ctx, classID)
	if err != nil {
		return fmt.Errorf("could not delete class: %w", err)
	}

	return nil
}

func (s *Service) UpdateClass(
	ctx context.Context, id uuid.UUID, update models.UpdateClass,
) (models.Class, error) {
	err := validateClassUpdate(update)
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

	err = s.sendInformationAboutClassUpdateToUsers(ctx, update, updatedClass)
	if err != nil {
		return models.Class{}, fmt.Errorf("could not get class after update: %w", err)
	}

	return updatedClass, nil
}

func (s *Service) sendInformationAboutClassUpdateToUsers(
	ctx context.Context, update models.UpdateClass, updatedClass models.Class,
) error {
	if update.Location == nil && update.StartTime == nil {
		return nil
	}

	bookings, err := s.bookingsRepo.ListByClassID(ctx, updatedClass.ID)
	if err != nil {
		return fmt.Errorf("could not get bookings for class %v: %w", updatedClass.ID, err)
	}

	msg, err := setMessageForNotification(update.StartTime, update.Location)
	if err != nil {
		return fmt.Errorf("could not set msg for notification: %w", err)
	}

	for _, booking := range bookings {
		err = s.MessageSender.SendInfoAboutUpdate(
			booking.Email,
			booking.FirstName,
			msg,
			updatedClass,
		)
		if err != nil {
			return fmt.Errorf("could not send info about class update to %s: %w", booking.Email, err)
		}
	}

	return nil
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

	if update.MaxCapacity != nil {
		updateData["max_capacity"] = *update.MaxCapacity
	}

	if update.Location != nil {
		updateData["location"] = *update.Location
	}

	if len(updateData) == 0 {
		return nil, errors.New("no fields to update class")
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

func validateClassUpdate(
	update models.UpdateClass,
) error {
	if update.StartTime != nil {
		if update.StartTime.Before(time.Now()) {
			return fmt.Errorf("class start_time: %v expired", update.StartTime)
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

	return nil
}
