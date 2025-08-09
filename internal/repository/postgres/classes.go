package postgres

import (
	"database/sql"
	"errors"
	"fmt"
	"main/internal/repository"

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

func (c ClassesRepo) GetAll() ([]repository.Class, error) {
	query := fmt.Sprintf("SELECT * FROM %s ORDER BY id ASC;", c.collName)

	rows, err := c.db.Query(query)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	classes := make([]repository.Class, 0)

	for rows.Next() {
		class := repository.Class{}
		err = rows.Scan(
			&class.ID,
			&class.DayOfWeek,
			&class.StartTime,
			&class.ClassLevel,
			&class.ClassCategory,
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

func (c ClassesRepo) Get(id uuid.UUID) (repository.Class, error) {
	query := fmt.Sprintf("SELECT * FROM %s WHERE id = $1;", c.collName)

	var class repository.Class
	err := c.db.QueryRow(query, id).Scan(
		&class.ID,
		&class.DayOfWeek,
		&class.StartTime,
		&class.ClassLevel,
		&class.ClassCategory,
		&class.MaxCapacity,
		&class.Location,
	)
	if err != nil {
		return repository.Class{}, err
	}

	return class, nil
}

func (c ClassesRepo) DecrementMaxCapacity(id uuid.UUID) error {
	query := fmt.Sprintf(
		"UPDATE %s SET max_capacity = max_capacity - 1 WHERE id = $1 AND max_capacity > 0",
		c.collName)

	result, err := c.db.Exec(query, id)
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
