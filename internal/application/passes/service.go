package passes

import (
	"context"
	"fmt"

	"main/internal/domain/errs/api"
	"main/internal/domain/models"
	"main/internal/domain/notifier"
	"main/internal/domain/repositories"
	"main/internal/domain/services"
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
	actualPass, err := s.passesRepo.GetByEmail(ctx, params.Email)
	if err != nil {
		return models.Pass{}, fmt.Errorf("could not insert pass for %s: %w", params.Email, err)
	}

	if actualPass.Exists() {
		pass := actualPass.Get()

		bookingsCount, err := s.bookingsRepo.CountForPassID(ctx, pass.ID)
		if err != nil {
			return models.Pass{}, fmt.Errorf("could not count bookings for %d: %w", pass.ID, err)
		}

		if bookingsCount < pass.TotalSlots {
			return models.Pass{}, api.ErrPreviousPassNotFinished(
				fmt.Errorf("previous pass with passID: %d for %s still has empty slots, use it first", pass.ID, params.Email),
			)
		}
	}

	if params.UsedSlots > params.TotalSlots {
		return models.Pass{},
			api.ErrValidation(
				fmt.Errorf("usedSlots: %d is grater than totalSlots: %d",
					params.UsedSlots,
					params.TotalSlots),
			)
	}

	pass, err := s.passesRepo.Insert(
		ctx,
		params.Email,
		params.TotalSlots,
	)
	if err != nil {
		return models.Pass{}, fmt.Errorf("could not insert pass for %s: %w", params.Email, err)
	}

	// when user booked one or more classes in future - system needs to add this bookings to Pass
	bookingsForUsedSlots, err := s.getBookingsForUsedSlots(ctx, params.Email, params.UsedSlots)
	if err != nil {
		return models.Pass{},
			fmt.Errorf("could not get bookings for pass for email %s: %w", params.Email, err)
	}

	for _, booking := range bookingsForUsedSlots {
		_, err = s.bookingsRepo.Update(ctx, booking.ID, map[string]any{
			"pass_id": pass.ID,
		})
		if err != nil {
			return models.Pass{},
				fmt.Errorf("could not update booking %s with pass_id %d: %w", booking.ID, pass.ID, err)
		}
	}

	passSlots := s.passManager.BuildPassSlots(bookingsForUsedSlots, params.TotalSlots)

	err = s.notifier.NotifyPassActivation(params.Email, passSlots)
	if err != nil {
		return models.Pass{}, fmt.Errorf("could notify pass activation with %v: %w", pass, err)
	}

	return pass, nil
}

func (s *service) getBookingsForUsedSlots(
	ctx context.Context,
	email string,
	usedSlots int,
) ([]models.Booking, error) {
	if usedSlots == 0 {
		return nil, nil
	}

	bookings, err := s.bookingsRepo.ListByEmail(ctx, email, usedSlots)
	if err != nil {
		return nil, fmt.Errorf("could not list bookings for email %s: %w", email, err)
	}

	return bookings, nil
}
