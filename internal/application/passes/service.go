package passes

import (
	"context"
	"fmt"

	"main/internal/domain/errs/api"
	"main/internal/domain/models"
	"main/internal/domain/notifier"
	"main/internal/domain/repositories"
	"main/internal/domain/services"

	"github.com/google/uuid"
)

type service struct {
	passesRepo   repositories.IPasses
	bookingsRepo repositories.IBookings
	notifier     notifier.INotifier
	passManager  services.IPassManager
}

func NewService(
	passesRepo repositories.IPasses,
	bookingsRepo repositories.IBookings,
	notifier notifier.INotifier,
	passManager services.IPassManager,
) *service {
	return &service{
		passesRepo:   passesRepo,
		bookingsRepo: bookingsRepo,
		notifier:     notifier,
		passManager:  passManager,
	}
}

func (s *service) ActivatePass(
	ctx context.Context, params models.PassActivationParams,
) (models.Pass, error) {
	if params.UsedBookingsCount > params.TotalBookingsCount {
		return models.Pass{},
			api.ErrValidation(
				fmt.Errorf("usedBookings: %d is grater than totalBookings: %d",
					params.UsedBookingsCount,
					params.TotalBookingsCount),
			)
	}

	passOpt, err := s.passesRepo.GetByEmail(ctx, params.Email)
	if err != nil {
		return models.Pass{}, fmt.Errorf("could not get pass by email %s: %w", params.Email, err)
	}

	// when user booked one or more classes in future - system needs to add this bookings to Pass
	usedBookings, err := s.getUsedBookingsForPass(ctx, params.Email, params.UsedBookingsCount)
	if err != nil {
		return models.Pass{},
			fmt.Errorf("could not get usedBookingIDs for email %s: %w", params.Email, err)
	}

	usedBookingIDs := make([]uuid.UUID, 0, len(usedBookings))
	for _, booking := range usedBookings {
		usedBookingIDs = append(usedBookingIDs, booking.ID)
	}

	passItems, err := s.passManager.BuildPassItems(ctx, usedBookings, params.TotalBookingsCount)
	if err != nil {
		return models.Pass{},
			fmt.Errorf("could not build passItems %s: %w", params.Email, err)
	}

	if !passOpt.Exists() {
		pass, err := s.passesRepo.Insert(
			ctx,
			params.Email,
			usedBookingIDs,
			params.TotalBookingsCount,
		)
		if err != nil {
			return models.Pass{}, fmt.Errorf("could not insert pass for %s: %w", params.Email, err)
		}

		err = s.notifier.NotifyPassActivation(params.Email, passItems)
		if err != nil {
			return models.Pass{}, fmt.Errorf("could not send pass %v: %w", pass, err)
		}

		return pass, nil
	}

	pass := passOpt.Get()

	updatedPass, err := s.passesRepo.Update(ctx, pass.ID, usedBookingIDs, params.TotalBookingsCount)
	if err != nil {
		return models.Pass{}, fmt.Errorf("could not update pass with %+v: %w", usedBookingIDs, err)
	}

	err = s.notifier.NotifyPassActivation(params.Email, passItems)
	if err != nil {
		return models.Pass{}, fmt.Errorf("could notify pass activation with %v: %w", pass, err)
	}

	return updatedPass, nil
}

func (s *service) getUsedBookingsForPass(
	ctx context.Context,
	email string,
	passUsedBookings int,
) ([]models.Booking, error) {
	if passUsedBookings == 0 {
		return nil, nil
	}

	usedBookings, err := s.bookingsRepo.ListByEmail(ctx, email, passUsedBookings)
	if err != nil {
		return nil, fmt.Errorf("could not list usedBookings for email %s: %w", email, err)
	}

	return usedBookings, nil
}
