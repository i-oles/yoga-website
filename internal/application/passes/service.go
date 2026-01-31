package passes

import (
	"context"
	"fmt"

	"main/internal/domain/errs/api"
	"main/internal/domain/models"
	"main/internal/domain/repositories"
	"main/internal/domain/sender"

	"github.com/google/uuid"
)

type Service struct {
	passesRepo    repositories.IPasses
	bookingsRepo  repositories.IBookings
	messageSender sender.ISender
}

func NewService(
	passesRepo repositories.IPasses,
	bookingsRepo repositories.IBookings,
	messageSender sender.ISender,
) *Service {
	return &Service{
		passesRepo:    passesRepo,
		bookingsRepo:  bookingsRepo,
		messageSender: messageSender,
	}
}

func (s Service) ActivatePass(ctx context.Context, params models.PassActivationParams) (models.Pass, error) {
	if params.UsedBookings > params.TotalBookings {
		return models.Pass{}, api.ErrValidation(fmt.Errorf("usedBookings: %d can not be grater than totalBookings: %d", params.UsedBookings, params.TotalBookings))
	}

	passOpt, err := s.passesRepo.GetByEmail(ctx, params.Email)
	if err != nil {
		return models.Pass{}, fmt.Errorf("could not get pass by email %s: %w", params.Email, err)
	}

	// when user booked one or more classes in future - system needs to add this bookings to Pass
	usedBookingIDs, err := s.getUsedBookingIDsForPass(ctx, params.Email, params.UsedBookings)
	if err != nil {
		return models.Pass{}, fmt.Errorf("could not get usedBookingIDs for email %s: %w", params.Email, err)
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

		err = s.messageSender.SendPass(pass)
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

	err = s.messageSender.SendPass(newPass)
	if err != nil {
		return models.Pass{}, fmt.Errorf("could not send pass: %w", err)
	}

	return newPass, nil
}

func (s Service) getUsedBookingIDsForPass(
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
