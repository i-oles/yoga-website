package passes

import (
	"context"
	"errors"

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
	return models.Pass{}, errors.New("implement me")
}
