package db

import (
	"time"

	"main/internal/domain/models"

	"github.com/google/uuid"
)

type SQLPass struct {
	ID             int         `gorm:"primaryKey"`
	Email          string      `gorm:"unique;not null"`
	UsedBookingIDs []uuid.UUID `gorm:"type:json;serializer:json"`
	TotalBookings  int         `gorm:"not null"`
	UpdatedAt      time.Time   `gorm:"autoUpdateTime"`
	CreatedAt      time.Time   `gorm:"autoCreateTime"`
}

func (SQLPass) TableName() string {
	return "passes"
}

func (s SQLPass) ToDomain() models.Pass {
	pass := models.Pass{
		ID:             s.ID,
		Email:          s.Email,
		UsedBookingIDs: s.UsedBookingIDs,
		TotalBookings:  s.TotalBookings,
		UpdatedAt:      s.UpdatedAt,
		CreatedAt:      s.CreatedAt,
	}

	return pass
}
