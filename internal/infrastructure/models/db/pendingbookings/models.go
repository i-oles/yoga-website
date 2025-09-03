package pendingbookings

import (
	"main/internal/domain/models"
	"time"

	"github.com/google/uuid"
)

type SQLPendingBooking struct {
	ID                uuid.UUID `gorm:"type:uuid;primaryKey"`
	ClassID           uuid.UUID `gorm:"type:uuid;not null"`
	Operation         string    `gorm:"type:text;not null;check:operation IN ('create_booking','cancel_booking')"`
	Email             string    `gorm:"not null"`
	FirstName         string    `gorm:"not null"`
	LastName          *string
	ConfirmationToken string    `gorm:"unique;not null"`
	CreatedAt         time.Time `gorm:"autoCreateTime"`
}

func (SQLPendingBooking) TableName() string {
	return "pending_operations"
}

func (s SQLPendingBooking) ToDomain() models.PendingBooking {
	return models.PendingBooking{
		ID:                s.ID,
		ClassID:           s.ClassID,
		Operation:         models.Operation(s.Operation),
		Email:             s.Email,
		FirstName:         s.FirstName,
		LastName:          s.LastName,
		ConfirmationToken: s.ConfirmationToken,
		CreatedAt:         s.CreatedAt,
	}
}

func FromDomain(d models.PendingBooking) SQLPendingBooking {
	return SQLPendingBooking{
		ID:                d.ID,
		ClassID:           d.ClassID,
		Operation:         string(d.Operation),
		Email:             d.Email,
		FirstName:         d.FirstName,
		LastName:          d.LastName,
		ConfirmationToken: d.ConfirmationToken,
		CreatedAt:         d.CreatedAt,
	}
}
