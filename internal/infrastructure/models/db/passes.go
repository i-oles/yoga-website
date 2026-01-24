package db

import (
	"time"

	"main/internal/domain/models"
)

type SQLPass struct {
	ID           int       `gorm:"primaryKey"`
	Email        string    `gorm:"unique;not null"`
	UsedCredits  int       `gorm:"not null"`
	TotalCredits int       `gorm:"not null"`
	CreatedAt    time.Time `gorm:"autoCreateTime"`
}

func (SQLPass) TableName() string {
	return "passes"
}

func (s SQLPass) ToDomain() models.Pass {
	pass := models.Pass{
		ID:           s.ID,
		Email:        s.Email,
		UsedCredits:  s.UsedCredits,
		TotalCredits: s.TotalCredits,
		CreatedAt:    s.CreatedAt,
	}

	return pass
}
