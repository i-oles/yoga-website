package classes

import (
	"context"
	"fmt"
	"main/internal/domain/errs"
	"main/internal/domain/models"
	"main/internal/domain/repositories"
	"time"
)

const classesLimit = 6

type Service struct {
	classesRepo repositories.IClasses
}

func New(classesRepo repositories.IClasses) *Service {
	return &Service{classesRepo: classesRepo}
}

func (s *Service) GetAllClasses(ctx context.Context) ([]models.Class, error) {
	filteredClasses := make([]models.Class, 0)

	classes, err := s.classesRepo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not get all classes: %w", err)
	}

	counter := 0

	for _, class := range classes {
		if counter >= classesLimit {
			break
		}

		if class.StartTime.Before(time.Now()) {
			continue
		}

		filteredClasses = append(filteredClasses, class)
		counter++
	}

	return filteredClasses, nil
}

func (s *Service) CreateClasses(ctx context.Context, classes []models.Class) ([]models.Class, error) {
	err := s.validateClasses(classes)
	if err != nil {
		return nil, errs.ErrClassBadRequest(err)
	}

	insertedClasses, err := s.classesRepo.Insert(ctx, classes)
	if err != nil {
		return nil, fmt.Errorf("could not insert classes: %w", err)
	}

	return insertedClasses, nil
}

func (s *Service) validateClasses(classes []models.Class) error {
	for _, class := range classes {
		if class.StartTime.Before(time.Now()) {
			return fmt.Errorf("class start_time: %v expired", class.StartTime)
		}

		if class.CurrentCapacity != class.MaxCapacity {
			return fmt.Errorf("%d != %d: current_capacity should be equal to max_capacity when creating class",
				class.CurrentCapacity, class.MaxCapacity,
			)
		}
	}

	return nil
}
