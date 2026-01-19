package bookings

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	domainErrors "main/internal/domain/errs"
	"main/internal/domain/models"
	"main/internal/domain/repositories"
	"main/internal/domain/sender"
	"main/internal/infrastructure/errs"

	"github.com/google/uuid"
)

type Service struct {
	ClassesRepo         repositories.IClasses
	BookingsRepo        repositories.IBookings
	PendingBookingsRepo repositories.IPendingBookings
	PassesRepo          repositories.IPasses
	MessageSender       sender.ISender
	DomainAddr          string
}

func NewService(
	classesRepo repositories.IClasses,
	bookingsRepo repositories.IBookings,
	pendingBookingsRepo repositories.IPendingBookings,
	passesRepo repositories.IPasses,
	messageSender sender.ISender,
	domainAddr string,
) *Service {
	return &Service{
		ClassesRepo:         classesRepo,
		BookingsRepo:        bookingsRepo,
		PendingBookingsRepo: pendingBookingsRepo,
		PassesRepo:          passesRepo,
		MessageSender:       messageSender,
		DomainAddr:          domainAddr,
	}
}

// CreateBooking TODO: this should return models.Booking with class field taken from relation.
func (s *Service) CreateBooking(ctx context.Context, token string) (models.Class, error) {
	pendingBooking, err := s.PendingBookingsRepo.GetByConfirmationToken(ctx, token)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.Class{}, domainErrors.ErrPendingBookingNotFound(
				fmt.Errorf("pending booking for token: %s not found", token),
			)
		}

		return models.Class{}, fmt.Errorf("could not get pending booking: %w", err)
	}

	_, err = s.BookingsRepo.GetByEmailAndClassID(ctx, pendingBooking.ClassID, pendingBooking.Email)
	if err == nil {
		return models.Class{},
			domainErrors.ErrBookingAlreadyExists(pendingBooking.ClassID, pendingBooking.Email,
				err,
			)
	}

	class, err := s.ClassesRepo.Get(ctx, pendingBooking.ClassID)
	if err != nil {
		return models.Class{}, fmt.Errorf(
			"could not get class with id: %s, %w", pendingBooking.ClassID, err,
		)
	}

	if class.StartTime.Before(time.Now()) {
		return models.Class{}, domainErrors.ErrClassExpired(
			class.ID,
			fmt.Errorf("class %s has expired at %v", pendingBooking.ClassID, class.StartTime),
		)
	}

	bookingCount, err := s.BookingsRepo.CountForClassID(ctx, class.ID)
	if err != nil {
		return models.Class{}, fmt.Errorf(
			"could not count bookings for class %v: %w ", class.ID, err,
		)
	}

	if bookingCount == class.MaxCapacity {
		return models.Class{}, domainErrors.ErrSomeoneBookedClassFaster(
			fmt.Errorf("max capacity of class %d exceeded", class.MaxCapacity),
		)
	}

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
		return models.Class{}, fmt.Errorf("could not insert booking: %w", err)
	}

	err = s.PendingBookingsRepo.Delete(ctx, pendingBooking.ID)
	if err != nil {
		return models.Class{}, fmt.Errorf("could not delete pending booking: %w", err)
	}

	cancellationLink := fmt.Sprintf("%s/bookings/%s/cancel_form?token=%s", s.DomainAddr, bookingID, token)

	msg := models.ConfirmationMsg{
		RecipientEmail:     pendingBooking.Email,
		RecipientFirstName: pendingBooking.FirstName,
		RecipientLastName:  pendingBooking.LastName,
		ClassName:          class.ClassName,
		ClassLevel:         class.ClassLevel,
		StartTime:          class.StartTime,
		Location:           class.Location,
		CancellationLink:   cancellationLink,
	}

	pass, err := s.PassesRepo.GetByEmail(ctx, pendingBooking.Email)
	if err != nil && err != errs.ErrNotFound {
		return models.Class{}, fmt.Errorf("could not get pass: %w", err)
	}

	if pass.Credits+1 <= pass.TotalCredits {
		newCredits := pass.Credits + 1
		msg.PassCredits = newCredits
		msg.TotalPassCredits = pass.TotalCredits

		update := map[string]any{
			"credits": newCredits,
		}

		err = s.PassesRepo.Update(ctx, update)
		if err != nil {
			return models.Class{}, fmt.Errorf("could not update pass for %s with %v", pendingBooking.Email, update)
		}
	}

	err = s.MessageSender.SendConfirmations(msg)
	if err != nil {
		return models.Class{}, fmt.Errorf("error while sending final-confirmation: %w", err)
	}

	return class, nil
}

func (s *Service) CancelBooking(ctx context.Context, bookingID uuid.UUID, token string) error {
	booking, err := s.BookingsRepo.GetByID(ctx, bookingID)
	if err != nil {
		return fmt.Errorf("could not get booking for id %s: %w", bookingID, err)
	}

	// TODO: do I need this check?
	if booking.ConfirmationToken != token {
		return domainErrors.ErrInvalidCancellationLink(
			fmt.Errorf("cancel booking failed due to invalid token: %s for email: %s", booking.Email, token),
		)
	}

	if booking.Class == nil {
		return errors.New("booking.Class field should not be empty")
	}

	if booking.Class.StartTime.Before(time.Now()) {
		return domainErrors.ErrClassExpired(
			booking.Class.ID,
			fmt.Errorf("class %s has expired at %v", booking.ClassID, booking.Class.StartTime),
		)
	}

	err = s.BookingsRepo.Delete(ctx, booking.ID)
	if err != nil {
		if errors.Is(err, errs.ErrNoRowsAffected) {
			return domainErrors.ErrBookingNotFound(
				booking.ClassID,
				booking.Email,
				fmt.Errorf("could not find booking with email %s for class %s",
					booking.Email,
					booking.ClassID,
				),
			)
		}

		return fmt.Errorf("could not delete booking: %w", err)
	}

	class, err := s.ClassesRepo.Get(ctx, booking.ClassID)
	if err != nil {
		return fmt.Errorf("could not get class: %w", err)
	}

	err = s.MessageSender.SendInfoAboutCancellationToOwner(
		booking.FirstName, booking.LastName, class.StartTime)
	if err != nil {
		return fmt.Errorf("could not send info about cancellation to owner: %w", err)
	}

	return nil
}

func (s *Service) GetBookingForCancellation(
	ctx context.Context, bookingID uuid.UUID, token string,
) (models.Booking, error) {
	booking, err := s.BookingsRepo.GetByID(ctx, bookingID)
	if err != nil {
		return models.Booking{}, fmt.Errorf("could not get booking for id %s: %w", bookingID, err)
	}

	if booking.ConfirmationToken != token {
		return models.Booking{}, domainErrors.ErrInvalidCancellationLink(err)
	}

	return booking, nil
}

func (s *Service) DeleteBooking(ctx context.Context, bookingID uuid.UUID) error {
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

	if booking.Class.StartTime.After(time.Now()) {
		err = s.MessageSender.SendInfoAboutBookingCancellation(
			booking.Email, booking.FirstName, *booking.Class,
		)
		if err != nil {
			return fmt.Errorf("could not send info about booking cancellation to %s: %w", booking.Email, err)
		}
	}

	return nil
}
