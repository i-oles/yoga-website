package sqlite

import (
	"context"
	"errors"
	"fmt"
	"main/internal/domain/models"
	"main/internal/infrastructure/errs"
	dbModels "main/internal/infrastructure/models/db/classes"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ClassesRepo struct {
	db *gorm.DB
}

func NewClassesRepo(db *gorm.DB) *ClassesRepo {
	return &ClassesRepo{
		db: db,
	}
}

func (c ClassesRepo) GetAll(ctx context.Context) ([]models.Class, error) {
	var sqlClasses []dbModels.SQLClass

	if err := c.db.WithContext(ctx).Order("start_time ASC").Find(&sqlClasses).Error; err != nil {
		return nil, fmt.Errorf("could not get all classes: %w", err)
	}

	classes := make([]models.Class, len(sqlClasses))

	for i, sqlClass := range sqlClasses {
		classes[i] = sqlClass.ToDomain()
	}

	return classes, nil
}

func (c ClassesRepo) Get(ctx context.Context, id uuid.UUID) (models.Class, error) {
	var sqlClass dbModels.SQLClass

	if err := c.db.WithContext(ctx).First(&sqlClass, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.Class{}, errs.ErrNotFound
		}
		return models.Class{}, fmt.Errorf("could not get class: %w", err)
	}

	return sqlClass.ToDomain(), nil
}

func (c ClassesRepo) DecrementCurrentCapacity(ctx context.Context, id uuid.UUID) error {
	result := c.db.WithContext(ctx).
		Model(&dbModels.SQLClass{}).
		Where("id = ? AND current_capacity > 0", id).
		Update("current_capacity", gorm.Expr("current_capacity - ?", 1))

	if result.Error != nil {
		return fmt.Errorf("could not decrement current capacity: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return errs.ErrNoRowsAffected
	}

	return nil
}

func (c ClassesRepo) IncrementCurrentCapacity(ctx context.Context, id uuid.UUID) error {
	result := c.db.WithContext(ctx).
		Model(&dbModels.SQLClass{}).
		Where("id = ? AND current_capacity < max_capacity", id).
		Update("current_capacity", gorm.Expr("current_capacity + ?", 1))

	if result.Error != nil {
		return fmt.Errorf("could not increment current capacity: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return errs.ErrNoRowsAffected
	}

	return nil
}

func (c ClassesRepo) Insert(ctx context.Context, classes []models.Class) ([]models.Class, error) {
	sqlClass := make([]dbModels.SQLClass, len(classes))
	for i, class := range classes {
		sqlClass[i] = dbModels.FromDomain(class)
	}

	err := c.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&sqlClass).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("could not insert classes: %w", err)
	}

	insertedClasses := make([]models.Class, len(sqlClass))
	for i, SQLClass := range sqlClass {
		insertedClasses[i] = SQLClass.ToDomain()
	}

	return insertedClasses, nil
}
