package db

import (
	"time"

	"main/internal/domain/models"
)

type SQLPass struct {
	ID           int       `gorm:"primaryKey"`
	Email        string    `gorm:"not null"`
	Credits      int       `gorm:"not null"`
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
		Credits:      s.Credits,
		TotalCredits: s.TotalCredits,
		CreatedAt:    s.CreatedAt,
	}

	return pass
}

func SQLPassFromDomain(domain models.Pass) SQLPass {
	return SQLPass{
		ID:           domain.ID,
		Email:        domain.Email,
		Credits:      domain.Credits,
		TotalCredits: domain.TotalCredits,
		CreatedAt:    domain.CreatedAt,
	}
}
