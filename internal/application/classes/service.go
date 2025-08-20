package classes

import (
	"context"
	"fmt"
	"main/internal/domain/models"
	"main/internal/domain/repositories"
	"time"
)

const classesLimit = 6

type Service struct {
	classesRepo repositories.Classes
}

func New(classesRepo repositories.Classes) *Service {
	return &Service{classesRepo: classesRepo}
}

func (s *Service) GetAllClasses(ctx context.Context) ([]models.Class, error) {
	filteredClasses := make([]models.Class, 0)

	classes, err := s.classesRepo.GetAllClasses(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not get all classes: %w", err)
	}

	counter := 0

	for _, class := range classes {
		if counter >= classesLimit {
			break
		}

		classStartTime := class.StartTime

		loc, err := time.LoadLocation("Europe/Warsaw")
		if err != nil {
			panic(fmt.Errorf("could not load location: %w", err))
		}

		warsawTime := classStartTime.In(loc)

		now := time.Now()

		fmt.Println(now)

		class.StartTime = warsawTime

		if class.StartTime.After(time.Now()) {
			filteredClasses = append(filteredClasses, class)
			counter++
		}
	}

	return filteredClasses, nil
}
