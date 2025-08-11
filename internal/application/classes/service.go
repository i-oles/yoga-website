package classes

import (
	"context"
	"fmt"
	"main/internal/domain/models"
	"main/internal/domain/repositories"
)

type Service struct {
	classesRepo repositories.Classes
}

func New(classesRepo repositories.Classes) *Service {
	return &Service{classesRepo: classesRepo}
}

func (s *Service) GetAllClasses(ctx context.Context) ([]models.Class, error) {
	classes, err := s.classesRepo.GetAllClasses(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not get all classes: %w", err)
	}

	return classes, nil
}
