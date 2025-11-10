package sqlite

import (
	"context"
	"errors"
	"fmt"
	"main/internal/domain/models"
	"main/internal/infrastructure/errs"
	"main/internal/infrastructure/models/db"

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

func (r BookingsRepo) Get(ctx context.Context, id uuid.UUID) (models.Booking, error) {
	var sqlBooking db.SQLBooking

	tx := r.db.WithContext(ctx).Where("id = ?", id).Preload("Class").First(&sqlBooking)

	if tx.Error != nil {
		if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
			return models.Booking{}, errs.ErrNotFound
		}

		return models.Booking{},
			fmt.Errorf("could not get booking for id %s: %w", id, tx.Error)
	}

	return sqlBooking.ToDomain(), nil
}

func (r BookingsRepo) GetByEmailAndClassID(
	ctx context.Context,
	classID uuid.UUID,
	email string,
) (models.Booking, error) {
	var sqlBooking db.SQLBooking

	tx := r.db.WithContext(ctx).
		Where("class_id = ? AND email = ?", classID, email).
		First(&sqlBooking)

	if tx.Error != nil {
		if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
			return models.Booking{}, errs.ErrNotFound
		}

		return models.Booking{},
			fmt.Errorf("could not get booking for email %s and classID %s: %w", email, classID, tx.Error)
	}

	return sqlBooking.ToDomain(), nil
}

func (r BookingsRepo) GetAll(ctx context.Context) ([]models.Booking, error) {
	var SQLBookings []db.SQLBooking

	if err := r.db.WithContext(ctx).Find(&SQLBookings).Error; err != nil {
		return nil, fmt.Errorf("could not get all bookings: %w", err)
	}

	result := make([]models.Booking, len(SQLBookings))

	for i, SQLBooking := range SQLBookings {
		result[i] = SQLBooking.ToDomain()
	}

	return result, nil
}

func (r BookingsRepo) GetAllByClassID(ctx context.Context, classID uuid.UUID) ([]models.Booking, error) {
	var SQLBookings []db.SQLBooking

    if err := r.db.WithContext(ctx).
        Preload("Class").
        Where("class_id = ?", classID).
        Find(&SQLBookings).Error; err != nil {
        return nil, fmt.Errorf("could not get bookings for classID %s: %w", classID, err)
    }

	result := make([]models.Booking, len(SQLBookings))

	for i, SQLBooking := range SQLBookings {
		result[i] = SQLBooking.ToDomain()
	}

	return result, nil
}

func (r BookingsRepo) Insert(
	ctx context.Context,
	booking models.Booking,
) (uuid.UUID, error) {
	sqlBooking := db.SQLBookingsFromDomain(booking)

	if err := r.db.WithContext(ctx).Create(&sqlBooking).Error; err != nil {
		return uuid.Nil, fmt.Errorf("could not insert booking: %w", err)
	}

	return booking.ID, nil
}

func (r BookingsRepo) Delete(ctx context.Context, id uuid.UUID) error {
	var sqlBooking db.SQLBooking

	tx := r.db.WithContext(ctx).
		Where("id = ?", id).
		Delete(&sqlBooking)
	if tx.Error != nil {
		return fmt.Errorf("could not delete booking: %w", tx.Error)
	}

	if tx.RowsAffected == 0 {
		return errs.ErrNoRowsAffected
	}

	return nil
}
