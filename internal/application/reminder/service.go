package reminder

import (
	"context"
	"fmt"
	"time"

	"main/internal/domain/models"
	"main/internal/domain/notifier"
	"main/internal/domain/repositories"

	"github.com/google/uuid"
)

type service struct {
	classesRepo  repositories.IClasses
	bookingsRepo repositories.IBookings
	passesRepo   repositories.IPasses
	notifier     notifier.INotifier
	domainAddr   string
}

func NewReminderService(
	classesRepo repositories.IClasses,
	bookingsRepo repositories.IBookings,
	passesRepo repositories.IPasses,
	domainAddr string,
	notifier notifier.INotifier,
) *service {
	return &service{
		classesRepo:  classesRepo,
		bookingsRepo: bookingsRepo,
		passesRepo:   passesRepo,
		notifier:     notifier,
		domainAddr:   domainAddr,
	}
}

func (s *service) RemindClass(ctx context.Context) error {
	classes, err := s.classesRepo.List(ctx)
	if err != nil {
		return err
	}

	now := time.Now()

	var class models.Class

	for _, c := range classes {
		if isClassToday(now, c) {
			class = c

			break
		}
	}

	if class.ID == uuid.Nil {
		return nil
	}

	bookings, err := s.bookingsRepo.ListByClassID(ctx, class.ID)
	if err != nil {
		return err
	}

	// TODO: handler errors

	// TODO: add transaction
	for _, booking := range bookings {
		notifierParams := models.NotifierParams{
			RecipientEmail:     booking.Email,
			RecipientFirstName: booking.FirstName,
			RecipientLastName:  booking.LastName,
			ClassName:          class.ClassName,
			ClassLevel:         class.ClassLevel,
			StartTime:          class.StartTime,
			Location:           class.Location,
		}

		passOpt, err := s.passesRepo.GetByEmail(ctx, booking.Email)
		if err != nil {
			return err
		}

		if passOpt.Exists() {
			pass := passOpt.Get()
			notifierParams.PassUsedBookingIDs = pass.UsedBookingIDs
			notifierParams.PassTotalBookings = &pass.TotalBookings
		}

		cancellationLink := fmt.Sprintf(
			"%s/bookings/%s/cancel_form?token=%s", s.domainAddr, booking.ID, booking.ConfirmationToken,
		)

		err = s.notifier.NotifyReminder(notifierParams, cancellationLink)
		if err != nil {
			return err
		}

		// TODO: save new booking field --> reminded_at (timestamp)
	}

	return nil
}

func isClassToday(now time.Time, class models.Class) bool {
	return class.StartTime.Day() == now.Day() &&
		class.StartTime.Month() == now.Month() &&
		class.StartTime.Year() == now.Year()
}
