package db

import (
	"main/internal/domain/models"
)

type SQLContact struct {
	ID        int    `gorm:"primaryKey"`
	Email     string `gorm:"not null"`
	FirstName string `gorm:"not null"`
	LastName  string `gorm:"not null"`
}

func (SQLContact) TableName() string {
	return "contacts"
}

func (s SQLContact) ToDomain() models.Contact {
	return models.Contact{
		ID:        s.ID,
		Email:     s.Email,
		FirstName: s.FirstName,
		LastName:  s.LastName,
	}
}

func SQLContactFromDomain(domain models.Contact) SQLContact {
	return SQLContact{
		ID:        domain.ID,
		Email:     domain.Email,
		FirstName: domain.FirstName,
		LastName:  domain.LastName,
	}
}
