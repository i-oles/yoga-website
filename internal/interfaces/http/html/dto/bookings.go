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
type ClassView struct {
	WeekDay    string
	StartDate  string
	StartHour  string
	ClassLevel string
	ClassName  string
	Location   string
}

func ToClassView(class models.Class) (ClassView, error) {
	warsawStartTime, err := converter.ConvertToWarsawTime(class.StartTime)
	if err != nil {
		return ClassView{}, fmt.Errorf("could not convert class start time from booking: %w", err)
	}

	weekDay, err := translator.TranslateToWeekDayToPolish(warsawStartTime.Weekday())
	if err != nil {
		return ClassView{}, fmt.Errorf("could not convert weekday from booking: %w", err)
	}

	return ClassView{
		WeekDay:    weekDay,
		StartDate:  warsawStartTime.Format(converter.DateLayout),
		StartHour:  warsawStartTime.Format(converter.HourLayout),
		ClassLevel: class.ClassLevel,
		ClassName:  class.ClassName,
		Location:   class.Location,
	}, nil
}

type BookingCancelView struct {
	WeekDay           string
	StartDate         string
	StartHour         string
	ClassLevel        string
	ClassName         string
	Location          string
	BookingID         uuid.UUID
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
		WeekDay:           weekDay,
		StartDate:         warsawStartTime.Format(converter.DateLayout),
		StartHour:         warsawStartTime.Format(converter.HourLayout),
		ClassLevel:        booking.Class.ClassLevel,
		ClassName:         booking.Class.ClassName,
		Location:          booking.Class.Location,
		BookingID:         booking.ID,
		ConfirmationToken: booking.ConfirmationToken,
	}, nil
}
