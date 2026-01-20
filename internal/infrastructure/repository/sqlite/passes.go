package sqlite

import (
	"context"
	"errors"
	"fmt"

	"main/internal/domain/models"
	"main/internal/infrastructure/errs"
	"main/internal/infrastructure/models/db"

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

func (r PassesRepo) GetByEmail(ctx context.Context, email string) (models.Pass, error) {
	var sqlPass db.SQLPass

	tx := r.db.WithContext(ctx).Where("email = ?", email).First(&sqlPass)

	if tx.Error != nil {
		if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
			return models.Pass{}, errs.ErrNotFound
		}

		return models.Pass{},
			fmt.Errorf("could not get pass for email %s: %w", email, tx.Error)
	}

	return sqlPass.ToDomain(), nil
}

func (r PassesRepo) Update(ctx context.Context, id int, update map[string]any) error {
	if err := r.db.WithContext(ctx).
		Model(&db.SQLPass{}).
		Where("id = ?", id).
		Updates(update).Error; err != nil {
		return fmt.Errorf("could not update pass: %v with data: %v, %w", id, update, err)
	}

	return nil
}
