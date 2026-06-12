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
	// TODO: verify czy poprzedni karnet jest zapelniony - jesli nie to blad dla usera.

	if params.BookingsCountForPass > params.TotalBookingsCount {
		return models.Pass{},
			api.ErrValidation(
				fmt.Errorf("usedBookings: %d is grater than totalBookings: %d",
					params.BookingsCountForPass,
					params.TotalBookingsCount),
			)
	}

	pass, err := s.passesRepo.Insert(
		ctx,
		params.Email,
		params.TotalBookingsCount,
	)
	if err != nil {
		return models.Pass{}, fmt.Errorf("could not insert pass for %s: %w", params.Email, err)
	}

	// when user booked one or more classes in future - system needs to add this bookings to Pass
	// TODO: zmien nazwe usesBookings w tym miejscu bo chodzi tu o przyszle albo niewbite w karnet bookingi
	// TODO: rozwiaz problem posiadania dwoch karnetow na raz - jak pobieramy karnet to jest on po created at a moze powinien byc pobierany pierwszy kotry ma najmniej wolnych slotow?
	bookingsForPass, err := s.getBookingsForPass(ctx, params.Email, params.BookingsCountForPass)
	if err != nil {
		return models.Pass{},
			fmt.Errorf("could not get usedBookingIDs for email %s: %w", params.Email, err)
	}

	for _, booking := range bookingsForPass {
		_, err = s.bookingsRepo.Update(ctx, booking.ID, map[string]any{
			"pass_id": pass.ID,
		})
		if err != nil {
			return models.Pass{}, fmt.Errorf("could not update booking %s with pass_id %d: %w", booking.ID, pass.ID, err)
		}
	}

	passItems := s.passManager.BuildPassItems(bookingsForPass, params.TotalBookingsCount)

	err = s.notifier.NotifyPassActivation(params.Email, passItems)
	if err != nil {
		return models.Pass{}, fmt.Errorf("could notify pass activation with %v: %w", pass, err)
	}

	return pass, nil
}

func (s *service) getBookingsForPass(
	ctx context.Context,
	email string,
	bookingsCountForPass int,
) ([]models.Booking, error) {
	if bookingsCountForPass == 0 {
		return nil, nil
	}

	bookings, err := s.bookingsRepo.ListByEmail(ctx, email, bookingsCountForPass)
	if err != nil {
		return nil, fmt.Errorf("could not list usedBookings for email %s: %w", email, err)
	}

	return bookings, nil
}
