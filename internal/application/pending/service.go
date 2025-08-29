package pending

import (
	"context"
	"errors"
	"fmt"
	domainErrors "main/internal/domain/errs"
	"main/internal/domain/models"
	"main/internal/domain/repositories"
	"main/internal/domain/services"
	"main/internal/infrastructure/errs"
	"time"

	"github.com/google/uuid"
)

//TODO: refactor this methods - duplicated code

type Service struct {
	ClassesRepo           repositories.Classes
	PendingOperationsRepo repositories.PendingOperations
	ConfirmedBookingsRepo repositories.ConfirmedBookings
	TokenGenerator        services.ITokenGenerator
	MessageSender         services.ISender
	DomainAddr            string
}

func New(
	classesRepo repositories.Classes,
	pendingOperationsRepo repositories.PendingOperations,
	confirmedBookingsRepo repositories.ConfirmedBookings,
	tokenGenerator services.ITokenGenerator,
	messageSender services.ISender,
	domainAddr string,
) *Service {
	return &Service{
		ClassesRepo:           classesRepo,
		PendingOperationsRepo: pendingOperationsRepo,
		ConfirmedBookingsRepo: confirmedBookingsRepo,
		TokenGenerator:        tokenGenerator,
		MessageSender:         messageSender,
		DomainAddr:            domainAddr,
	}
}

func (s *Service) CreateBooking(
	ctx context.Context,
	createParams models.CreateParams,
) (uuid.UUID, error) {
	_, err := s.ConfirmedBookingsRepo.Get(ctx, createParams.ClassID, createParams.Email)
	if err == nil {
		return uuid.Nil, domainErrors.ErrConfirmedBookingAlreadyExists(createParams.ClassID, createParams.Email, err)
	}

	if !errors.Is(err, errs.ErrNotFound) {
		return uuid.Nil, fmt.Errorf("could not get confirmed booking: %w", err)
	}

	err = s.validatePendingOperationNumberPerUser(ctx, createParams.ClassID, createParams.Email, models.CreateBooking)
	if err != nil {
		return uuid.Nil, fmt.Errorf("validate pending operation number per user: %w", err)
	}

	class, err := s.ClassesRepo.Get(ctx, createParams.ClassID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("could not get class: %w", err)
	}

	if class.CurrentCapacity == 0 {
		return uuid.Nil, domainErrors.ErrClassFullyBooked(
			createParams.ClassID,
			fmt.Errorf("no spots left in class with id: %d", createParams.ClassID),
		)
	}

	if class.StartTime.Before(time.Now()) {
		return uuid.Nil, domainErrors.ErrExpiredClassBooking(
			createParams.ClassID,
			fmt.Errorf("class %s has expired at %v", createParams.ClassID, class.StartTime),
		)
	}

	confirmationToken, err := s.TokenGenerator.Generate(32)
	if err != nil {
		return uuid.Nil, fmt.Errorf("could not generate confirmation token: %w", err)
	}

	pendingBooking := models.PendingOperation{
		ID:                uuid.New(),
		ClassID:           createParams.ClassID,
		Operation:         models.CreateBooking,
		Email:             createParams.Email,
		FirstName:         createParams.FirstName,
		LastName:          &createParams.LastName,
		ConfirmationToken: confirmationToken,
		CreatedAt:         time.Now().UTC(),
	}

	err = s.PendingOperationsRepo.Insert(ctx, pendingBooking)
	if err != nil {
		return uuid.Nil, fmt.Errorf("could not insert pending booking: %w", err)
	}

	msgParams := models.ConfirmationCreateParams{
		RecipientEmail:         createParams.Email,
		RecipientName:          createParams.FirstName,
		ConfirmationCreateLink: fmt.Sprintf("%s/confirmation/create_booking?token=%s", s.DomainAddr, confirmationToken),
	}

	err = s.MessageSender.SendConfirmationCreateLink(msgParams)
	if err != nil {
		return uuid.Nil, fmt.Errorf("could not send confirmation create link: %w", err)
	}

	return class.ID, nil
}

func (s *Service) CancelBooking(
	ctx context.Context,
	cancelParams models.CancelParams,
) (uuid.UUID, error) {
	confirmedBooking, err := s.ConfirmedBookingsRepo.Get(ctx, cancelParams.ClassID, cancelParams.Email)
	if err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			return uuid.Nil, domainErrors.ErrConfirmedBookingNotFound(
				cancelParams.ClassID, cancelParams.Email,
				fmt.Errorf("no such booking with email %s for class %v",
					cancelParams.Email, cancelParams.ClassID),
			)
		}

		return uuid.Nil, fmt.Errorf("could not get confirmed booking: %w", err)
	}

	err = s.validatePendingOperationNumberPerUser(ctx, cancelParams.ClassID, cancelParams.Email, models.CancelBooking)
	if err != nil {
		return uuid.Nil, fmt.Errorf("validate pending operation number per user: %w", err)
	}

	class, err := s.ClassesRepo.Get(ctx, cancelParams.ClassID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("could not get class: %w", err)
	}

	if class.StartTime.Before(time.Now()) {
		return uuid.Nil, domainErrors.ErrExpiredClassBooking(
			cancelParams.ClassID,
			fmt.Errorf("class %s has expired at %v", cancelParams.ClassID, class.StartTime),
		)
	}

	if class.CurrentCapacity == class.MaxCapacity {
		return uuid.Nil, domainErrors.ErrClassEmpty(
			cancelParams.ClassID,
			fmt.Errorf("class max capacity: %d exceeded", class.MaxCapacity))
	}

	confirmationToken, err := s.TokenGenerator.Generate(32)
	if err != nil {
		return uuid.Nil, fmt.Errorf("could not generate confirmation token: %w", err)
	}

	cancelPendingOperation := models.PendingOperation{
		ID:                uuid.New(),
		ClassID:           class.ID,
		Operation:         models.CancelBooking,
		Email:             cancelParams.Email,
		FirstName:         confirmedBooking.FirstName,
		ConfirmationToken: confirmationToken,
		CreatedAt:         time.Now().UTC(),
	}

	err = s.PendingOperationsRepo.Insert(ctx, cancelPendingOperation)
	if err != nil {
		return uuid.Nil, fmt.Errorf("could not insert pending booking: %w", err)
	}

	msgParams := models.ConfirmationCancelParams{
		RecipientEmail:         cancelParams.Email,
		RecipientName:          confirmedBooking.FirstName,
		ConfirmationCancelLink: fmt.Sprintf("%s/confirmation/cancel_booking?token=%s", s.DomainAddr, confirmationToken),
	}

	err = s.MessageSender.SendConfirmationCancelLink(msgParams)
	if err != nil {
		return uuid.Nil, fmt.Errorf("could not send confirmation cancel link: %w", err)
	}

	return class.ID, nil
}

func (s *Service) validatePendingOperationNumberPerUser(
	ctx context.Context,
	classID uuid.UUID,
	email string,
	operation models.Operation,
) error {
	count, err := s.PendingOperationsRepo.CountPendingOperationsPerUser(ctx, email, operation, classID)
	if err != nil {
		return fmt.Errorf("could not count pending operations for email: %s, error: %w", email, err)
	}

	if count >= 2 {
		return domainErrors.ErrTooManyPendingOperations(
			classID,
			email,
			fmt.Errorf("found %d pending operations per user", count),
		)
	}

	return nil
}
