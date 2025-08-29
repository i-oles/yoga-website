package repositories

import (
	"context"
	"main/internal/domain/models"

	"github.com/google/uuid"
)

type Classes interface {
	Get(ctx context.Context, id uuid.UUID) (models.Class, error)
	GetAll(ctx context.Context) ([]models.Class, error)
	Insert(ctx context.Context, classes []models.Class) ([]models.Class, error)
	DecrementCurrentCapacity(ctx context.Context, id uuid.UUID) error
	IncrementCurrentCapacity(ctx context.Context, id uuid.UUID) error
}

type ConfirmedBookings interface {
	Get(ctx context.Context, classID uuid.UUID, email string) (models.ConfirmedBooking, error)
	Insert(ctx context.Context, confirmedBooking models.ConfirmedBooking) error
	Delete(ctx context.Context, classID uuid.UUID, email string) error
}

type PendingOperations interface {
	Get(ctx context.Context, token string) (models.PendingOperation, error)
	Insert(ctx context.Context, booking models.PendingOperation) error
	Delete(ctx context.Context, id uuid.UUID) error
	CountPendingOperationsPerUser(
		ctx context.Context,
		email string,
		operation models.Operation,
		classID uuid.UUID,
	) (int8, error)
}
