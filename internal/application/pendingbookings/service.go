package pendingbookings

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
	ClassesRepo         repositories.IClasses
	PendingBookingsRepo repositories.IPendingBookings
	BookingsRepo        repositories.IBookings
	TokenGenerator      services.ITokenGenerator
	MessageSender       services.ISender
	DomainAddr          string
}

func NewService(
	classesRepo repositories.IClasses,
	pendingBookingsRepo repositories.IPendingBookings,
	bookingsRepo repositories.IBookings,
	tokenGenerator services.ITokenGenerator,
	messageSender services.ISender,
	domainAddr string,
) *Service {
	return &Service{
		ClassesRepo:         classesRepo,
		PendingBookingsRepo: pendingBookingsRepo,
		BookingsRepo:        bookingsRepo,
		TokenGenerator:      tokenGenerator,
		MessageSender:       messageSender,
		DomainAddr:          domainAddr,
	}
}

func (s *Service) CreatePendingBooking(
	ctx context.Context,
	pendingBookingParams models.PendingBookingParams,
) (uuid.UUID, error) {
	_, err := s.BookingsRepo.GetByEmailAndClassID(ctx, pendingBookingParams.ClassID, pendingBookingParams.Email)
	if err == nil {
		return uuid.Nil, domainErrors.ErrBookingAlreadyExists(pendingBookingParams.ClassID, pendingBookingParams.Email, err)
	}

	if !errors.Is(err, errs.ErrNotFound) {
		return uuid.Nil, fmt.Errorf("could not get booking: %w", err)
	}

	err = s.validatePendingBookingsPerUser(ctx, pendingBookingParams.ClassID, pendingBookingParams.Email, models.CreateBooking)
	if err != nil {
		return uuid.Nil, fmt.Errorf("validation failed for pending booking: %w", err)
	}

	class, err := s.ClassesRepo.Get(ctx, pendingBookingParams.ClassID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("could not get class: %w", err)
	}

	if class.CurrentCapacity == 0 {
		return uuid.Nil, domainErrors.ErrClassFullyBooked(
			pendingBookingParams.ClassID,
			fmt.Errorf("no spots left in class with id: %d", pendingBookingParams.ClassID),
		)
	}

	if class.StartTime.Before(time.Now()) {
		return uuid.Nil, domainErrors.ErrClassExpired(
			pendingBookingParams.ClassID,
			fmt.Errorf("class %s has expired at %v", pendingBookingParams.ClassID, class.StartTime),
		)
	}

	confirmationToken, err := s.TokenGenerator.Generate(32)
	if err != nil {
		return uuid.Nil, fmt.Errorf("could not generate confirmation token: %w", err)
	}

	pendingBooking := models.PendingBooking{
		ID:                uuid.New(),
		ClassID:           pendingBookingParams.ClassID,
		Operation:         models.CreateBooking,
		Email:             pendingBookingParams.Email,
		FirstName:         pendingBookingParams.FirstName,
		LastName:          &pendingBookingParams.LastName,
		ConfirmationToken: confirmationToken,
		CreatedAt:         time.Now().UTC(),
	}

	err = s.PendingBookingsRepo.Insert(ctx, pendingBooking)
	if err != nil {
		return uuid.Nil, fmt.Errorf("could not insert pending booking: %w", err)
	}

	msg := models.ConfirmationCreateMsg{
		RecipientEmail:         pendingBookingParams.Email,
		RecipientFirstName:     pendingBookingParams.FirstName,
		ConfirmationCreateLink: fmt.Sprintf("%s/bookings?token=%s", s.DomainAddr, confirmationToken),
	}

	err = s.MessageSender.SendConfirmationCreateLink(msg)
	if err != nil {
		return uuid.Nil, fmt.Errorf("could not send confirmation create link: %w", err)
	}

	return class.ID, nil
}

func (s *Service) CancelPendingBooking(
	ctx context.Context,
	cancelParams models.CancelBookingParams,
) (uuid.UUID, error) {
	confirmedBooking, err := s.BookingsRepo.GetByEmailAndClassID(ctx, cancelParams.ClassID, cancelParams.Email)
	if err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			return uuid.Nil, domainErrors.ErrBookingNotFound(
				cancelParams.ClassID, cancelParams.Email,
				fmt.Errorf("no such booking with email %s for class %v",
					cancelParams.Email, cancelParams.ClassID),
			)
		}

		return uuid.Nil, fmt.Errorf("could not get confirmed booking: %w", err)
	}

	err = s.validatePendingBookingsPerUser(ctx, cancelParams.ClassID, cancelParams.Email, models.CancelBooking)
	if err != nil {
		return uuid.Nil, fmt.Errorf("validation failed for pending booking: %w", err)
	}

	class, err := s.ClassesRepo.Get(ctx, cancelParams.ClassID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("could not get class: %w", err)
	}

	if class.StartTime.Before(time.Now()) {
		return uuid.Nil, domainErrors.ErrClassExpired(
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

	cancelPendingOperation := models.PendingBooking{
		ID:                uuid.New(),
		ClassID:           class.ID,
		Operation:         models.CancelBooking,
		Email:             cancelParams.Email,
		FirstName:         confirmedBooking.FirstName,
		ConfirmationToken: confirmationToken,
		CreatedAt:         time.Now().UTC(),
	}

	err = s.PendingBookingsRepo.Insert(ctx, cancelPendingOperation)
	if err != nil {
		return uuid.Nil, fmt.Errorf("could not insert pending booking: %w", err)
	}

	msg := models.ConfirmationCancelMsg{
		RecipientEmail:         cancelParams.Email,
		RecipientFirstName:     confirmedBooking.FirstName,
		ConfirmationCancelLink: fmt.Sprintf("%s/confirmation/cancel_booking?token=%s", s.DomainAddr, confirmationToken),
	}

	err = s.MessageSender.SendConfirmationCancelLink(msg)
	if err != nil {
		return uuid.Nil, fmt.Errorf("could not send confirmation cancel link: %w", err)
	}

	return class.ID, nil
}

func (s *Service) validatePendingBookingsPerUser(
	ctx context.Context,
	classID uuid.UUID,
	email string,
	operation models.Operation,
) error {
	count, err := s.PendingBookingsRepo.CountPendingBookingsPerUser(ctx, email, operation, classID)
	if err != nil {
		return fmt.Errorf("could not count pending bookings for email: %s, error: %w", email, err)
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
