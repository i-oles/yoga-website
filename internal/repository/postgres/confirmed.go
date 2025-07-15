package postgres

import (
	"context"
	"database/sql"
	"fmt"
)

type ConfirmedBookingsRepo struct {
	db       *sql.DB
	collName string
}

func NewConfirmedBookingsRepo(db *sql.DB) ConfirmedBookingsRepo {
	return ConfirmedBookingsRepo{
		db:       db,
		collName: "confirmed_bookings"}
}

func (r ConfirmedBookingsRepo) Insert(
	ctx context.Context,
	classID int,
	name, lastName, email string,
) error {
	query := fmt.Sprintf("INSERT INTO %s (class_id, name, last_name, email) VALUES ($1, $2, $3, $4);", r.collName)

	_, err := r.db.ExecContext(ctx, query, classID, name, lastName, email)
	if err != nil {
		return fmt.Errorf("insert booking: %w", err)
	}

	return nil
}
