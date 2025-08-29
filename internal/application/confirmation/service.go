package confirmation

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	domainErrors "main/internal/domain/errs"
	"main/internal/domain/models"
	"main/internal/domain/repositories"
	"main/internal/domain/services"
	"main/internal/infrastructure/errs"
	"main/pkg/converter"
	"time"

	"github.com/google/uuid"
)

type Service struct {
	ClassesRepo           repositories.Classes
	ConfirmedBookingRepo  repositories.ConfirmedBookings
	PendingOperationsRepo repositories.PendingOperations
	MessageSender         services.ISender
}

func New(
	classesRepo repositories.Classes,
	confirmedBookingsRepo repositories.ConfirmedBookings,
	pendingOperationsRepo repositories.PendingOperations,
	messageSender services.ISender,
) *Service {
	return &Service{
		ClassesRepo:           classesRepo,
		ConfirmedBookingRepo:  confirmedBookingsRepo,
		PendingOperationsRepo: pendingOperationsRepo,
		MessageSender:         messageSender,
	}
}

const (
	recordExistsCode = "23505"
)

func (s *Service) CreateBooking(
	ctx context.Context, token string,
) (models.Class, error) {
	pendingOperation, err := s.PendingOperationsRepo.Get(ctx, token)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.Class{}, domainErrors.ErrPendingOperationNotFound(
				fmt.Errorf("pending operation for token: %s not found", token),
			)
		}

		return models.Class{}, fmt.Errorf("error while getting pending pendingOperation: %w", err)
	}

	class, err := s.ClassesRepo.Get(ctx, pendingOperation.ClassID)
	if err != nil {
		return models.Class{}, fmt.Errorf("error while getting class: %w", err)
	}

	if class.StartTime.Before(time.Now()) {
		return models.Class{}, domainErrors.ErrExpiredClassBooking(
			class.ID,
			fmt.Errorf("class %s has expired at %v", pendingOperation.ClassID, class.StartTime),
		)
	}

	if class.CurrentCapacity < 1 {
		return models.Class{}, domainErrors.ErrSomeoneBookedClassFaster(
			fmt.Errorf("max capacity of class %d exceeded", class.MaxCapacity),
		)
	}

	if pendingOperation.Operation != models.CreateBooking {
		return models.Class{}, fmt.Errorf("invalid operation type: %s", pendingOperation.Operation)
	}

	confirmedBooking := models.ConfirmedBooking{
		ID:        uuid.New(),
		ClassID:   pendingOperation.ClassID,
		FirstName: pendingOperation.FirstName,
		LastName:  *pendingOperation.LastName,
		Email:     pendingOperation.Email,
		CreatedAt: time.Now().UTC(),
	}

	err = s.ConfirmedBookingRepo.Insert(ctx, confirmedBooking)
	if err != nil {
		return models.Class{}, fmt.Errorf("error while inserting pendingOperation: %w", err)
	}

	err = s.PendingOperationsRepo.Delete(ctx, pendingOperation.ID)
	if err != nil {
		return models.Class{}, fmt.Errorf("error while deleting pending pendingOperation: %w", err)
	}

	err = s.ClassesRepo.DecrementCurrentCapacity(ctx, pendingOperation.ClassID)
	if err != nil {
		return models.Class{}, fmt.Errorf("error while decrementing class max capacity: %w", err)
	}

	class, err = s.ClassesRepo.Get(ctx, pendingOperation.ClassID)
	if err != nil {
		return models.Class{}, fmt.Errorf("error while getting class: %w", err)
	}

	warsawTime, err := converter.ConvertToWarsawTime(class.StartTime)
	if err != nil {
		return models.Class{}, fmt.Errorf("error while converting to warsaw time: %w", err)
	}

	msgParams := models.ConfirmationFinalParams{
		RecipientEmail: pendingOperation.Email,
		RecipientName:  pendingOperation.FirstName,
		ClassName:      class.ClassName,
		ClassLevel:     class.ClassLevel,
		WeekDay:        warsawTime.Weekday().String(),
		Hour:           warsawTime.Format(converter.HourLayout),
		Date:           warsawTime.Format(converter.DateLayout),
		Location:       class.Location,
	}

	err = s.MessageSender.SendFinalConfirmation(msgParams)
	if err != nil {
		return models.Class{}, fmt.Errorf("error while sending final-confirmation: %w", err)
	}

	return class, nil
}

func (s *Service) CancelBooking(ctx context.Context, token string) (models.Class, error) {
	pendingOperation, err := s.PendingOperationsRepo.Get(ctx, token)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.Class{}, domainErrors.ErrPendingOperationNotFound(
				fmt.Errorf("pending operation for token: %s not found", token),
			)
		}

		return models.Class{}, fmt.Errorf("error while getting pending pendingOperation: %w", err)
	}

	class, err := s.ClassesRepo.Get(ctx, pendingOperation.ClassID)
	if err != nil {
		return models.Class{}, fmt.Errorf("error while getting class: %w", err)
	}

	if class.StartTime.Before(time.Now()) {
		return models.Class{}, domainErrors.ErrExpiredClassBooking(
			class.ID,
			fmt.Errorf("class %s has expired at %v", pendingOperation.ClassID, class.StartTime),
		)
	}

	if pendingOperation.Operation != models.CancelBooking {
		return models.Class{}, fmt.Errorf("invalid operation type: %s", pendingOperation.Operation)
	}

	err = s.ConfirmedBookingRepo.Delete(ctx, pendingOperation.ClassID, pendingOperation.Email)
	if err != nil {
		if errors.Is(err, errs.ErrNoRowsAffected) {
			return models.Class{}, domainErrors.ErrConfirmedBookingNotFound(
				pendingOperation.ClassID,
				pendingOperation.Email,
				fmt.Errorf("no such booking with email %s for class %v",
					pendingOperation.Email,
					pendingOperation.ClassID,
				),
			)
		}

		return models.Class{}, fmt.Errorf("error while deleting confirmed booking: %w", err)
	}

	err = s.PendingOperationsRepo.Delete(ctx, pendingOperation.ID)
	if err != nil {
		return models.Class{}, fmt.Errorf("error while deleting pending pendingOperation: %w", err)
	}

	err = s.ClassesRepo.IncrementCurrentCapacity(ctx, pendingOperation.ClassID)
	if err != nil {
		return models.Class{}, fmt.Errorf("error while incrementing class max capacity: %w", err)
	}

	class, err = s.ClassesRepo.Get(ctx, pendingOperation.ClassID)
	if err != nil {
		return models.Class{}, fmt.Errorf("error while getting class: %w", err)
	}

	return class, nil
}
