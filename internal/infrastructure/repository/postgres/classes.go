package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"main/internal/domain/models"

	"github.com/google/uuid"
)

type ClassesRepo struct {
	db       *sql.DB
	collName string
}

func NewClassesRepo(db *sql.DB) *ClassesRepo {
	return &ClassesRepo{
		db: db,
		//TODO: move to config
		collName: "classes"}
}

func (c ClassesRepo) GetAll(ctx context.Context) ([]models.Class, error) {
	query := fmt.Sprintf("SELECT * FROM %s ORDER BY start_time ASC;", c.collName)

	rows, err := c.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	classes := make([]models.Class, 0)

	for rows.Next() {
		class := models.Class{}
		err = rows.Scan(
			&class.ID,
			&class.StartTime,
			&class.ClassLevel,
			&class.ClassName,
			&class.CurrentCapacity,
			&class.MaxCapacity,
			&class.Location,
		)
		if err != nil {
			return nil, err
		}

		classes = append(classes, class)
	}

	return classes, nil
}

func (c ClassesRepo) Get(ctx context.Context, id uuid.UUID) (models.Class, error) {
	query := fmt.Sprintf("SELECT * FROM %s WHERE id = $1;", c.collName)

	var class models.Class
	err := c.db.QueryRowContext(ctx, query, id).Scan(
		&class.ID,
		&class.StartTime,
		&class.ClassLevel,
		&class.ClassName,
		&class.CurrentCapacity,
		&class.MaxCapacity,
		&class.Location,
	)
	if err != nil {
		return models.Class{}, err
	}

	return class, nil
}

func (c ClassesRepo) DecrementCurrentCapacity(ctx context.Context, id uuid.UUID) error {
	query := fmt.Sprintf(
		"UPDATE %s SET current_capacity = current_capacity - 1 WHERE id = $1 AND max_capacity > 0",
		c.collName)

	result, err := c.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("no rows affected")
	}

	return nil
}

func (c ClassesRepo) IncrementCurrentCapacity(ctx context.Context, id uuid.UUID) error {
	query := fmt.Sprintf(
		"UPDATE %s SET current_capacity = current_capacity + 1 WHERE id = $1 AND current_capacity < max_capacity",
		c.collName)

	result, err := c.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("no rows affected")
	}

	return nil
}

func (c ClassesRepo) Insert(ctx context.Context, classes []models.Class) ([]models.Class, error) {
	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("could not begin transaction: %w", err)
	}

	defer tx.Rollback()

	query := fmt.Sprintf(
		"INSERT INTO %s (id, start_time, class_level, class_name, current_capacity, max_capacity, location) VALUES ($1, $2, $3, $4, $5, $6, $7)", c.collName)

	inserted := make([]models.Class, 0, len(classes))

	for _, class := range classes {
		_, err = tx.ExecContext(
			ctx,
			query,
			class.ID,
			class.StartTime,
			class.ClassLevel,
			class.ClassName,
			class.CurrentCapacity,
			class.MaxCapacity,
			class.Location,
		)
		if err != nil {
			return nil, fmt.Errorf("could not insert class: %w", err)
		}

		inserted = append(inserted, class)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("could not commit transaction: %w", err)
	}

	return inserted, nil
}
