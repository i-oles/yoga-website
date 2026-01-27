package sqlite

import (
	"context"
	"errors"
	"fmt"

	"main/internal/domain/models"
	"main/internal/infrastructure/models/db"
	"main/pkg/optional"

	"gorm.io/gorm"
)

type PassesRepo struct {
	db *gorm.DB
}

func NewPassesRepo(db *gorm.DB) PassesRepo {
	return PassesRepo{
		db: db,
	}
}

func (r PassesRepo) GetByEmail(ctx context.Context, email string) (optional.Optional[models.Pass], error) {
	var sqlPass db.SQLPass

	result := r.db.WithContext(ctx).Where("email = ?", email).First(&sqlPass)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return optional.Empty[models.Pass](), nil
		}

		return optional.Empty[models.Pass](), fmt.Errorf("could not get pass: %w", result.Error)
	}

	return optional.Of(sqlPass.ToDomain()), nil
}

func (r PassesRepo) Update(ctx context.Context, id int, update map[string]any) error {
	var pass db.SQLPass

	result := r.db.WithContext(ctx).Model(&pass).Where("id = ?", id).Updates(update)
	if result.Error != nil {
		return fmt.Errorf("could not update pass: %v with data: %v, %w", id, update, result.Error)
	}

	return nil
}

func (r PassesRepo) Insert(ctx context.Context, email string, usedCredits, totalCredits int) (models.Pass, error) {
	pass := db.SQLPass{
		Email:        email,
		UsedCredits:  usedCredits,
		TotalCredits: totalCredits,
	}

	if err := r.db.WithContext(ctx).Create(&pass).Error; err != nil {
		return models.Pass{}, fmt.Errorf("could not insert pass: %w", err)
	}

	return pass.ToDomain(), nil
}
