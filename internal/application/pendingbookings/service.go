package pendingbookings

import (
	"context"
	"errors"
	"fmt"
	"time"

	viewErrors "main/internal/domain/errs/view"
	"main/internal/domain/models"
	"main/internal/domain/notifier"
	"main/internal/domain/repositories"
	"main/internal/domain/services"
	"main/internal/infrastructure/errs"

	"github.com/google/uuid"
)

const (
	allowedPendingBookingsLimit = 200
	tokenLength                 = 32
)

type service struct {
	classesRepo         repositories.IClasses
	pendingBookingsRepo repositories.IPendingBookings
	bookingsRepo        repositories.IBookings
	tokenGenerator      services.ITokenGenerator
	notifier            notifier.INotifier
	domainAddr          string
}

func NewService(
	classesRepo repositories.IClasses,
	pendingBookingsRepo repositories.IPendingBookings,
	bookingsRepo repositories.IBookings,
	tokenGenerator services.ITokenGenerator,
	notifier notifier.INotifier,
	domainAddr string,
) *service {
	return &service{
		classesRepo:         classesRepo,
		pendingBookingsRepo: pendingBookingsRepo,
		bookingsRepo:        bookingsRepo,
		tokenGenerator:      tokenGenerator,
		notifier:            notifier,
		domainAddr:          domainAddr,
	}
}

func (s *service) CreatePendingBooking(
	ctx context.Context,
	pendingBookingParams models.PendingBookingParams,
) (uuid.UUID, error) {
	err := s.ensurePendingBookingCreationAllowed(
		ctx, pendingBookingParams.ClassID, pendingBookingParams.Email,
	)
	if err != nil {
		return uuid.Nil, fmt.Errorf("pending booking creation not allowed: %w", err)
	}

	err = s.checkClassAvailability(ctx, pendingBookingParams.ClassID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("class not available: %w", err)
	}

	confirmationToken, err := s.tokenGenerator.Generate(tokenLength)
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

	err = s.pendingBookingsRepo.Insert(ctx, pendingBooking)
	if err != nil {
		return uuid.Nil, fmt.Errorf("could not insert pending booking: %w", err)
	}

	err = s.notifier.NotifyConfirmationLink(
		pendingBookingParams.Email,
		pendingBookingParams.FirstName,
		fmt.Sprintf("%s/bookings?token=%s", s.domainAddr, confirmationToken),
	)
	if err != nil {
		return uuid.Nil, fmt.Errorf("could not notify confirmation link: %w", err)
	}

	return pendingBookingParams.ClassID, nil
}

func (s *service) ensurePendingBookingCreationAllowed(
	ctx context.Context,
	classID uuid.UUID,
	email string,
) error {
	_, err := s.bookingsRepo.GetByEmailAndClassID(ctx, classID, email)
	if err == nil {
		return viewErrors.ErrBookingAlreadyExists(classID, email, err)
	}

	if !errors.Is(err, errs.ErrNotFound) {
		return fmt.Errorf("could not get booking: %w", err)
	}

	pendingBookings, err := s.pendingBookingsRepo.List(ctx)
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

func (s *service) checkClassAvailability(ctx context.Context, classID uuid.UUID) error {
	bookingCount, err := s.bookingsRepo.CountForClassID(ctx, classID)
	if err != nil {
		return fmt.Errorf("could not count bookings for class %v: %w ", classID, err)
	}

	class, err := s.classesRepo.Get(ctx, classID)
	if err != nil {
		return fmt.Errorf("could not get class: %w", err)
	}

	if bookingCount == class.MaxCapacity {
		return viewErrors.ErrClassFullyBooked(classID, fmt.Errorf("no spots left in class with id: %d", classID))
	}

	if class.StartTime.Before(time.Now()) {
		return viewErrors.ErrClassExpired(classID, fmt.Errorf("class %s has expired at %v", classID, class.StartTime))
	}

	return nil
}
