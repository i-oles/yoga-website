package repositories

import (
	"context"
	"main/internal/domain/models"

	"github.com/google/uuid"
)

type IClasses interface {
	Get(ctx context.Context, id uuid.UUID) (models.Class, error)
	GetAll(ctx context.Context) ([]models.Class, error)
	Insert(ctx context.Context, classes []models.Class) ([]models.Class, error)
	DecrementCurrentCapacity(ctx context.Context, id uuid.UUID) error
	IncrementCurrentCapacity(ctx context.Context, id uuid.UUID) error
}

type IBookings interface {
	Get(ctx context.Context, classID uuid.UUID, email string) (models.Booking, error)
	GetAll(ctx context.Context) ([]models.Booking, error)
	Insert(ctx context.Context, confirmedBooking models.Booking) error
	Delete(ctx context.Context, classID uuid.UUID, email string) error
}

type IPendingBookings interface {
	Get(ctx context.Context, token string) (models.PendingBooking, error)
	Insert(ctx context.Context, booking models.PendingBooking) error
	Delete(ctx context.Context, id uuid.UUID) error
	CountPendingBookingsPerUser(
		ctx context.Context,
		email string,
		operation models.Operation,
		classID uuid.UUID,
	) (int8, error)
}
