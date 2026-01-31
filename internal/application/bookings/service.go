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
	"main/internal/domain/repositories"
	"main/internal/domain/sender"
	"main/internal/infrastructure/errs"
	"main/pkg/optional"
	"main/pkg/tools"

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
			return models.Class{}, viewErrors.ErrPendingBookingNotFound(
				fmt.Errorf("pending booking for token: %s not found", token),
			)
		}

		return models.Class{}, fmt.Errorf("could not get pending booking: %w", err)
	}

	_, err = s.BookingsRepo.GetByEmailAndClassID(ctx, pendingBooking.ClassID, pendingBooking.Email)
	if err == nil {
		return models.Class{},
			viewErrors.ErrBookingAlreadyExists(pendingBooking.ClassID, pendingBooking.Email,
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
		return models.Class{}, viewErrors.ErrClassExpired(
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
		return models.Class{}, viewErrors.ErrSomeoneBookedClassFaster(
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

	senderParams := models.SenderParams{
		RecipientEmail:     pendingBooking.Email,
		RecipientFirstName: pendingBooking.FirstName,
		RecipientLastName:  &pendingBooking.LastName,
		ClassName:          class.ClassName,
		ClassLevel:         class.ClassLevel,
		StartTime:          class.StartTime,
		Location:           class.Location,
	}

	passOpt, err := s.PassesRepo.GetByEmail(ctx, pendingBooking.Email)
	if err != nil && passOpt.Exists() {
		return models.Class{}, fmt.Errorf("could not get pass: %w", err)
	}

	pass := passOpt.Get()

	if len(pass.UsedBookingIDs)+1 <= pass.TotalBookings {
		updatedBookingIDs := pass.UsedBookingIDs
		updatedBookingIDs = append(updatedBookingIDs, bookingID)

		err = s.PassesRepo.Update(ctx, pass.ID, updatedBookingIDs, pass.TotalBookings)
		if err != nil {
			return models.Class{}, fmt.Errorf("could not update pass for %s with %v, %d", pendingBooking.Email, updatedBookingIDs, pass.TotalBookings)
		}

		senderParams.PassUsedBookingIDs = updatedBookingIDs
		senderParams.PassTotalBookings = &pass.TotalBookings
	}

	cancellationLink := fmt.Sprintf("%s/bookings/%s/cancel_form?token=%s", s.DomainAddr, bookingID, token)

	err = s.MessageSender.SendConfirmations(senderParams, cancellationLink)
	if err != nil {
		return models.Class{}, fmt.Errorf("error while sending final-confirmation: %w", err)
	}

	err = s.PendingBookingsRepo.Delete(ctx, pendingBooking.ID)
	if err != nil {
		return models.Class{}, fmt.Errorf("could not delete pending booking: %w", err)
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
		return viewErrors.ErrInvalidCancellationLink(
			fmt.Errorf("cancel booking failed due to invalid token: %s for email: %s", booking.Email, token),
		)
	}

	if booking.Class == nil {
		return errors.New("booking.Class field should not be empty")
	}

	if booking.Class.StartTime.Before(time.Now()) {
		return viewErrors.ErrClassExpired(
			booking.Class.ID,
			fmt.Errorf("class %s has expired at %v", booking.ClassID, booking.Class.StartTime),
		)
	}

	err = s.BookingsRepo.Delete(ctx, booking.ID)
	if err != nil {
		if errors.Is(err, errs.ErrNoRowsAffected) {
			return viewErrors.ErrBookingNotFound(
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

	passOpt, err := s.PassesRepo.GetByEmail(ctx, booking.Email)
	if err != nil {
		return fmt.Errorf("could not get users passOpt for %s: %w", booking.Email, err)
	}

	senderParams := models.SenderParams{
		RecipientFirstName: booking.FirstName,
		RecipientEmail:     booking.Email,
		ClassName:          class.ClassName,
		ClassLevel:         class.ClassLevel,
		StartTime:          class.StartTime,
		Location:           class.Location,
	}

	if passOpt.Exists() {
		senderParams, err = s.updateSenderParamsWithPass(ctx, bookingID, passOpt, senderParams)
		if err != nil {
			return fmt.Errorf("could not send info about cancellation to %s: %w", booking.Email, err)
		}
	}

	err = s.MessageSender.SendInfoAboutBookingCancellation(senderParams)
	if err != nil {
		return fmt.Errorf("could not send info about cancellation to %s: %w", booking.Email, err)
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
		return models.Booking{}, viewErrors.ErrInvalidCancellationLink(err)
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

	if booking.Class.StartTime.Before(time.Now()) {
		return nil
	}

	passOpt, err := s.PassesRepo.GetByEmail(ctx, booking.Email)
	if err != nil {
		return fmt.Errorf("could not get users passOpt for: %s", booking.Email)
	}

	senderParams := models.SenderParams{
		RecipientFirstName: booking.FirstName,
		RecipientEmail:     booking.Email,
		ClassName:          booking.Class.ClassName,
		ClassLevel:         booking.Class.ClassLevel,
		StartTime:          booking.Class.StartTime,
		Location:           booking.Class.Location,
	}

	if passOpt.Exists() {
		senderParams, err = s.updateSenderParamsWithPass(ctx, bookingID, passOpt, senderParams)
		if err != nil {
			return fmt.Errorf("could not send info about cancellation to %s: %w", booking.Email, err)
		}
	}

	err = s.MessageSender.SendInfoAboutBookingCancellation(senderParams)
	if err != nil {
		return fmt.Errorf("could not send info about cancellation to %s: %w", booking.Email, err)
	}

	return nil
}

func (s Service) updateSenderParamsWithPass(
	ctx context.Context,
	bookingID uuid.UUID,
	passOpt optional.Optional[models.Pass],
	senderParams models.SenderParams,
) (models.SenderParams, error) {
	pass := passOpt.Get()

	if len(pass.UsedBookingIDs) > 0 {
		updatedBookingIDs, err := tools.RemoveFromSlice(pass.UsedBookingIDs, bookingID)
		if errors.Is(err, sharedErrors.ErrBookingIDNotFoundInPass) {
			return senderParams, nil
		}

		if err != nil {
			return models.SenderParams{}, fmt.Errorf("could not remove bookingID %v from usedBookingIDs", bookingID)
		}

		err = s.PassesRepo.Update(ctx, pass.ID, updatedBookingIDs, pass.TotalBookings)
		if err != nil {
			return models.SenderParams{},
				fmt.Errorf("could not update pass for %s with %v: %w", pass.Email, updatedBookingIDs, err)
		}

		senderParams.PassUsedBookingIDs = updatedBookingIDs
		senderParams.PassTotalBookings = &pass.TotalBookings
	}

	return senderParams, nil
}
