package repositories

import (
	"context"
	"main/internal/domain/models"

	"github.com/google/uuid"
)

// TODO: here should be ctx added
type Classes interface {
	Get(ctx context.Context, id uuid.UUID) (models.Class, error)
	GetAllClasses(ctx context.Context) ([]models.Class, error)
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
	Delete(ctx context.Context, token string) error
}
