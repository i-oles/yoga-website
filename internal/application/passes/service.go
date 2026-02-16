package passes

import (
	"context"
	"fmt"

	"main/internal/domain/errs/api"
	"main/internal/domain/models"
	"main/internal/domain/notifier"
	"main/internal/domain/repositories"

	"github.com/google/uuid"
)

type service struct {
	passesRepo   repositories.IPasses
	bookingsRepo repositories.IBookings
	notifier     notifier.INotifier
}

func NewService(
	passesRepo repositories.IPasses,
	bookingsRepo repositories.IBookings,
	notifier notifier.INotifier,
) *service {
	return &service{
		passesRepo:   passesRepo,
		bookingsRepo: bookingsRepo,
		notifier:     notifier,
	}
}

func (s *service) ActivatePass(
	ctx context.Context, params models.PassActivationParams,
) (models.Pass, error) {
	if params.UsedBookings > params.TotalBookings {
		return models.Pass{},
			api.ErrValidation(fmt.Errorf("usedBookings: %d is grater than totalBookings: %d", params.UsedBookings, params.TotalBookings))
	}

	passOpt, err := s.passesRepo.GetByEmail(ctx, params.Email)
	if err != nil {
		return models.Pass{}, fmt.Errorf("could not get pass by email %s: %w", params.Email, err)
	}

	// when user booked one or more classes in future - system needs to add this bookings to Pass
	usedBookingIDs, err := s.getUsedBookingIDsForPass(ctx, params.Email, params.UsedBookings)
	if err != nil {
		return models.Pass{},
			fmt.Errorf("could not get usedBookingIDs for email %s: %w", params.Email, err)
	}

	if !passOpt.Exists() {
		pass, err := s.passesRepo.Insert(
			ctx,
			params.Email,
			usedBookingIDs,
			params.TotalBookings,
		)
		if err != nil {
			return models.Pass{}, fmt.Errorf("could not insert pass for %s: %w", params.Email, err)
		}

		err = s.notifier.NotifyPassActivation(pass)
		if err != nil {
			return models.Pass{}, fmt.Errorf("could not send pass %v: %w", pass, err)
		}

		return pass, nil
	}

	pass := passOpt.Get()

	err = s.passesRepo.Update(ctx, pass.ID, usedBookingIDs, params.TotalBookings)
	if err != nil {
		return models.Pass{}, fmt.Errorf("could not update pass with %+v: %w", usedBookingIDs, err)
	}

	newPass := models.Pass{
		ID:             pass.ID,
		Email:          pass.Email,
		UsedBookingIDs: usedBookingIDs,
		TotalBookings:  params.TotalBookings,
		CreatedAt:      pass.CreatedAt,
		UpdatedAt:      pass.UpdatedAt,
	}

	err = s.notifier.NotifyPassActivation(newPass)
	if err != nil {
		return models.Pass{}, fmt.Errorf("could notify pass activation with %v: %w", pass, err)
	}

	return newPass, nil
}

func (s *service) getUsedBookingIDsForPass(
	ctx context.Context,
	email string,
	passUsedBookings int,
) ([]uuid.UUID, error) {
	if passUsedBookings == 0 {
		return []uuid.UUID{}, nil
	}

	usedBookingIDs, err := s.bookingsRepo.GetIDsByEmail(ctx, email, passUsedBookings)
	if err != nil {
		return nil, fmt.Errorf("could not get usedBookingIDs for email %s: %w", email, err)
	}

	return usedBookingIDs, nil
}
