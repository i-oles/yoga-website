package dto

import (
	"fmt"
	"main/internal/domain/models"
	"main/pkg/converter"
	"main/pkg/translator"

	"github.com/google/uuid"
)

type BookingCancelForm struct {
	Token string `form:"token" binding:"required,len=44"`
}

type BookingCancelURI struct {
	BookingID string `uri:"id" binding:"required"`
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

type BookingCancelView struct {
	BookingID         uuid.UUID
	WeekDay           string
	StartDate         string
	StartHour         string
	ClassLevel        string
	ClassName         string
	MaxCapacity       int
	Location          string
	ConfirmationToken string
}

func ToBookingCancelView(booking models.Booking) (BookingCancelView, error) {
	warsawStartTime, err := converter.ConvertToWarsawTime(booking.Class.StartTime)
	if err != nil {
		return BookingCancelView{}, fmt.Errorf("could not convert class start time from booking: %w", err)
	}

	weekDay, err := translator.TranslateToWeekDayToPolish(warsawStartTime.Weekday())
	if err != nil {
		return BookingCancelView{}, fmt.Errorf("could not convert weekday from booking: %w", err)
	}

	return BookingCancelView{
		BookingID:         booking.ID,
		WeekDay:           weekDay,
		StartDate:         warsawStartTime.Format(converter.DateLayout),
		StartHour:         warsawStartTime.Format(converter.HourLayout),
		ClassLevel:        booking.Class.ClassLevel,
		ClassName:         booking.Class.ClassName,
		MaxCapacity:       booking.Class.MaxCapacity,
		Location:          booking.Class.Location,
		ConfirmationToken: booking.ConfirmationToken,
	}, nil
}
