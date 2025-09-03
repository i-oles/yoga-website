package bookings

import (
	"main/internal/domain/models"
	"time"

	"github.com/google/uuid"
)

type SQLBooking struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	ClassID   uuid.UUID `gorm:"type:uuid;not null"`
	FirstName string    `gorm:"not null"`
	LastName  string    `gorm:"not null"`
	Email     string    `gorm:"not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

func (SQLBooking) TableName() string {
	return "confirmed_bookings"
}

func (s SQLBooking) ToDomain() models.Booking {
	return models.Booking{
		ID:        s.ID,
		ClassID:   s.ClassID,
		FirstName: s.FirstName,
		LastName:  s.LastName,
		Email:     s.Email,
		CreatedAt: s.CreatedAt,
	}
}

func FromDomain(domain models.Booking) SQLBooking {
	return SQLBooking{
		ID:        domain.ID,
		ClassID:   domain.ClassID,
		FirstName: domain.FirstName,
		LastName:  domain.LastName,
		Email:     domain.Email,
		CreatedAt: domain.CreatedAt,
	}
}
