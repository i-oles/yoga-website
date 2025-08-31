package confirmedbookings

import (
	"main/internal/domain/models"
	"strings"
	"time"

	"github.com/google/uuid"
)

type SQLConfirmedBooking struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	ClassID   uuid.UUID `gorm:"type:uuid;not null"`
	FirstName string    `gorm:"not null"`
	LastName  string    `gorm:"not null"`
	Email     string    `gorm:"not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

func (SQLConfirmedBooking) TableName() string {
	return "confirmed_bookings"
}

func (s SQLConfirmedBooking) ToDomain() models.ConfirmedBooking {
	return models.ConfirmedBooking{
		ID:        s.ID,
		ClassID:   s.ClassID,
		FirstName: s.FirstName,
		LastName:  s.LastName,
		Email:     s.Email,
		CreatedAt: s.CreatedAt,
	}
}

func FromDomain(domain models.ConfirmedBooking) SQLConfirmedBooking {
	return SQLConfirmedBooking{
		ID:        domain.ID,
		ClassID:   domain.ClassID,
		FirstName: domain.FirstName,
		LastName:  domain.LastName,
		Email:     strings.ToLower(domain.Email),
		CreatedAt: domain.CreatedAt,
	}
}
