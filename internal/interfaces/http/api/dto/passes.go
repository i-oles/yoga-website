package dto

import (
	"fmt"
	"time"

	"main/internal/domain/models"
	"main/pkg/converter"
)

type ActivatePassRequest struct {
	Email        string `binding:"required,min=3,max=40" json:"email"`
	UsedCredits  int    `binding:"required,gte=0" json:"used_credits"`
	TotalCredits int    `binding:"required,gte=1" json:"total_credits"`
}

type PassDTO struct {
	ID           int       `json:"id"`
	Email        string    `json:"email"`
	UsedCredits  int       `json:"used_credits"`
	TotalCredits int       `json:"total_credits"`
	CreatedAt    time.Time `json:"week_day"`
}

func ToPassDTO(pass models.Pass) (PassDTO, error) {
	cratedAtWarsawTime, err := converter.ConvertToWarsawTime(pass.CreatedAt)
	if err != nil {
		return PassDTO{}, fmt.Errorf("error while converting createdAt to warsaw time: %w", err)
	}

	return PassDTO{
		ID:           pass.ID,
		Email:        pass.Email,
		UsedCredits:  pass.UsedCredits,
		TotalCredits: pass.TotalCredits,
		CreatedAt:    cratedAtWarsawTime,
	}, nil
}
