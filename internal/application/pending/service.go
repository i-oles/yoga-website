package pending

import (
	"context"
	"fmt"
	errs2 "main/internal/domain/errs"
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
	TokenGenerator        services.ITokenGenerator
	MessageSender         services.ISender
	DomainAddr            string
}

func New(
	classesRepo repositories.Classes,
	pendingOperationsRepo repositories.PendingOperations,
	tokenGenerator services.ITokenGenerator,
	messageSender services.ISender,
	domainAddr string,
) *Service {
	return &Service{
		ClassesRepo:           classesRepo,
		PendingOperationsRepo: pendingOperationsRepo,
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
		return uuid.UUID{}, errs2.ErrClassFullyBooked(fmt.Errorf("no spots left in class with id: %d", class.ID))
	}

	confirmationToken, err := s.TokenGenerator.Generate(32)
	if err != nil {
		return uuid.UUID{}, &errs2.BookingError{Code: http.StatusInternalServerError, Message: err.Error()}
	}

	expiry := time.Now().Add(24 * time.Hour)

	pendingBooking := models.PendingOperation{
		ID:             uuid.New(),
		ClassID:        class.ID,
		Operation:      models.CreateBooking,
		Email:          createParams.Email,
		FirstName:      createParams.FirstName,
		LastName:       &createParams.LastName,
		AuthToken:      confirmationToken,
		TokenExpiresAt: expiry,
		CreatedAt:      time.Now(),
	}

	err = s.PendingOperationsRepo.Insert(ctx, pendingBooking)
	if err != nil {
		return uuid.UUID{}, &errs2.BookingError{Code: http.StatusInternalServerError, Message: err.Error()}
	}

	msgParams := models.ConfirmationCreateParams{
		RecipientEmail:         createParams.Email,
		RecipientName:          createParams.FirstName,
		ConfirmationCreateLink: fmt.Sprintf("%s/confirmation/create_booking?token=%s", s.DomainAddr, confirmationToken),
	}

	err = s.MessageSender.SendConfirmationCreateLink(msgParams)
	if err != nil {
		return uuid.UUID{}, &errs2.BookingError{Code: http.StatusInternalServerError, Message: err.Error()}
	}

	return class.ID, nil
}

func (s *Service) CancelBooking(
	ctx context.Context,
	cancelParams models.CancelParams,
) (uuid.UUID, error) {
	class, err := s.ClassesRepo.Get(ctx, cancelParams.ClassID)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("could not get class: %w", err)
	}

	if class.CurrentCapacity == class.MaxCapacity {
		return uuid.UUID{}, errs2.ErrClassEmpty(fmt.Errorf("max capacity exceeded"))
	}

	confirmationToken, err := s.TokenGenerator.Generate(32)
	if err != nil {
		return uuid.UUID{}, &errs2.BookingError{Code: http.StatusInternalServerError, Message: err.Error()}
	}

	expiry := time.Now().Add(24 * time.Hour)

	cancelPendingOperation := models.PendingOperation{
		ID:             uuid.New(),
		ClassID:        class.ID,
		Operation:      models.CancelBooking,
		Email:          cancelParams.Email,
		FirstName:      cancelParams.FirstName,
		AuthToken:      confirmationToken,
		TokenExpiresAt: expiry,
		CreatedAt:      time.Now(),
	}

	err = s.PendingOperationsRepo.Insert(ctx, cancelPendingOperation)
	if err != nil {
		return uuid.UUID{}, &errs2.BookingError{Code: http.StatusInternalServerError, Message: err.Error()}
	}

	msgParams := models.ConfirmationCancelParams{
		RecipientEmail:         cancelParams.Email,
		RecipientName:          cancelParams.FirstName,
		ConfirmationCancelLink: fmt.Sprintf("%s/confirmation/cancel_booking?token=%s", s.DomainAddr, confirmationToken),
	}

	err = s.MessageSender.SendConfirmationCancelLink(msgParams)
	if err != nil {
		return uuid.UUID{}, &errs2.BookingError{Code: http.StatusInternalServerError, Message: err.Error()}
	}

	return class.ID, nil

}
