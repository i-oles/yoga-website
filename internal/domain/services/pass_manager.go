package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	sharedErrors "main/internal/domain/errs"
	"main/internal/domain/models"
	"main/pkg/tools"

	"github.com/google/uuid"
)

type PassManager struct{}

func (p *PassManager) BuildPassItems(
	ctx context.Context,
	bookings []models.Booking,
	totalBookings int,
) ([]models.PassItem, error) {
	passItems := make([]models.PassItem, 0, totalBookings)

	for _, booking := range bookings {
		classStartTime := booking.Class.StartTime
		passItem := models.PassItem{
			ClassStartTime: &classStartTime,
		}

		if classStartTime.Before(time.Now()) {
			passItem.Status = models.FuturePassStatus
		} else {
			passItem.Status = models.UsedPassStatus
		}

		passItems = append(passItems, passItem)
	}

	if len(passItems) < totalBookings {
		for i := len(passItems); i < totalBookings; i++ {
			passItems = append(passItems, models.PassItem{
				Status: models.BlankPassStatus,
			})
		}
	}

	return passItems, nil
}

func (p *PassManager) TryIncrementPass(
	ctx context.Context,
	pass models.Pass,
	bookingID uuid.UUID,
) (models.Pass, error) {
	if len(pass.UsedBookingIDs)+1 <= pass.TotalBookings {
		updatedBookingIDs := pass.UsedBookingIDs
		updatedBookingIDs = append(updatedBookingIDs, bookingID)

		return models.Pass{
			ID:             pass.ID,
			Email:          pass.Email,
			UsedBookingIDs: updatedBookingIDs,
			TotalBookings:  pass.TotalBookings,
		}, nil
	}

	return pass, nil
}

func (p *PassManager) TryDecrementPass(
	ctx context.Context,
	pass models.Pass,
	bookingID uuid.UUID,
) (models.Pass, error) {
	updatedBookingIDs, err := tools.RemoveFromSlice(pass.UsedBookingIDs, bookingID)
	if errors.Is(err, sharedErrors.ErrBookingIDNotFoundInPass) {
		return models.Pass{}, nil
	}

	if err != nil {
		return models.Pass{},
			fmt.Errorf("could not remove bookingID %v from usedBookingIDs", bookingID)
	}

	return models.Pass{
		ID:             pass.ID,
		Email:          pass.Email,
		UsedBookingIDs: updatedBookingIDs,
		TotalBookings:  pass.TotalBookings,
	}, nil
}
