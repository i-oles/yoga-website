package sqlite

import (
	"context"
	"errors"
	"fmt"

	"main/internal/domain/models"
	"main/internal/infrastructure/models/db"
	"main/pkg/optional"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type passesRepo struct {
	db *gorm.DB
}

func NewPassesRepo(db *gorm.DB) *passesRepo {
	return &passesRepo{
		db: db,
	}
}

func (r *passesRepo) GetByEmail(
	ctx context.Context, email string,
) (optional.Optional[models.Pass], error) {
	var sqlPass db.SQLPass

	result := r.db.WithContext(ctx).
		Where("email = ?", email).
		Order("created_at DESC").
		First(&sqlPass)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return optional.Empty[models.Pass](), nil
		}

		return optional.Empty[models.Pass](), fmt.Errorf("could not get pass: %w", result.Error)
	}

	return optional.Of(sqlPass.ToDomain()), nil
}

func (r *passesRepo) Update(
	ctx context.Context, pass models.Pass,
) (models.Pass, error) {
	var sqlPass db.SQLPass

	update := db.SQLPass{
		TotalSlots: pass.TotalSlots,
	}

	if err := r.db.WithContext(ctx).
		Model(&sqlPass).
		Clauses(clause.Returning{}).
		Where("id = ?", pass.ID).
		Updates(update).
		Error; err != nil {
		return models.Pass{},
			fmt.Errorf("could not update pass: %v, %w", pass.ID, err)
	}

	return sqlPass.ToDomain(), nil
}

func (r *passesRepo) Insert(
	ctx context.Context,
	email string,
	totalSlots int,
) (models.Pass, error) {
	pass := db.SQLPass{
		Email:      email,
		TotalSlots: totalSlots,
	}

	if err := r.db.WithContext(ctx).Create(&pass).Error; err != nil {
		return models.Pass{}, fmt.Errorf("could not insert pass: %w", err)
	}

	return pass.ToDomain(), nil
}
