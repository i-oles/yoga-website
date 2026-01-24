package passes

import (
	"context"
	"errors"
	"fmt"

	"main/internal/domain/models"
	"main/internal/domain/repositories"
	"main/internal/domain/sender"
)

type Service struct {
	passRepo      repositories.IPasses
	messageSender sender.ISender
}

func NewService(
	passRepo repositories.IPasses,
	messageSender sender.ISender,
) *Service {
	return &Service{
		passRepo:      passRepo,
		messageSender: messageSender,
	}
}

func (s Service) ActivatePass(ctx context.Context, params models.PassActivationParams) (models.Pass, error) {
	// TODO: refactor error domain and leave here proper domainError 400
	if params.UsedCredits > params.TotalCredits {
		return models.Pass{}, errors.New("bad request pass - implement me")
	}

	pass, err := s.passRepo.Upsert(ctx, params.Email, params.UsedCredits, params.TotalCredits)
	if err != nil {
		return models.Pass{}, fmt.Errorf("could not upsert pass data: %w", err)
	}

	return pass, nil
}
