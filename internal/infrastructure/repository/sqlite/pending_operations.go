package sqlite

import (
	"context"
	"errors"
	"fmt"
	"main/internal/domain/models"
	"main/internal/infrastructure/errs"
	"main/internal/infrastructure/models/db/pendingoperations"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PendingOperationsRepo struct {
	db *gorm.DB
}

func NewPendingOperationsRepo(db *gorm.DB) *PendingOperationsRepo {
	return &PendingOperationsRepo{
		db: db,
	}
}

func (r PendingOperationsRepo) Insert(
	ctx context.Context,
	pendingOperation models.PendingOperation,
) error {
	sqlPendingOperation := pendingoperations.FromDomain(pendingOperation)

	if err := r.db.WithContext(ctx).Create(&sqlPendingOperation).Error; err != nil {
		return fmt.Errorf("failed to insert pending operation: %w", err)
	}

	return nil
}

func (r PendingOperationsRepo) Get(ctx context.Context, token string) (models.PendingOperation, error) {
	var sqlPendingOperation pendingoperations.SQLPendingOperation

	if err := r.db.WithContext(ctx).
		Where("confirmation_token = ?", token).
		First(&sqlPendingOperation).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.PendingOperation{}, errs.ErrNotFound
		}

		return models.PendingOperation{}, fmt.Errorf("failed to get pending operation: %w", err)
	}

	return sqlPendingOperation.ToDomain(), nil
}

func (r PendingOperationsRepo) CountPendingOperationsPerUser(
	ctx context.Context,
	email string,
	operation models.Operation,
	classID uuid.UUID,
) (int8, error) {
	var count int64

	if err := r.db.WithContext(ctx).
		Model(&pendingoperations.SQLPendingOperation{}).
		Where("email = ? AND class_id = ? AND operation = ?", email, classID, operation).
		Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count pending operations: %w", err)
	}

	return int8(count), nil
}

func (r PendingOperationsRepo) Delete(ctx context.Context, id uuid.UUID) error {
	var sqlPendingOperation pendingoperations.SQLPendingOperation

	tx := r.db.WithContext(ctx).Where("id = ?", id).Delete(&sqlPendingOperation)
	if tx.Error != nil {
		return fmt.Errorf("failed to delete pending operation: %w", tx.Error)
	}

	if tx.RowsAffected == 0 {
		return errs.ErrNoRowsAffected
	}

	return nil
}
