package sqlite

import (
	"context"
	"errors"
	"fmt"
	"main/internal/domain/models"
	"main/internal/infrastructure/errs"
	"main/internal/infrastructure/models/db/bookings"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BookingsRepo struct {
	db *gorm.DB
}

func NewBookingsRepo(db *gorm.DB) BookingsRepo {
	return BookingsRepo{
		db: db,
	}
}

func (r BookingsRepo) Get(
	ctx context.Context,
	classID uuid.UUID,
	email string,
) (models.Booking, error) {
	var sqlConfirmedBooking bookings.SQLBooking

	tx := r.db.WithContext(ctx).
		Where("class_id = ? AND email = ?", classID, email).
		First(&sqlConfirmedBooking)

	if tx.Error != nil {
		if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
			return models.Booking{}, errs.ErrNotFound
		}

		return models.Booking{}, fmt.Errorf("failed to get confirmed booking: %w", tx.Error)
	}

	return sqlConfirmedBooking.ToDomain(), nil
}

func (r BookingsRepo) GetAll(ctx context.Context) ([]models.Booking, error) {
	var SQLConfirmedBookings []bookings.SQLBooking

	if err := r.db.WithContext(ctx).Find(&SQLConfirmedBookings).Error; err != nil {
		return nil, fmt.Errorf("failed to get all confirmed bookings: %w", err)
	}

	confirmedBookings := make([]models.Booking, len(SQLConfirmedBookings))

	for i, SQLConfirmedBooking := range SQLConfirmedBookings {
		confirmedBookings[i] = SQLConfirmedBooking.ToDomain()
	}

	return confirmedBookings, nil
}

func (r BookingsRepo) Insert(
	ctx context.Context,
	confirmedBooking models.Booking,
) error {
	sqlConfirmedBooking := bookings.FromDomain(confirmedBooking)

	if err := r.db.WithContext(ctx).Create(&sqlConfirmedBooking).Error; err != nil {
		return fmt.Errorf("failed to insert confirmed booking: %w", err)
	}

	return nil
}

func (r BookingsRepo) Delete(ctx context.Context, classID uuid.UUID, email string) error {
	var sqlConfirmedBooking bookings.SQLBooking

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
