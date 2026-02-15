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

type bookingsRepo struct {
	db *gorm.DB
}

func NewBookingsRepo(db *gorm.DB) *bookingsRepo {
	return &bookingsRepo{
		db: db,
	}
}

func (r *bookingsRepo) GetByID(
	ctx context.Context, bookingID uuid.UUID,
) (models.Booking, error) {
	var sqlBooking db.SQLBooking

	result := r.db.WithContext(ctx).Where("id = ?", bookingID).Preload("Class").First(&sqlBooking)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return models.Booking{}, errs.ErrNotFound
		}

		return models.Booking{},
			fmt.Errorf("could not get booking for id %s: %w", bookingID, result.Error)
	}

	return sqlBooking.ToDomain(), nil
}

func (r *bookingsRepo) GetByEmailAndClassID(
	ctx context.Context,
	classID uuid.UUID,
	email string,
) (models.Booking, error) {
	var sqlBooking db.SQLBooking

	result := r.db.WithContext(ctx).
		Where("class_id = ? AND email = ?", classID, email).
		First(&sqlBooking)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return models.Booking{}, errs.ErrNotFound
		}

		return models.Booking{},
			fmt.Errorf("could not get booking by email %s, classID %s: %w", email, classID, result.Error)
	}

	return sqlBooking.ToDomain(), nil
}

func (r *bookingsRepo) GetIDsByEmail(
	ctx context.Context, email string, limit int) ([]uuid.UUID, error,
) {
	var sqlBookings []db.SQLBooking

	if limit <= 0 {
		return nil, fmt.Errorf("limit must be positive: %d", limit)
	}

	if err := r.db.WithContext(ctx).
		Select("id", "created_at").
		Where("email = ?", email).
		Order("created_at DESC").
		Limit(limit).
		Find(&sqlBookings).Error; err != nil {
		return nil, fmt.Errorf("could not get booking IDs for email %s: %w", email, err)
	}

	result := make([]uuid.UUID, len(sqlBookings))
	for i, booking := range sqlBookings {
		result[i] = booking.ID
	}

	return result, nil
}

func (r *bookingsRepo) List(ctx context.Context) ([]models.Booking, error) {
	var SQLBookings []db.SQLBooking

	if err := r.db.WithContext(ctx).Preload("Class").Find(&SQLBookings).Error; err != nil {
		return nil, fmt.Errorf("could not list bookings: %w", err)
	}

	result := make([]models.Booking, len(SQLBookings))

	for i, SQLBooking := range SQLBookings {
		result[i] = SQLBooking.ToDomain()
	}

	return result, nil
}

func (r *bookingsRepo) CountForClassID(ctx context.Context, classID uuid.UUID) (int, error) {
	var count int64

	var SQLBooking db.SQLBooking

	if err := r.db.WithContext(ctx).
		Model(&SQLBooking).
		Where("class_id = ?", classID).
		Count(&count).Error; err != nil {
		return 0, fmt.Errorf("could count bookings for classID %s: %w", classID, err)
	}

	return int(count), nil
}

func (r *bookingsRepo) ListByClassID(
	ctx context.Context,
	classID uuid.UUID,
) ([]models.Booking, error) {
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

func (r *bookingsRepo) Insert(
	ctx context.Context,
	booking models.Booking,
) (uuid.UUID, error) {
	sqlBooking := db.SQLBookingsFromDomain(booking)

	if err := r.db.WithContext(ctx).Create(&sqlBooking).Error; err != nil {
		return uuid.Nil, fmt.Errorf("could not insert booking: %w", err)
	}

	return booking.ID, nil
}

func (r *bookingsRepo) Delete(ctx context.Context, id uuid.UUID) error {
	var sqlBooking db.SQLBooking

	result := r.db.WithContext(ctx).
		Where("id = ?", id).
		Delete(&sqlBooking)
	if result.Error != nil {
		return fmt.Errorf("could not delete booking: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return errs.ErrNoRowsAffected
	}

	return nil
}
