package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"main/internal/repository"
	"main/pkg/optional"
)

type PendingBookingsRepo struct {
	db       *sql.DB
	collName string
}

func NewPendingBookingsRepo(db *sql.DB) *PendingBookingsRepo {
	return &PendingBookingsRepo{
		db:       db,
		collName: "pending_bookings",
	}
}

func (r PendingBookingsRepo) Insert(
	ctx context.Context,
	pendingBooking repository.PendingBooking,
) error {
	query := fmt.Sprintf("INSERT INTO %s (class_id, class_type, place, date, token, name, last_name, email, expires_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9);", r.collName)

	_, err := r.db.ExecContext(
		ctx,
		query,
		pendingBooking.ClassID,
		pendingBooking.ClassType,
		pendingBooking.Place,
		pendingBooking.Date,
		pendingBooking.Token,
		pendingBooking.Name,
		pendingBooking.LastName,
		pendingBooking.Email,
		pendingBooking.ExpiresAt,
	)
	if err != nil {
		return fmt.Errorf("insert pending booking: %w", err)
	}

	return nil
}

func (r PendingBookingsRepo) Get(ctx context.Context, token string) (optional.Optional[repository.PendingBooking], error) {
	var booking repository.PendingBooking

	query := fmt.Sprintf("SELECT class_id, class_type, date, place, email, name, last_name FROM %s WHERE token = $1;", r.collName)

	err := r.db.QueryRowContext(ctx, query, token).Scan(
		&booking.ClassID,
		&booking.ClassType,
		&booking.Date,
		&booking.Place,
		&booking.Email,
		&booking.Name,
		&booking.LastName,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return optional.Empty[repository.PendingBooking](), fmt.Errorf("pending booking does not exists in database: %w", err)
		}

		return optional.Empty[repository.PendingBooking](), fmt.Errorf("failed while getting pending booking: %w", err)
	}

	return optional.Of(booking), nil
}

func (r PendingBookingsRepo) Delete(ctx context.Context, token string) error {
	query := fmt.Sprintf("DELETE FROM %s WHERE token = $1;", r.collName)

	_, err := r.db.ExecContext(ctx, query, token)
	if err != nil {
		return fmt.Errorf("delete pending booking: %w", err)
	}

	return nil
}
