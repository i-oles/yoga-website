package repositories

import (
	"context"
	"main/internal/domain/models"
	"main/pkg/optional"

	"github.com/google/uuid"
)

// TODO: here should be ctx added
type Classes interface {
	GetAllClasses(ctx context.Context) ([]models.Class, error)
	Get(ctx context.Context, id uuid.UUID) (models.Class, error)
	DecrementMaxCapacity(ctx context.Context, id uuid.UUID) error
	IncrementMaxCapacity(ctx context.Context, id uuid.UUID) error
}

type ConfirmedBookings interface {
	Insert(ctx context.Context, confirmedBooking models.ConfirmedBooking) error
	Delete(ctx context.Context, classID uuid.UUID, email string) error
}

type PendingOperations interface {
	Insert(ctx context.Context, booking models.PendingOperation) error
	Get(ctx context.Context, token string) (optional.Optional[models.PendingOperation], error)
	Delete(ctx context.Context, token string) error
}
