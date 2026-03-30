package reminder

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"main/internal/domain/models"
	"main/internal/domain/notifier"
	"main/internal/domain/repositories"
)

type IReminderService interface {
	RemindBookings(ctx context.Context) error
}

type service struct {
	unitOfWork   repositories.IUnitOfWork
	classesRepo  repositories.IClasses
	bookingsRepo repositories.IBookings
	notifier     notifier.INotifier
	domainAddr   string
}

func New(
	unitOfWork repositories.IUnitOfWork,
	classesRepo repositories.IClasses,
	bookingsRepo repositories.IBookings,
	notifier notifier.INotifier,
	domainAddr string,
) *service {
	return &service{
		unitOfWork:   unitOfWork,
		classesRepo:  classesRepo,
		bookingsRepo: bookingsRepo,
		notifier:     notifier,
		domainAddr:   domainAddr,
	}
}

func (s *service) RemindBookings(ctx context.Context) error {
	classes, err := s.classesRepo.List(ctx)
	if err != nil {
		return fmt.Errorf("could not list classes: %w", err)
	}

	now := time.Now()

	var todayClasses []models.Class

	for _, c := range classes {
		if isTimeToRemind(c.StartTime, now) {
			todayClasses = append(todayClasses, c)
		}
	}

	if len(todayClasses) == 0 {
		return nil
	}

	// it is possible that will be more than one class the same day
	class := todayClasses[len(todayClasses)-1]

	bookings, err := s.bookingsRepo.ListByClassID(ctx, class.ID)
	if err != nil {
		return fmt.Errorf("could not list bookings for %v: %w", class.ID, err)
	}

	for _, booking := range bookings {
		if booking.RemindedAt != nil {
			continue
		}

		err := s.unitOfWork.WithTransaction(ctx, func(repos repositories.Repositories) error {
			update := map[string]any{"reminded_at": time.Now()}

			_, err = repos.Bookings.Update(ctx, booking.ID, update)
			if err != nil {
				return fmt.Errorf("could not update booking %v with %v: %w", booking.ID, update, err)
			}

			notifierParams := models.NotifierParams{
				RecipientEmail:     booking.Email,
				RecipientFirstName: booking.FirstName,
				RecipientLastName:  booking.LastName,
				ClassName:          class.ClassName,
				ClassLevel:         class.ClassLevel,
				StartTime:          class.StartTime,
				Location:           class.Location,
			}

			passOpt, err := repos.Passes.GetByEmail(ctx, booking.Email)
			if err != nil {
				return fmt.Errorf("could not get pass for %v: %w", booking.Email, err)
			}

			if passOpt.Exists() {
				pass := passOpt.Get()
				notifierParams.PassUsedBookingIDs = pass.UsedBookingIDs
				notifierParams.PassTotalBookings = &pass.TotalBookings
			}

			cancellationLink := fmt.Sprintf(
				"%s/bookings/%s/cancel_form?token=%s", s.domainAddr, booking.ID, booking.ConfirmationToken,
			)

			err = s.notifier.NotifyClassReminder(notifierParams, cancellationLink)
			if err != nil {
				return fmt.Errorf("could not remind about class with %v: %w", notifierParams, err)
			}

			slog.Info("class reminder", "email", booking.Email, "reminded_at", update["reminded_at"])

			return nil
		})
		if err != nil {
			return fmt.Errorf("remind class transaction failed: %w", err)
		}
	}

	return nil
}

func isTimeToRemind(classStartTime, now time.Time) bool {
	return classStartTime.Sub(now) < 10*time.Hour &&
		classStartTime.Day() == now.Day()
}
