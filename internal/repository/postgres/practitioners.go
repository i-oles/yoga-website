package postgres

import (
	"context"
	"database/sql"
	"fmt"
)

type PractitionersRepo struct {
	db       *sql.DB
	collName string
}

func NewPractitionersRepo(db *sql.DB) *PractitionersRepo {
	return &PractitionersRepo{
		db:       db,
		collName: "practitioners"}
}

func (r *PractitionersRepo) Insert(
	ctx context.Context,
	classID int,
	name, lastName, email string,
) error {
	//TODO two separate funcs
	var exists bool
	checkQuery := fmt.Sprintf(`SELECT EXISTS(SELECT 1 FROM %s WHERE class_id = $1 AND email = $2)`, r.collName)

	err := r.db.QueryRowContext(ctx, checkQuery, classID, email).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check existing practitioner: %w", err)
	}

	if exists {
		return fmt.Errorf("'%s' already booked in this class", email)
	}

	query := fmt.Sprintf("INSERT INTO %s (class_id, name, last_name, email) VALUES ($1, $2, $3, $4);", r.collName)

	_, err = r.db.ExecContext(ctx, query, classID, name, lastName, email)
	if err != nil {
		return err
	}

	return nil
}
