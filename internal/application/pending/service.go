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
	TokenGenerator        services.Token
	MessageSender         services.Message
	DomainAddr            string
}

func New(
	classesRepo repositories.Classes,
	pendingOperationsRepo repositories.PendingOperations,
	tokenGenerator services.Token,
	messageSender services.Message,
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

	if class.MaxCapacity == 0 {
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

	msgParams := models.ConfirmationMsgParams{
		RecipientEmail:   createParams.Email,
		RecipientName:    createParams.FirstName,
		ConfirmationLink: fmt.Sprintf("%s/confirmation/create_booking?token=%s", s.DomainAddr, confirmationToken),
	}

	err = s.MessageSender.SendConfirmationLink(msgParams)
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

	msgParams := models.ConfirmationMsgParams{
		RecipientEmail:   cancelParams.Email,
		RecipientName:    cancelParams.FirstName,
		ConfirmationLink: fmt.Sprintf("%s/confirmation/cancel_booking?token=%s", s.DomainAddr, confirmationToken),
	}

	err = s.MessageSender.SendConfirmationLink(msgParams)
	if err != nil {
		return uuid.UUID{}, &errs2.BookingError{Code: http.StatusInternalServerError, Message: err.Error()}
	}

	return class.ID, nil

}
