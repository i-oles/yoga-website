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
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
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

	if pendingOperation.Operation != models.CreateBooking {
		return models.Class{}, fmt.Errorf("invalid operation type: %s", pendingOperation.Operation)
	}

	confirmedBooking := models.ConfirmedBooking{
		ID:        uuid.New(),
		ClassID:   pendingOperation.ClassID,
		FirstName: pendingOperation.FirstName,
		LastName:  *pendingOperation.LastName,
		Email:     pendingOperation.Email,
		CreatedAt: time.Now(),
	}

	err = s.ConfirmedBookingRepo.Insert(ctx, confirmedBooking)
	if err != nil {
		var pgErr *pq.Error
		if errors.As(err, &pgErr) && pgErr.Code == recordExistsCode {
			return models.Class{}, domainErrors.ErrConfirmedBookingAlreadyExists(pendingOperation.Email, err)
		}

		return models.Class{}, fmt.Errorf("error while inserting pendingOperation: %w", err)
	}

	err = s.PendingOperationsRepo.Delete(ctx, token)
	if err != nil {
		return models.Class{}, fmt.Errorf("error while deleting pending pendingOperation: %w", err)
	}

	err = s.ClassesRepo.DecrementCurrentCapacity(ctx, pendingOperation.ClassID)
	if err != nil {
		return models.Class{}, fmt.Errorf("error while decrementing class max capacity: %w", err)
	}

	class, err := s.ClassesRepo.Get(ctx, pendingOperation.ClassID)
	if err != nil {
		return models.Class{}, fmt.Errorf("error while getting class: %w", err)
	}

	msgParams := models.ConfirmationFinalParams{
		RecipientEmail: pendingOperation.Email,
		RecipientName:  pendingOperation.FirstName,
		ClassName:      class.ClassCategory,
		ClassLevel:     class.ClassLevel,
		DayOfWeek:      class.DayOfWeek,
		Hour:           class.StartHour(),
		Date:           class.StartDate(),
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

	err = s.PendingOperationsRepo.Delete(ctx, token)
	if err != nil {
		return models.Class{}, fmt.Errorf("error while deleting pending pendingOperation: %w", err)
	}

	err = s.ClassesRepo.IncrementCurrentCapacity(ctx, pendingOperation.ClassID)
	if err != nil {
		return models.Class{}, fmt.Errorf("error while incrementing class max capacity: %w", err)
	}

	class, err := s.ClassesRepo.Get(ctx, pendingOperation.ClassID)
	if err != nil {
		return models.Class{}, fmt.Errorf("error while getting class: %w", err)
	}

	return class, nil
}
