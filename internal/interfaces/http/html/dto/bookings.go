package dto

import (
	"fmt"
	"main/internal/domain/models"
	"main/pkg/converter"
)

type BookingCancelForm struct {
	Token     string `form:"token" binding:"required,len=44"`
	BookingID string `uri:"booking_id" binding:"required"`
}

type BookingCreateForm struct {
	Token string `form:"token" binding:"required,len=44"`
}
type BookingView struct {
	ClassName string
	Date      string
	Hour      string
	Location  string
}

func ToBookingView(class models.Class) (BookingView, error) {
	warsawTimeDate, err := converter.ConvertToWarsawTime(class.StartTime)
	if err != nil {
		return BookingView{}, fmt.Errorf("could not convert start time: %w", err)
	}

	return BookingView{
		ClassName: class.ClassName,
		Date:      warsawTimeDate.Format(converter.DateLayout),
		Hour:      warsawTimeDate.Format(converter.HourLayout),
		Location:  class.Location,
	}, nil
}
