package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"main/internal/domain/models"
	"time"

	"github.com/google/uuid"
)

type ConfirmedBookingsRepo struct {
	db       *sql.DB
	collName string
}

func NewConfirmedBookingsRepo(db *sql.DB) ConfirmedBookingsRepo {
	return ConfirmedBookingsRepo{
		db: db,
		//TODO: move to config
		collName: "confirmed_bookings"}
}

func (r ConfirmedBookingsRepo) Insert(
	ctx context.Context,
	confirmedBooking models.ConfirmedBooking,
) error {
	query := fmt.Sprintf("INSERT INTO %s (id, class_id, first_name, last_name, email, created_at) VALUES ($1, $2, $3, $4, $5, $6);", r.collName)

	_, err := r.db.ExecContext(
		ctx,
		query,
		confirmedBooking.ID,
		confirmedBooking.ClassID,
		confirmedBooking.FirstName,
		confirmedBooking.LastName,
		confirmedBooking.Email,
		time.Now(),
	)
	if err != nil {
		return fmt.Errorf("could not insert confirmed confirmation: %w", err)
	}

	return nil
}

func (r ConfirmedBookingsRepo) Delete(ctx context.Context, classID uuid.UUID, email string) error {
	query := fmt.Sprintf("DELETE FROM %s WHERE class_id=$1 AND email=$2", r.collName)
	_, err := r.db.ExecContext(ctx, query, classID, email)
	if err != nil {
		return fmt.Errorf("could not delete confirmed booking: %w", err)
	}

	return nil
}
