package sqlite

import (
	"context"
	"errors"
	"fmt"
	"main/internal/domain/models"
	"main/internal/infrastructure/errs"
	"main/internal/infrastructure/models/db/confirmedbookings"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ConfirmedBookingsRepo struct {
	db *gorm.DB
}

func NewConfirmedBookingsRepo(db *gorm.DB) ConfirmedBookingsRepo {
	return ConfirmedBookingsRepo{
		db: db,
	}
}

func (r ConfirmedBookingsRepo) Get(
	ctx context.Context,
	classID uuid.UUID,
	email string,
) (models.ConfirmedBooking, error) {
	var sqlConfirmedBooking confirmedbookings.SQLConfirmedBooking

	tx := r.db.WithContext(ctx).
		Where("class_id = ? AND email = ?", classID, email).
		First(&sqlConfirmedBooking)

	if tx.Error != nil {
		if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
			return models.ConfirmedBooking{}, errs.ErrNotFound
		}

		return models.ConfirmedBooking{}, fmt.Errorf("failed to get confirmed booking: %w", tx.Error)
	}

	return sqlConfirmedBooking.ToDomain(), nil
}

func (r ConfirmedBookingsRepo) GetAll(ctx context.Context) ([]models.ConfirmedBooking, error) {
	var SQLConfirmedBookings []confirmedbookings.SQLConfirmedBooking

	if err := r.db.WithContext(ctx).Find(&SQLConfirmedBookings).Error; err != nil {
		return nil, fmt.Errorf("failed to get all confirmed bookings: %w", err)
	}

	confirmedBookings := make([]models.ConfirmedBooking, len(SQLConfirmedBookings))

	for i, SQLConfirmedBooking := range SQLConfirmedBookings {
		confirmedBookings[i] = SQLConfirmedBooking.ToDomain()
	}

	return confirmedBookings, nil
}

func (r ConfirmedBookingsRepo) Insert(
	ctx context.Context,
	confirmedBooking models.ConfirmedBooking,
) error {
	sqlConfirmedBooking := confirmedbookings.FromDomain(confirmedBooking)

	if err := r.db.WithContext(ctx).Create(&sqlConfirmedBooking).Error; err != nil {
		return fmt.Errorf("failed to insert confirmed booking: %w", err)
	}

	return nil
}

func (r ConfirmedBookingsRepo) Delete(ctx context.Context, classID uuid.UUID, email string) error {
	var sqlConfirmedBooking confirmedbookings.SQLConfirmedBooking

	tx := r.db.WithContext(ctx).
		Where("class_id = ? AND email = ?", classID, email).
		Delete(&sqlConfirmedBooking)
	if tx.Error != nil {
		return fmt.Errorf("failed to delete confirmed booking: %w", tx.Error)
	}

	if tx.RowsAffected == 0 {
		return errs.ErrNoRowsAffected
	}

	return nil
}
