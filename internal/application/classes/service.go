package classes

import (
	"context"
	"errors"
	"fmt"
	"time"

	"main/internal/domain/errs"
	"main/internal/domain/errs/api"
	"main/internal/domain/models"
	"main/internal/domain/notifier"
	"main/internal/domain/repositories"
	repositoryError "main/internal/infrastructure/errs"
	"main/pkg/tools"

	"github.com/google/uuid"
)

type service struct {
	classesRepo  repositories.IClasses
	bookingsRepo repositories.IBookings
	passesRepo   repositories.IPasses
	notifier     notifier.INotifier
}

func NewService(
	classesRepo repositories.IClasses,
	bookingsRepo repositories.IBookings,
	passesRepo repositories.IPasses,
	notifier notifier.INotifier,
) *service {
	return &service{
		classesRepo:  classesRepo,
		bookingsRepo: bookingsRepo,
		passesRepo:   passesRepo,
		notifier:     notifier,
	}
}

func (s *service) ListClasses(
	ctx context.Context,
	onlyUpcomingClasses bool,
	classesLimit *int,
) ([]models.ClassWithCurrentCapacity, error) {
	if classesLimit != nil && *classesLimit < 0 {
		return nil, api.ErrValidation(
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

func (s *service) CreateClasses(
	ctx context.Context, classes []models.Class,
) ([]models.Class, error) {
	err := validateClasses(classes)
	if err != nil {
		return nil, api.ErrValidation(err)
	}

	insertedClasses, err := s.classesRepo.Insert(ctx, classes)
	if err != nil {
		return nil, fmt.Errorf("could not insert classes: %w", err)
	}

	return insertedClasses, nil
}

func (s *service) DeleteClass(ctx context.Context, classID uuid.UUID, msg *string) error {
	bookings, err := s.bookingsRepo.ListByClassID(ctx, classID)
	if err != nil {
		return fmt.Errorf("could not get classes for classID %v: %w", classID, err)
	}

	if len(bookings) > 0 && msg == nil {
		return api.ErrValidation(
			errors.New("reason msg can not be empty, when classes has bookings"),
		)
	}

	for _, booking := range bookings {
		err = s.handleBookingBeforeClassDeletion(ctx, booking, msg)
		if err != nil {
			return fmt.Errorf("could not handle booking before class deletion: %w", err)
		}
	}

	err = s.classesRepo.Delete(ctx, classID)
	if err != nil {
		return fmt.Errorf("could not delete class: %w", err)
	}

	return nil
}

func (s *service) handleBookingBeforeClassDeletion(
	ctx context.Context,
	booking models.Booking,
	msgToUser *string,
) error {
	if booking.Class == nil {
		return errors.New("class field should not be empty")
	}

	err := s.bookingsRepo.Delete(ctx, booking.ID)
	if err != nil {
		return fmt.Errorf("could not delete booking for id %v: %w", booking.ID, err)
	}

	usedBookingIDs, totalBookings, err := s.decrementPassIfValid(ctx, booking.Email, booking.ID)
	if err != nil {
		return fmt.Errorf("could not dectemetnt pass for %s: %w", booking.Email, err)
	}

	notifierParams := models.NotifierParams{
		RecipientFirstName: booking.FirstName,
		RecipientLastName:  booking.LastName,
		RecipientEmail:     booking.Email,
		ClassName:          booking.Class.ClassName,
		ClassLevel:         booking.Class.ClassLevel,
		StartTime:          booking.Class.StartTime,
		Location:           booking.Class.Location,
		PassUsedBookingIDs: usedBookingIDs,
		PassTotalBookings:  totalBookings,
	}

	err = s.notifier.NotifyClassCancellation(notifierParams, *msgToUser)
	if err != nil {
		return fmt.Errorf("could not notify class cancellation with %+v: %w", notifierParams, err)
	}

	return nil
}

func (s *service) decrementPassIfValid(
	ctx context.Context,
	email string,
	bookingID uuid.UUID,
) ([]uuid.UUID, *int, error) {
	passOpt, err := s.passesRepo.GetByEmail(ctx, email)
	if err != nil && passOpt.Exists() {
		return nil, nil, fmt.Errorf("could not get pass: %w", err)
	}

	var updatedBookingIDs []uuid.UUID

	var totalBookings *int

	if passOpt.Exists() {
		pass := passOpt.Get()

		updatedBookingIDs, err = tools.RemoveFromSlice(pass.UsedBookingIDs, bookingID)
		if errors.Is(err, errs.ErrBookingIDNotFoundInPass) {
			return nil, nil, nil
		}

		if err != nil {
			return nil, nil, fmt.Errorf("could not remove bookingID %v from usedBookingIDs", bookingID)
		}

		err = s.passesRepo.Update(ctx, pass.ID, updatedBookingIDs, pass.TotalBookings)
		if err != nil {
			return nil, nil,
				fmt.Errorf("could not update pass for %s with %v: %w", pass.Email, updatedBookingIDs, err)
		}

		totalBookings = &pass.TotalBookings
	}

	return updatedBookingIDs, totalBookings, nil
}

func (s *service) UpdateClass(
	ctx context.Context, id uuid.UUID, update models.UpdateClass,
) (models.Class, error) {
	err := validateClassUpdate(update)
	if err != nil {
		return models.Class{}, api.ErrValidation(err)
	}

	_, err = s.classesRepo.Get(ctx, id)
	if err != nil {
		if errors.Is(err, repositoryError.ErrNotFound) {
			return models.Class{}, api.ErrNotFound(err)
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

func (s *service) sendInformationAboutClassUpdateToUsers(
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
		notifierParams := models.NotifierParams{
			RecipientEmail:     booking.Email,
			RecipientFirstName: booking.FirstName,
			RecipientLastName:  booking.LastName,
			ClassName:          updatedClass.ClassName,
			ClassLevel:         updatedClass.ClassLevel,
			StartTime:          updatedClass.StartTime,
			Location:           updatedClass.Location,
		}

		err = s.notifier.NotifyClassUpdate(notifierParams, msg)
		if err != nil {
			return fmt.Errorf("could not notify class update to with %+v: %w", notifierParams, err)
		}
	}

	return nil
}

func getDataForClassUpdate(update models.UpdateClass) (map[string]any, error) {
	updateData := map[string]any{}
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
