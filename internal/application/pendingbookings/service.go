package pendingbookings

import (
	"context"
	"errors"
	"fmt"
	"time"

	viewErrors "main/internal/domain/errs/view"
	"main/internal/domain/models"
	"main/internal/domain/repositories"
	"main/internal/domain/sender"
	"main/internal/domain/services"
	"main/internal/infrastructure/errs"

	"github.com/google/uuid"
)

const (
	allowedPendingBookingsLimit = 200
	tokenLength                 = 32
)

type service struct {
	ClassesRepo         repositories.IClasses
	PendingBookingsRepo repositories.IPendingBookings
	BookingsRepo        repositories.IBookings
	TokenGenerator      services.ITokenGenerator
	MessageSender       sender.ISender
	DomainAddr          string
}

func NewService(
	classesRepo repositories.IClasses,
	pendingBookingsRepo repositories.IPendingBookings,
	bookingsRepo repositories.IBookings,
	tokenGenerator services.ITokenGenerator,
	messageSender sender.ISender,
	domainAddr string,
) *service {
	return &service{
		ClassesRepo:         classesRepo,
		PendingBookingsRepo: pendingBookingsRepo,
		BookingsRepo:        bookingsRepo,
		TokenGenerator:      tokenGenerator,
		MessageSender:       messageSender,
		DomainAddr:          domainAddr,
	}
}

func (s *service) CreatePendingBooking(
	ctx context.Context,
	pendingBookingParams models.PendingBookingParams,
) (uuid.UUID, error) {
	_, err := s.BookingsRepo.GetByEmailAndClassID(
		ctx, pendingBookingParams.ClassID, pendingBookingParams.Email,
	)
	if err == nil {
		return uuid.Nil, viewErrors.ErrBookingAlreadyExists(
			pendingBookingParams.ClassID, pendingBookingParams.Email, err,
		)
	}

	if !errors.Is(err, errs.ErrNotFound) {
		return uuid.Nil, fmt.Errorf("could not get booking: %w", err)
	}

	err = s.ensurePendingBookingAvailability(
		ctx, pendingBookingParams.ClassID, pendingBookingParams.Email,
	)
	if err != nil {
		return uuid.Nil, fmt.Errorf("validation failed for pending booking: %w", err)
	}

	bookingCount, err := s.BookingsRepo.CountForClassID(ctx, pendingBookingParams.ClassID)
	if err != nil {
		return uuid.Nil,
			fmt.Errorf("could not count bookings for class %v: %w ", pendingBookingParams.ClassID, err)
	}

	class, err := s.ClassesRepo.Get(ctx, pendingBookingParams.ClassID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("could not get class: %w", err)
	}

	if bookingCount == class.MaxCapacity {
		return uuid.Nil, viewErrors.ErrClassFullyBooked(
			pendingBookingParams.ClassID,
			fmt.Errorf("no spots left in class with id: %d", pendingBookingParams.ClassID),
		)
	}

	if class.StartTime.Before(time.Now()) {
		return uuid.Nil, viewErrors.ErrClassExpired(
			pendingBookingParams.ClassID,
			fmt.Errorf("class %s has expired at %v", pendingBookingParams.ClassID, class.StartTime),
		)
	}

	confirmationToken, err := s.TokenGenerator.Generate(tokenLength)
	if err != nil {
		return uuid.Nil, fmt.Errorf("could not generate confirmation token: %w", err)
	}

	pendingBooking := models.PendingBooking{
		ID:                uuid.New(),
		ClassID:           pendingBookingParams.ClassID,
		Email:             pendingBookingParams.Email,
		FirstName:         pendingBookingParams.FirstName,
		LastName:          pendingBookingParams.LastName,
		ConfirmationToken: confirmationToken,
		CreatedAt:         time.Now().UTC(),
	}

	err = s.PendingBookingsRepo.Insert(ctx, pendingBooking)
	if err != nil {
		return uuid.Nil, fmt.Errorf("could not insert pending booking: %w", err)
	}

	err = s.MessageSender.SendLinkToConfirmation(
		pendingBookingParams.Email,
		pendingBookingParams.FirstName,
		fmt.Sprintf("%s/bookings?token=%s", s.DomainAddr, confirmationToken),
	)
	if err != nil {
		return uuid.Nil, fmt.Errorf("could not send confirmation create link: %w", err)
	}

	return class.ID, nil
}

func (s *service) ensurePendingBookingAvailability(
	ctx context.Context,
	classID uuid.UUID,
	email string,
) error {
	pendingBookings, err := s.PendingBookingsRepo.List(ctx)
	if err != nil {
		return fmt.Errorf("could not list pending bookings: %w", err)
	}

	if len(pendingBookings) >= allowedPendingBookingsLimit {
		return fmt.Errorf("limit: %d of pending bookings exceeded", allowedPendingBookingsLimit)
	}

	var count int

	for _, pendingBooking := range pendingBookings {
		if pendingBooking.Email == email && pendingBooking.ClassID == classID {
			count++
		}
	}

	if count >= 2 {
		return viewErrors.ErrTooManyPendingBookings(
			classID,
			email,
			fmt.Errorf("found %d pending operations per user", count),
		)
	}

	return nil
}
