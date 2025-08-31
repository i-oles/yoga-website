package pendingoperations

import (
	"main/internal/domain/models"
	"strings"
	"time"

	"github.com/google/uuid"
)

type SQLPendingOperation struct {
	ID                uuid.UUID `gorm:"type:uuid;primaryKey"`
	ClassID           uuid.UUID `gorm:"type:uuid;not null"`
	Operation         string    `gorm:"type:text;not null;check:operation IN ('create_booking','cancel_booking')"`
	Email             string    `gorm:"not null"`
	FirstName         string    `gorm:"not null"`
	LastName          *string
	ConfirmationToken string    `gorm:"unique;not null"`
	CreatedAt         time.Time `gorm:"autoCreateTime"`
}

func (SQLPendingOperation) TableName() string {
	return "pending_operations"
}

func (s SQLPendingOperation) ToDomain() models.PendingOperation {
	return models.PendingOperation{
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

func FromDomain(d models.PendingOperation) SQLPendingOperation {
	return SQLPendingOperation{
		ID:                d.ID,
		ClassID:           d.ClassID,
		Operation:         string(d.Operation),
		Email:             strings.ToLower(d.Email),
		FirstName:         d.FirstName,
		LastName:          d.LastName,
		ConfirmationToken: d.ConfirmationToken,
		CreatedAt:         d.CreatedAt,
	}
}
