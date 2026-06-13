package dto

import (
	"fmt"
	"time"

	"main/internal/domain/models"
	"main/pkg/converter"
)

type ActivatePassRequest struct {
	Email      string `binding:"required,min=3,max=40" json:"email"`
	UsedSlots  int    `binding:"min=0" json:"used_slots"`
	TotalSlots int    `binding:"min=1" json:"total_slots"`
}

type PassDTO struct {
	ID         int       `json:"id"`
	Email      string    `json:"email"`
	TotalSlots int       `json:"total_bookings"`
	UpdatedAt  time.Time `json:"updated_at"`
	CreatedAt  time.Time `json:"created_at"`
}

func ToPassDTO(pass models.Pass) (PassDTO, error) {
	cratedAtWarsawTime, err := converter.ConvertToWarsawTime(pass.CreatedAt)
	if err != nil {
		return PassDTO{}, fmt.Errorf("error while converting createdAt to warsaw time: %w", err)
	}

	updatedAtWarsawTime, err := converter.ConvertToWarsawTime(pass.UpdatedAt)
	if err != nil {
		return PassDTO{}, fmt.Errorf("error while converting createdAt to warsaw time: %w", err)
	}

	return PassDTO{
		ID:         pass.ID,
		Email:      pass.Email,
		TotalSlots: pass.TotalSlots,
		UpdatedAt:  updatedAtWarsawTime,
		CreatedAt:  cratedAtWarsawTime,
	}, nil
}
