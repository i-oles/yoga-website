package bookings

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	sharedErrors "main/internal/domain/errs"
	viewErrors "main/internal/domain/errs/view"
	"main/internal/domain/models"
	"main/internal/domain/notifier"
	"main/internal/domain/repositories"
	"main/internal/infrastructure/errs"
	"main/pkg/tools"

	"github.com/google/uuid"
)

type service struct {
	ClassesRepo         repositories.IClasses
	BookingsRepo        repositories.IBookings
	PendingBookingsRepo repositories.IPendingBookings
	PassesRepo          repositories.IPasses
	Notifier            notifier.INotifier
	DomainAddr          string
}

func NewService(
	classesRepo repositories.IClasses,
	bookingsRepo repositories.IBookings,
	pendingBookingsRepo repositories.IPendingBookings,
	passesRepo repositories.IPasses,
	notifier notifier.INotifier,
	domainAddr string,
) *service {
	return &service{
		ClassesRepo:         classesRepo,
		BookingsRepo:        bookingsRepo,
		PendingBookingsRepo: pendingBookingsRepo,
		PassesRepo:          passesRepo,
		Notifier:            notifier,
		DomainAddr:          domainAddr,
	}
}

func (s *service) CreateBooking(ctx context.Context, token string) (models.Class, error) {
	pendingBooking, err := s.PendingBookingsRepo.GetByConfirmationToken(ctx, token)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.Class{}, viewErrors.ErrPendingBookingNotFound(
				fmt.Errorf("pending booking for token: %s not found", token),
			)
		}

		return models.Class{}, fmt.Errorf("could not get pending booking: %w", err)
	}

	_, err = s.BookingsRepo.GetByEmailAndClassID(ctx, pendingBooking.ClassID, pendingBooking.Email)
	if err == nil {
		return models.Class{},
			viewErrors.ErrBookingAlreadyExists(pendingBooking.ClassID, pendingBooking.Email, err)
	}

	class, err := s.ClassesRepo.Get(ctx, pendingBooking.ClassID)
	if err != nil {
		return models.Class{},
			fmt.Errorf("could not get class with id: %s, %w", pendingBooking.ClassID, err)
	}

	err = s.checkClassAvailability(ctx, class)
	if err != nil {
		return models.Class{}, fmt.Errorf("class unavailable: %w", err)
	}

	bookingID, err := s.createBooking(ctx, pendingBooking)
	if err != nil {
		return models.Class{},
			fmt.Errorf("could not create booking for pendingBooking %+v: %w", pendingBooking, err)
	}

	usedBookingIDs, totalBookings, err := s.incrementPassIfValid(ctx, pendingBooking.Email, bookingID)
	if err != nil {
		return models.Class{},
			fmt.Errorf("could not increment pass for %s: %w", pendingBooking.Email, err)
	}

	err = s.sendConfirmationEmails(
		pendingBooking, class, usedBookingIDs, totalBookings, token, bookingID,
	)
	if err != nil {
		return models.Class{},
			fmt.Errorf("could not send confirmation email %s: %w", pendingBooking.Email, err)
	}

	return class, nil
}

func (s *service) sendConfirmationEmails(
	pendingBooking models.PendingBooking,
	class models.Class,
	passUsedBookingIDs []uuid.UUID,
	passTotalBookingIDs *int,
	token string,
	bookingID uuid.UUID,
) error {
	notifierParams := models.NotifierParams{
		RecipientEmail:     pendingBooking.Email,
		RecipientFirstName: pendingBooking.FirstName,
		RecipientLastName:  pendingBooking.LastName,
		ClassName:          class.ClassName,
		ClassLevel:         class.ClassLevel,
		StartTime:          class.StartTime,
		Location:           class.Location,
		PassUsedBookingIDs: passUsedBookingIDs,
		PassTotalBookings:  passTotalBookingIDs,
	}

	cancellationLink := fmt.Sprintf(
		"%s/bookings/%s/cancel_form?token=%s", s.DomainAddr, bookingID, token,
	)

	err := s.Notifier.NotifyBookingConfirmation(notifierParams, cancellationLink)
	if err != nil {
		return fmt.Errorf("could not notify booking confirmation: %w", err)
	}

	return nil
}

func (s *service) createBooking(
	ctx context.Context, pendingBooking models.PendingBooking,
) (uuid.UUID, error) {
	booking := models.Booking{
		ID:                uuid.New(),
		ClassID:           pendingBooking.ClassID,
		FirstName:         pendingBooking.FirstName,
		LastName:          pendingBooking.LastName,
		Email:             pendingBooking.Email,
		CreatedAt:         time.Now().UTC(),
		ConfirmationToken: pendingBooking.ConfirmationToken,
	}

	bookingID, err := s.BookingsRepo.Insert(ctx, booking)
	if err != nil {
		return uuid.Nil, fmt.Errorf("could not insert booking: %w", err)
	}

	return bookingID, nil
}

func (s *service) incrementPassIfValid(
	ctx context.Context,
	email string,
	bookingID uuid.UUID,
) ([]uuid.UUID, *int, error) {
	passOpt, err := s.PassesRepo.GetByEmail(ctx, email)
	if err != nil && passOpt.Exists() {
		return nil, nil, fmt.Errorf("could not get pass: %w", err)
	}

	var updatedBookingIDs []uuid.UUID

	var totalBookings *int

	if passOpt.Exists() {
		pass := passOpt.Get()

		if len(pass.UsedBookingIDs)+1 <= pass.TotalBookings {
			updatedBookingIDs = pass.UsedBookingIDs
			updatedBookingIDs = append(updatedBookingIDs, bookingID)

			err = s.PassesRepo.Update(ctx, pass.ID, updatedBookingIDs, pass.TotalBookings)
			if err != nil {
				return nil, nil,
					fmt.Errorf("could not update pass for %s with %v, %d", email, updatedBookingIDs, pass.TotalBookings)
			}
		}

		totalBookings = &pass.TotalBookings
	}

	return updatedBookingIDs, totalBookings, nil
}

func (s *service) checkClassAvailability(ctx context.Context, class models.Class) error {
	if class.StartTime.Before(time.Now()) {
		return viewErrors.ErrClassExpired(class.ID, fmt.Errorf("class %s has expired at %v", class.ID, class.StartTime))
	}

	bookingCount, err := s.BookingsRepo.CountForClassID(ctx, class.ID)
	if err != nil {
		return fmt.Errorf("could not count bookings for class %v: %w ", class.ID, err)
	}

	if bookingCount == class.MaxCapacity {
		return viewErrors.ErrSomeoneBookedClassFaster(fmt.Errorf("max capacity of class %d exceeded", class.MaxCapacity))
	}

	return nil
}

func (s *service) CancelBooking(ctx context.Context, bookingID uuid.UUID, token string) error {
	booking, err := s.ensureBookingCancellationAllowed(ctx, bookingID, token)
	if err != nil {
		return fmt.Errorf("booking cancellation not allowed for bookingID %s: %w", bookingID, err)
	}

	err = s.BookingsRepo.Delete(ctx, bookingID)
	if err != nil {
		if errors.Is(err, errs.ErrNoRowsAffected) {
			return viewErrors.ErrBookingNotFound(
				booking.ClassID,
				booking.Email,
				fmt.Errorf("could not find booking with email %s for class %s", booking.Email, booking.ClassID),
			)
		}

		return fmt.Errorf("could not delete booking: %w", err)
	}

	usedBookingIDs, totalBookings, err := s.decrementPassIfValid(ctx, booking.Email, bookingID)
	if err != nil {
		return fmt.Errorf("could not dectemetnt pass for %s: %w", booking.Email, err)
	}

	notifierParams := models.NotifierParams{
		RecipientFirstName: booking.FirstName,
		RecipientLastName:  booking.LastName,
		RecipientEmail:     booking.Email,
		ClassName:          booking.Class.ClassName,
		ClassLevel:         booking.Class.ClassLevel,
		StartTime:          booking.Class.StartTime,
		Location:           booking.Class.Location,
		PassUsedBookingIDs: usedBookingIDs,
		PassTotalBookings:  totalBookings,
	}

	err = s.Notifier.NotifyBookingCancellation(notifierParams)
	if err != nil {
		return fmt.Errorf("could not notify booking cancellation with %+v: %w", notifierParams, err)
	}

	return nil
}

func (s *service) ensureBookingCancellationAllowed(
	ctx context.Context, bookingID uuid.UUID, token string,
) (models.Booking, error) {
	booking, err := s.BookingsRepo.GetByID(ctx, bookingID)
	if err != nil {
		return models.Booking{}, fmt.Errorf("could not get booking for id %s: %w", bookingID, err)
	}

	if booking.ConfirmationToken != token {
		return models.Booking{}, viewErrors.ErrInvalidCancellationLink(
			fmt.Errorf("cancel booking failed due to invalid token: %s for email: %s", booking.Email, token),
		)
	}

	if booking.Class == nil {
		return models.Booking{}, errors.New("booking.Class field should not be empty")
	}

	if booking.Class.StartTime.Before(time.Now()) {
		return models.Booking{}, viewErrors.ErrClassExpired(
			booking.Class.ID,
			fmt.Errorf("class %s has expired at %v", booking.ClassID, booking.Class.StartTime),
		)
	}

	return booking, nil
}

func (s *service) GetBookingForCancellation(
	ctx context.Context, bookingID uuid.UUID, token string,
) (models.Booking, error) {
	booking, err := s.BookingsRepo.GetByID(ctx, bookingID)
	if err != nil {
		return models.Booking{}, fmt.Errorf("could not get booking for id %s: %w", bookingID, err)
	}

	if booking.ConfirmationToken != token {
		return models.Booking{}, viewErrors.ErrInvalidCancellationLink(err)
	}

	return booking, nil
}

func (s *service) DeleteBooking(ctx context.Context, bookingID uuid.UUID) error {
	booking, err := s.BookingsRepo.GetByID(ctx, bookingID)
	if err != nil {
		return fmt.Errorf("could get booking for id %s: %w", bookingID, err)
	}

	if booking.Class == nil {
		return errors.New("booking.Class field should not be empty")
	}

	err = s.BookingsRepo.Delete(ctx, bookingID)
	if err != nil {
		return fmt.Errorf("could not delete booking for id %s: %w", bookingID, err)
	}

	if booking.Class.StartTime.Before(time.Now()) {
		return nil
	}

	usedBookingIDs, totalBookings, err := s.decrementPassIfValid(ctx, booking.Email, bookingID)
	if err != nil {
		return fmt.Errorf("could not dectemetnt pass for %s: %w", booking.Email, err)
	}

	notifierParams := models.NotifierParams{
		RecipientFirstName: booking.FirstName,
		RecipientLastName:  booking.LastName,
		RecipientEmail:     booking.Email,
		ClassName:          booking.Class.ClassName,
		ClassLevel:         booking.Class.ClassLevel,
		StartTime:          booking.Class.StartTime,
		Location:           booking.Class.Location,
		PassUsedBookingIDs: usedBookingIDs,
		PassTotalBookings:  totalBookings,
	}

	err = s.Notifier.NotifyBookingCancellation(notifierParams)
	if err != nil {
		return fmt.Errorf("could not nofify booking cancellation with %+v: %w", notifierParams, err)
	}

	return nil
}

func (s *service) decrementPassIfValid(
	ctx context.Context,
	email string,
	bookingID uuid.UUID,
) ([]uuid.UUID, *int, error) {
	passOpt, err := s.PassesRepo.GetByEmail(ctx, email)
	if err != nil && passOpt.Exists() {
		return nil, nil, fmt.Errorf("could not get pass: %w", err)
	}

	var updatedBookingIDs []uuid.UUID

	var totalBookings *int

	if passOpt.Exists() {
		pass := passOpt.Get()

		updatedBookingIDs, err = tools.RemoveFromSlice(pass.UsedBookingIDs, bookingID)
		if errors.Is(err, sharedErrors.ErrBookingIDNotFoundInPass) {
			return nil, nil, nil
		}

		if err != nil {
			return nil, nil, fmt.Errorf("could not remove bookingID %v from usedBookingIDs", bookingID)
		}

		err = s.PassesRepo.Update(ctx, pass.ID, updatedBookingIDs, pass.TotalBookings)
		if err != nil {
			return nil, nil,
				fmt.Errorf("could not update pass for %s with %v: %w", pass.Email, updatedBookingIDs, err)
		}

		totalBookings = &pass.TotalBookings
	}

	return updatedBookingIDs, totalBookings, nil
}
