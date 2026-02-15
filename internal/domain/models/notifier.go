package models

import (
	"time"

	"github.com/google/uuid"
)

type NotifierParams struct {
	RecipientEmail     string
	RecipientFirstName string
	RecipientLastName  string
	ClassName          string
	ClassLevel         string
	StartTime          time.Time
	Location           string
	PassUsedBookingIDs []uuid.UUID
	PassTotalBookings  *int
}

type BookingStatus string

const (
	StatusBooked    BookingStatus = "booked"
	StatusCancelled BookingStatus = "cancelled"
)
