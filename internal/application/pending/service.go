package pending

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	domainErrors "main/internal/domain/errs"
	"main/internal/domain/models"
	"main/internal/domain/repositories"
	"main/internal/domain/services"
	"net/http"
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
	class, err := s.ClassesRepo.Get(ctx, createParams.ClassID)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("could not get class: %w", err)
	}

	if class.CurrentCapacity == 0 {
		//TODO: verify this message error on front
		return uuid.UUID{}, domainErrors.ErrClassFullyBooked(fmt.Errorf("no spots left in class with id: %d", class.ID))
	}

	if class.StartTime.Before(time.Now()) {
		return uuid.UUID{}, domainErrors.ErrExpiredClassBooking(createParams.ClassID)
	}

	confirmationToken, err := s.TokenGenerator.Generate(32)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("could not generate confirmation token: %w", err)
	}

	pendingBooking := models.PendingOperation{
		ID:                uuid.New(),
		ClassID:           class.ID,
		Operation:         models.CreateBooking,
		Email:             createParams.Email,
		FirstName:         createParams.FirstName,
		LastName:          &createParams.LastName,
		ConfirmationToken: confirmationToken,
		CreatedAt:         time.Now(),
	}

	err = s.PendingOperationsRepo.Insert(ctx, pendingBooking)
	if err != nil {
		return uuid.UUID{}, &domainErrors.BookingError{Code: http.StatusInternalServerError, Message: err.Error()}
	}

	msgParams := models.ConfirmationCreateParams{
		RecipientEmail:         createParams.Email,
		RecipientName:          createParams.FirstName,
		ConfirmationCreateLink: fmt.Sprintf("%s/confirmation/create_booking?token=%s", s.DomainAddr, confirmationToken),
	}

	err = s.MessageSender.SendConfirmationCreateLink(msgParams)
	if err != nil {
		return uuid.UUID{}, &domainErrors.BookingError{Code: http.StatusInternalServerError, Message: err.Error()}
	}

	return class.ID, nil
}

func (s *Service) CancelBooking(
	ctx context.Context,
	cancelParams models.CancelParams,
) (uuid.UUID, error) {
	_, err := s.ConfirmedBookingsRepo.Get(ctx, cancelParams.ClassID, cancelParams.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return uuid.UUID{}, domainErrors.ErrConfirmedBookingNotFound(cancelParams.Email, cancelParams.ClassID)
		}

		return uuid.UUID{}, fmt.Errorf("could not get confirmed bookings: %w", err)
	}

	class, err := s.ClassesRepo.Get(ctx, cancelParams.ClassID)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("could not get class: %w", err)
	}

	if class.StartTime.Before(time.Now()) {
		return uuid.UUID{}, domainErrors.ErrExpiredClassBooking(cancelParams.ClassID)
	}

	if class.CurrentCapacity == class.MaxCapacity {
		return uuid.UUID{}, domainErrors.ErrClassEmpty(fmt.Errorf("max capacity exceeded"))
	}

	confirmationToken, err := s.TokenGenerator.Generate(32)
	if err != nil {
		return uuid.UUID{}, &domainErrors.BookingError{Code: http.StatusInternalServerError, Message: err.Error()}
	}

	cancelPendingOperation := models.PendingOperation{
		ID:                uuid.New(),
		ClassID:           class.ID,
		Operation:         models.CancelBooking,
		Email:             cancelParams.Email,
		FirstName:         cancelParams.FirstName,
		ConfirmationToken: confirmationToken,
		CreatedAt:         time.Now(),
	}

	err = s.PendingOperationsRepo.Insert(ctx, cancelPendingOperation)
	if err != nil {
		return uuid.UUID{}, &domainErrors.BookingError{Code: http.StatusInternalServerError, Message: err.Error()}
	}

	msgParams := models.ConfirmationCancelParams{
		RecipientEmail:         cancelParams.Email,
		RecipientName:          cancelParams.FirstName,
		ConfirmationCancelLink: fmt.Sprintf("%s/confirmation/cancel_booking?token=%s", s.DomainAddr, confirmationToken),
	}

	err = s.MessageSender.SendConfirmationCancelLink(msgParams)
	if err != nil {
		return uuid.UUID{}, &domainErrors.BookingError{Code: http.StatusInternalServerError, Message: err.Error()}
	}

	return class.ID, nil

}
