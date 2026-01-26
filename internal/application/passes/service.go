package passes

import (
	"context"
	"errors"
	"fmt"

	"main/internal/domain/models"
	"main/internal/domain/repositories"
	"main/internal/domain/sender"
	"main/internal/infrastructure/errs"
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

	pass, err := s.passRepo.GetByEmail(ctx, params.Email)
	if errors.Is(err, errs.ErrNotFound) {
		pass, err := s.passRepo.Insert(
			ctx,
			params.Email,
			params.UsedCredits,
			params.TotalCredits,
		)
		if err != nil {
			return models.Pass{}, fmt.Errorf("could not insert pass: %w", err)
		}

		err = s.messageSender.SendPass(pass)
		if err != nil {
			return models.Pass{}, fmt.Errorf("could not send pass: %w", err)
		}

		return pass, nil
	}

	if err != nil {
		return models.Pass{}, fmt.Errorf("could not get pass by email %s: %w", params.Email, err)
	}

	update := map[string]any{
		"used_credits":  params.UsedCredits,
		"total_credits": params.TotalCredits,
	}

	err = s.passRepo.Update(ctx, pass.ID, update)
	if err != nil {
		return models.Pass{}, fmt.Errorf("could not update pass with %+v: %w", update, err)
	}

	newPass := models.Pass{
		ID:           pass.ID,
		Email:        pass.Email,
		UsedCredits:  params.UsedCredits,
		TotalCredits: params.TotalCredits,
		CreatedAt:    pass.CreatedAt,
	}

	err = s.messageSender.SendPass(newPass)
	if err != nil {
		return models.Pass{}, fmt.Errorf("could not send pass: %w", err)
	}

	return newPass, nil
}
