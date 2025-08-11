package confirmation

import (
	"context"
	"errors"
	"fmt"
	errs2 "main/internal/domain/errs"
	"main/internal/domain/models"
	"main/internal/domain/repositories"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type Service struct {
	ClassesRepo           repositories.Classes
	ConfirmedBookingRepo  repositories.ConfirmedBookings
	PendingOperationsRepo repositories.PendingOperations
}

func New(
	classesRepo repositories.Classes,
	confirmedBookingsRepo repositories.ConfirmedBookings,
	pendingOperationsRepo repositories.PendingOperations,
) *Service {
	return &Service{
		ClassesRepo:           classesRepo,
		ConfirmedBookingRepo:  confirmedBookingsRepo,
		PendingOperationsRepo: pendingOperationsRepo,
	}
}

const (
	recordExistsCode = "23505"
)

func (s *Service) CreateBooking(
	ctx context.Context, token string,
) (models.Class, error) {
	pendingOperationOpt, err := s.PendingOperationsRepo.Get(ctx, token)
	if err != nil {
		return models.Class{}, fmt.Errorf("error while getting pending pendingOperation: %w", err)
	}

	if !pendingOperationOpt.Exists() {
		return models.Class{}, errors.New("invalid or expired confirmation link")
	}

	pendingOperation := pendingOperationOpt.Get()

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
			return models.Class{}, errs2.ErrAlreadyBooked(pendingOperation.Email)
		}

		return models.Class{}, fmt.Errorf("error while inserting pendingOperation: %w", err)
	}

	err = s.PendingOperationsRepo.Delete(ctx, token)
	if err != nil {
		return models.Class{}, fmt.Errorf("error while deleting pending pendingOperation: %w", err)
	}

	err = s.ClassesRepo.DecrementMaxCapacity(ctx, pendingOperation.ClassID)
	if err != nil {
		return models.Class{}, fmt.Errorf("error while decrementing class max capacity: %w", err)
	}

	class, err := s.ClassesRepo.Get(ctx, pendingOperation.ClassID)
	if err != nil {
		return models.Class{}, fmt.Errorf("error while getting class: %w", err)
	}

	return class, nil
}

func (s *Service) CancelBooking(ctx context.Context, token string) (models.Class, error) {
	pendingOperationOpt, err := s.PendingOperationsRepo.Get(ctx, token)
	if err != nil {
		return models.Class{}, fmt.Errorf("error while getting pending pendingOperation: %w", err)
	}

	if !pendingOperationOpt.Exists() {
		return models.Class{}, errors.New("invalid or expired confirmation link")
	}

	pendingOperation := pendingOperationOpt.Get()

	if pendingOperation.Operation != models.CancelBooking {
		return models.Class{}, fmt.Errorf("invalid operation type: %s", pendingOperation.Operation)
	}

	err = s.ConfirmedBookingRepo.Delete(ctx, pendingOperation.ClassID, pendingOperation.Email)
	if err != nil {
		return models.Class{}, fmt.Errorf("error while deleting confirmed booking: %w", err)
	}

	err = s.PendingOperationsRepo.Delete(ctx, token)
	if err != nil {
		return models.Class{}, fmt.Errorf("error while deleting pending pendingOperation: %w", err)
	}

	err = s.ClassesRepo.IncrementMaxCapacity(ctx, pendingOperation.ClassID)
	if err != nil {
		return models.Class{}, fmt.Errorf("error while decrementing class max capacity: %w", err)
	}

	class, err := s.ClassesRepo.Get(ctx, pendingOperation.ClassID)
	if err != nil {
		return models.Class{}, fmt.Errorf("error while getting class: %w", err)
	}

	return class, nil
}
