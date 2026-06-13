package bookings

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
	"main/pkg/optional"

	"github.com/google/uuid"
)

const (
	threeLastPasses = 3
)

type service struct {
	unitOfWork   repositories.IUnitOfWork
	bookingsRepo repositories.IBookings
	passManager  services.IPassManager
	notifier     notifier.INotifier
	domainAddr   string
}

func NewService(
	unitOfWork repositories.IUnitOfWork,
	bookingsRepo repositories.IBookings,
	passManager services.IPassManager,
	notifier notifier.INotifier,
	domainAddr string,
) *service {
	return &service{
		unitOfWork:   unitOfWork,
		bookingsRepo: bookingsRepo,
		passManager:  passManager,
		notifier:     notifier,
		domainAddr:   domainAddr,
	}
}

func (s *service) CreateBooking(ctx context.Context, token string) (models.Class, error) {
	var (
		pendingBooking models.PendingBooking
		class          models.Class
		bookingID      uuid.UUID
		passSlots      []models.PassSlot
	)

	err := s.unitOfWork.WithTransaction(ctx, func(repos repositories.Repositories) error {
		var err error

		pendingBooking, err = repos.PendingBookings.GetByConfirmationToken(ctx, token)
		if err != nil {
			if errors.Is(err, errs.ErrNotFound) {
				return viewErrors.ErrPendingBookingNotFound(
					fmt.Errorf("pending booking for token: %s not found", token),
				)
			}

			return fmt.Errorf("could not get pending booking: %w", err)
		}

		_, err = repos.Bookings.GetByEmailAndClassID(ctx, pendingBooking.ClassID, pendingBooking.Email)
		if err == nil {
			return viewErrors.ErrBookingAlreadyExists(
				pendingBooking.ClassID,
				pendingBooking.Email,
				fmt.Errorf("booking already exists for email %s and classID %s", pendingBooking.Email, pendingBooking.ClassID),
			)
		}

		if !errors.Is(err, errs.ErrNotFound) {
			return fmt.Errorf("could not get booking for email %s and classID %s: %w",
				pendingBooking.Email,
				pendingBooking.ClassID,
				err,
			)
		}

		class, err = repos.Classes.Get(ctx, pendingBooking.ClassID)
		if err != nil {
			return fmt.Errorf("could not get class with id: %s, %w", pendingBooking.ClassID, err)
		}

		err = s.checkClassAvailability(ctx, repos, class)
		if err != nil {
			return fmt.Errorf("class unavailable: %w", err)
		}

		// I need to make sure that I will check if previous pass will not have some empty slots. Three is enough.
		passes, err := repos.Passes.ListByEmail(ctx, pendingBooking.Email, threeLastPasses)
		if err != nil {
			return fmt.Errorf("could not get pass: %w", err)
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

		for _, pass := range passes {
			usedBookingsCount, err := s.bookingsRepo.CountForPassID(ctx, pass.ID)
			if err != nil {
				return fmt.Errorf("could not count bookings for passID %d: %w", pass.ID, err)
			}

			if usedBookingsCount < pass.TotalSlots {
				booking.PassID = optional.Of(pass.ID)
				booking.Pass = optional.Of(pass)

				bookingID, err = repos.Bookings.Insert(ctx, booking)
				if err != nil {
					return fmt.Errorf("could not insert booking: %w", err)
				}

				usedBookings, err := repos.Bookings.ListByPassID(ctx, pass.ID)
				if err != nil {
					return fmt.Errorf("could not list bookings by passID %d: %w", pass.ID, err)
				}

				passSlots = s.passManager.BuildPassSlots(usedBookings, pass.TotalSlots)

				return nil
			}
		}

		bookingID, err = repos.Bookings.Insert(ctx, booking)
		if err != nil {
			return fmt.Errorf("could not insert booking: %w", err)
		}

		return nil
	})
	if err != nil {
		return models.Class{}, fmt.Errorf("create booking transaction failed: %w", err)
	}

	err = s.sendConfirmation(
		pendingBooking, class, passSlots, token, bookingID,
	)
	if err != nil {
		return models.Class{},
			fmt.Errorf("could not send confirmation email %s: %w", pendingBooking.Email, err)
	}

	return class, nil
}

func (s *service) checkClassAvailability(
	ctx context.Context,
	repos repositories.Repositories,
	class models.Class,
) error {
	if class.StartTime.Before(time.Now()) {
		return viewErrors.ErrClassExpired(class.ID, fmt.Errorf("class %s has expired at %v", class.ID, class.StartTime))
	}

	bookingCount, err := repos.Bookings.CountForClassID(ctx, class.ID)
	if err != nil {
		return fmt.Errorf("could not count bookings for class %v: %w ", class.ID, err)
	}

	if bookingCount == class.MaxCapacity {
		return viewErrors.ErrSomeoneBookedClassFaster(fmt.Errorf("max capacity of class %d exceeded", class.MaxCapacity))
	}

	return nil
}

func (s *service) sendConfirmation(
	pendingBooking models.PendingBooking,
	class models.Class,
	passSlots []models.PassSlot,
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
		PassSlots:          passSlots,
	}

	cancellationLink := fmt.Sprintf(
		"%s/bookings/%s/cancel_form?token=%s", s.domainAddr, bookingID, token,
	)

	err := s.notifier.NotifyBookingConfirmation(notifierParams, cancellationLink)
	if err != nil {
		return fmt.Errorf("could not notify booking confirmation: %w", err)
	}

	return nil
}

func (s *service) CancelBooking(ctx context.Context, bookingID uuid.UUID, token string) error {
	var (
		booking   models.Booking
		passSlots []models.PassSlot
	)

	err := s.unitOfWork.WithTransaction(ctx, func(repos repositories.Repositories) error {
		var err error

		booking, err = s.ensureBookingCancellationAllowed(ctx, repos, bookingID, token)
		if err != nil {
			return fmt.Errorf("booking cancellation not allowed for bookingID %s: %w", bookingID, err)
		}

		err = repos.Bookings.Delete(ctx, bookingID)
		if err != nil {
			if errors.Is(err, errs.ErrNoRowsAffected) {
				return viewErrors.ErrBookingNotFound(
					fmt.Errorf("delete booking failure, booking with email %s for class %s not found", booking.Email, booking.ClassID),
				)
			}

			return fmt.Errorf("could not delete booking: %w", err)
		}

		if booking.Pass.Exists() {
			pass := booking.Pass.Get()

			usedBookings, err := repos.Bookings.ListByPassID(ctx, pass.ID)
			if err != nil {
				return fmt.Errorf("could not list bookings by pass id %d: %w", pass.ID, err)
			}

			passSlots = s.passManager.BuildPassSlots(usedBookings, pass.TotalSlots)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("cancel booking transaction failed: %w", err)
	}

	notifierParams := models.NotifierParams{
		RecipientFirstName: booking.FirstName,
		RecipientLastName:  booking.LastName,
		RecipientEmail:     booking.Email,
		ClassName:          booking.Class.ClassName,
		ClassLevel:         booking.Class.ClassLevel,
		StartTime:          booking.Class.StartTime,
		Location:           booking.Class.Location,
		PassSlots:          passSlots,
	}

	err = s.notifier.NotifyBookingCancellation(notifierParams)
	if err != nil {
		return fmt.Errorf("could not notify booking cancellation with %+v: %w", notifierParams, err)
	}

	return nil
}

func (s *service) ensureBookingCancellationAllowed(
	ctx context.Context, r repositories.Repositories, bookingID uuid.UUID, token string,
) (models.Booking, error) {
	booking, err := r.Bookings.GetByID(ctx, bookingID)
	if err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			return models.Booking{}, viewErrors.ErrBookingNotFound(
				fmt.Errorf("booking with id %s not found: %w", bookingID, err),
			)
		}

		return models.Booking{}, fmt.Errorf("could not get booking for id %s: %w", bookingID, err)
	}

	if booking.ConfirmationToken != token {
		return models.Booking{}, viewErrors.ErrInvalidCancellationLink(
			fmt.Errorf("cancel booking failed due to invalid token: %s for email: %s", booking.Email, token),
		)
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
	booking, err := s.bookingsRepo.GetByID(ctx, bookingID)
	if err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			return models.Booking{}, viewErrors.ErrBookingNotFound(
				fmt.Errorf("booking with id %s not found: %w", bookingID, err),
			)
		}

		return models.Booking{}, fmt.Errorf("could not get booking for id %s: %w", bookingID, err)
	}

	if booking.ConfirmationToken != token {
		return models.Booking{}, viewErrors.ErrInvalidCancellationLink(err)
	}

	return booking, nil
}

func (s *service) DeleteBooking(ctx context.Context, bookingID uuid.UUID) error {
	var (
		booking   models.Booking
		passSlots []models.PassSlot
	)

	err := s.unitOfWork.WithTransaction(ctx, func(repos repositories.Repositories) error {
		var err error

		booking, err = repos.Bookings.GetByID(ctx, bookingID)
		if err != nil {
			return fmt.Errorf("could get booking for id %s: %w", bookingID, err)
		}

		err = repos.Bookings.Delete(ctx, bookingID)
		if err != nil {
			return fmt.Errorf("could not delete booking for id %s: %w", bookingID, err)
		}

		if booking.Pass.Exists() {
			pass := booking.Pass.Get()
			usedBookings, err := repos.Bookings.ListByPassID(ctx, pass.ID)
			if err != nil {
				return fmt.Errorf("could not list bookings by pass id %d: %w", pass.ID, err)
			}

			passSlots = s.passManager.BuildPassSlots(usedBookings, pass.TotalSlots)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("delete booking transaction failed: %w", err)
	}

	notifierParams := models.NotifierParams{
		RecipientFirstName: booking.FirstName,
		RecipientLastName:  booking.LastName,
		RecipientEmail:     booking.Email,
		ClassName:          booking.Class.ClassName,
		ClassLevel:         booking.Class.ClassLevel,
		StartTime:          booking.Class.StartTime,
		Location:           booking.Class.Location,
		PassSlots:          passSlots,
	}

	err = s.notifier.NotifyBookingCancellation(notifierParams)
	if err != nil {
		return fmt.Errorf("could not nofify booking cancellation with %+v: %w", notifierParams, err)
	}

	return nil
}
