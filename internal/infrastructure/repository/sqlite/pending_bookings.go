package sqlite

import (
	"context"
	"errors"
	"fmt"
	"main/internal/domain/models"
	"main/internal/infrastructure/errs"
	"main/internal/infrastructure/models/db/pendingbookings"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PendingBookingsRepo struct {
	db *gorm.DB
}

func NewPendingBookingsRepo(db *gorm.DB) *PendingBookingsRepo {
	return &PendingBookingsRepo{
		db: db,
	}
}

func (r PendingBookingsRepo) Insert(
	ctx context.Context,
	pendingBooking models.PendingBooking,
) error {
	sqlPendingBooking := pendingbookings.FromDomain(pendingBooking)

	if err := r.db.WithContext(ctx).Create(&sqlPendingBooking).Error; err != nil {
		return fmt.Errorf("could not insert pending booking: %w", err)
	}

	return nil
}

func (r PendingBookingsRepo) GetByConfirmationToken(ctx context.Context, token string) (models.PendingBooking, error) {
	var sqlPendingBooking pendingbookings.SQLPendingBooking

	if err := r.db.WithContext(ctx).
		Where("confirmation_token = ?", token).
		First(&sqlPendingBooking).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.PendingBooking{}, errs.ErrNotFound
		}

		return models.PendingBooking{}, fmt.Errorf("could not get pending booking: %w", err)
	}

	return sqlPendingBooking.ToDomain(), nil
}

func (r PendingBookingsRepo) CountPendingBookingsPerUser(
	ctx context.Context,
	email string,
	operation models.Operation,
	classID uuid.UUID,
) (int8, error) {
	var count int64

	if err := r.db.WithContext(ctx).
		Model(&pendingbookings.SQLPendingBooking{}).
		Where("email = ? AND class_id = ? AND operation = ?", email, classID, operation).
		Count(&count).Error; err != nil {
		return 0, fmt.Errorf("could not count pending bookings: %w", err)
	}

	return int8(count), nil
}

func (r PendingBookingsRepo) Delete(ctx context.Context, id uuid.UUID) error {
	var sqlPendingBooking pendingbookings.SQLPendingBooking

	tx := r.db.WithContext(ctx).Where("id = ?", id).Delete(&sqlPendingBooking)
	if tx.Error != nil {
		return fmt.Errorf("could not delete pending booking: %w", tx.Error)
	}

	if tx.RowsAffected == 0 {
		return errs.ErrNoRowsAffected
	}

	return nil
}
