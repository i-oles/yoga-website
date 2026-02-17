package repositories

import (
	"context"

	"main/internal/domain/models"
	"main/pkg/optional"

	"github.com/google/uuid"
)

type IClasses interface {
	Get(ctx context.Context, id uuid.UUID) (models.Class, error)
	List(ctx context.Context) ([]models.Class, error)
	Insert(ctx context.Context, classes []models.Class) ([]models.Class, error)
	Delete(ctx context.Context, id uuid.UUID) error
	Update(ctx context.Context, id uuid.UUID, update map[string]any) error
}

type IBookings interface {
	GetByID(ctx context.Context, id uuid.UUID) (models.Booking, error)
	GetByEmailAndClassID(ctx context.Context, classID uuid.UUID, email string) (models.Booking, error)
	GetIDsByEmail(ctx context.Context, email string, limit int) ([]uuid.UUID, error)
	List(ctx context.Context) ([]models.Booking, error)
	ListByClassID(ctx context.Context, classID uuid.UUID) ([]models.Booking, error)
	CountForClassID(ctx context.Context, classID uuid.UUID) (int, error)
	Insert(ctx context.Context, confirmedBooking models.Booking) (uuid.UUID, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type IPendingBookings interface {
	GetByConfirmationToken(ctx context.Context, token string) (models.PendingBooking, error)
	Insert(ctx context.Context, booking models.PendingBooking) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context) ([]models.PendingBooking, error)
}

type IPasses interface {
	GetByEmail(ctx context.Context, email string) (optional.Optional[models.Pass], error)
	Update(ctx context.Context, id int, usedBookingIDs []uuid.UUID, totalBookings int) error
	Insert(
		ctx context.Context, email string, usedBookingIDs []uuid.UUID, totalBookings int,
	) (models.Pass, error)
}
