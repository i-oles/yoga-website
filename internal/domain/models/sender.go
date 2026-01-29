package models

import (
	"time"

	"github.com/google/uuid"
)

type SenderParams struct {
	RecipientEmail     string
	RecipientFirstName string
	RecipientLastName  *string
	ClassName          string
	ClassLevel         string
	StartTime          time.Time
	Location           string
	PassUsedBookingIDs []uuid.UUID
	PassTotalBookings  *int
}
