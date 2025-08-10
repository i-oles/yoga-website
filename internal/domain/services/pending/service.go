package pending

import (
	"context"
	"fmt"
	"main/internal/domain/models"
	"main/internal/domain/repositories"
	"main/internal/errs"
	"main/internal/generator"
	"main/internal/sender"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type Service struct {
	ClassesRepo           repositories.Classes
	PendingOperationsRepo repositories.PendingOperations
	TokenGenerator        generator.Token
	MessageSender         sender.Message
	DomainAddr            string
}

func New(
	classesRepo repositories.Classes,
	pendingOperationsRepo repositories.PendingOperations,
	tokenGenerator generator.Token,
	messageSender sender.Message,
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
		return uuid.UUID{}, errs.ErrClassFullyBooked(fmt.Errorf("no spots left in class with id: %d", class.ID))
	}

	confirmationToken, err := s.TokenGenerator.Generate(32)
	if err != nil {
		return uuid.UUID{}, &errs.BookingError{Code: http.StatusInternalServerError, Message: err.Error()}
	}

	//authToken = fmt.Sprintf("%d", rand.Int())
	//fmt.Printf("token: %s\n", authToken)
	//

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
		return uuid.UUID{}, &errs.BookingError{Code: http.StatusInternalServerError, Message: err.Error()}
	}

	bookingConfirmationData := sender.ConfirmationData{
		RecipientEmail:   createParams.Email,
		RecipientName:    createParams.FirstName,
		ConfirmationLink: fmt.Sprintf("%s/confirmation/create_booking?token=%s", s.DomainAddr, confirmationToken),
	}

	err = s.MessageSender.SendConfirmationLink(bookingConfirmationData)
	if err != nil {
		return uuid.UUID{}, &errs.BookingError{Code: http.StatusInternalServerError, Message: err.Error()}
	}

	return class.ID, nil
}

func (s *Service) CancelBooking(ctx context.Context, cancelParams models.CancelParams) (uuid.UUID, error) {
	class, err := s.ClassesRepo.Get(ctx, cancelParams.ClassID)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("could not get class: %w", err)
	}

	confirmationToken, err := s.TokenGenerator.Generate(32)
	if err != nil {
		return uuid.UUID{}, &errs.BookingError{Code: http.StatusInternalServerError, Message: err.Error()}
	}

	//token = fmt.Sprintf("test%d", rand.Int())
	//fmt.Printf("token: %s\n", token)

	expiry := time.Now().Add(24 * time.Hour)

	cancelPendingOperation := models.PendingOperation{
		ID:             uuid.New(),
		ClassID:        class.ID,
		Operation:      models.CancelBooking,
		Email:          cancelParams.Email,
		FirstName:      cancelParams.FirstName,
		AuthToken:      confirmationToken,
		TokenExpiresAt: expiry,
	}

	err = s.PendingOperationsRepo.Insert(ctx, cancelPendingOperation)
	if err != nil {
		return uuid.UUID{}, &errs.BookingError{Code: http.StatusInternalServerError, Message: err.Error()}
	}

	cancelBookingConfirmationData := sender.ConfirmationData{
		RecipientEmail:   cancelParams.Email,
		RecipientName:    cancelParams.FirstName,
		ConfirmationLink: fmt.Sprintf("%s/confirmation/cancel_booking?token=%s", s.DomainAddr, confirmationToken),
	}

	err = s.MessageSender.SendConfirmationLink(cancelBookingConfirmationData)
	if err != nil {
		return uuid.UUID{}, &errs.BookingError{Code: http.StatusInternalServerError, Message: err.Error()}
	}

	return class.ID, nil

}
