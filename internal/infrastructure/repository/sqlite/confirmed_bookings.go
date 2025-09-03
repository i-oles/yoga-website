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
	var sqlBooking bookings.SQLBooking

	tx := r.db.WithContext(ctx).
		Where("class_id = ? AND email = ?", classID, email).
		First(&sqlBooking)

	if tx.Error != nil {
		if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
			return models.Booking{}, errs.ErrNotFound
		}

		return models.Booking{}, fmt.Errorf("could not get booking: %w", tx.Error)
	}

	return sqlBooking.ToDomain(), nil
}

func (r BookingsRepo) GetAll(ctx context.Context) ([]models.Booking, error) {
	var SQLBookings []bookings.SQLBooking

	if err := r.db.WithContext(ctx).Find(&SQLBookings).Error; err != nil {
		return nil, fmt.Errorf("could not get all bookings: %w", err)
	}

	bookings := make([]models.Booking, len(SQLBookings))

	for i, SQLBooking := range SQLBookings {
		bookings[i] = SQLBooking.ToDomain()
	}

	return bookings, nil
}

func (r BookingsRepo) Insert(
	ctx context.Context,
	booking models.Booking,
) error {
	sqlBooking := bookings.FromDomain(booking)

	if err := r.db.WithContext(ctx).Create(&sqlBooking).Error; err != nil {
		return fmt.Errorf("could not insert booking: %w", err)
	}

	return nil
}

func (r BookingsRepo) Delete(ctx context.Context, classID uuid.UUID, email string) error {
	var sqlBooking bookings.SQLBooking

	tx := r.db.WithContext(ctx).
		Where("class_id = ? AND email = ?", classID, email).
		Delete(&sqlBooking)
	if tx.Error != nil {
		return fmt.Errorf("could not delete booking: %w", tx.Error)
	}

	if tx.RowsAffected == 0 {
		return errs.ErrNoRowsAffected
	}

	return nil
}
