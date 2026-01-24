package sqlite

import (
	"context"
	"errors"
	"fmt"

	"main/internal/domain/models"
	"main/internal/infrastructure/errs"
	"main/internal/infrastructure/models/db"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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

func (r PassesRepo) Upsert(ctx context.Context, email string, usedCredits, totalCredits int) (models.Pass, error) {
	pass := db.SQLPass{
		Email:        email,
		UsedCredits:  usedCredits,
		TotalCredits: totalCredits,
	}

	result := r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "email"}},
			DoUpdates: clause.AssignmentColumns([]string{"used_credits", "total_credits"}),
		}).
		Create(&pass)

	if err := result.Error; err != nil {
		return models.Pass{}, fmt.Errorf("could not upsert pass: %w", err)
	}

	return pass.ToDomain(), nil
}
