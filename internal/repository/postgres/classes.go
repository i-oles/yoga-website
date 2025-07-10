package postgres

import (
	"database/sql"
	"fmt"
	"main/internal/repository"
)

type ClassesRepo struct {
	db       *sql.DB
	collName string
}

func NewClassesRepo(db *sql.DB) *ClassesRepo {
	return &ClassesRepo{
		db:       db,
		collName: "classes"}
}

func (c ClassesRepo) GetAll() ([]repository.Class, error) {
	query := fmt.Sprintf("SELECT * FROM %s;", c.collName)

	rows, err := c.db.Query(query)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	classes := make([]repository.Class, 0)

	for rows.Next() {
		class := repository.Class{}
		err = rows.Scan(&class.ID, &class.Day, &class.Datetime, &class.Level, &class.Type, &class.SpotsLeft, &class.Place)
		if err != nil {
			return nil, err
		}

		classes = append(classes, class)
	}

	return classes, nil
}

func (c ClassesRepo) Update(update repository.UpdateClass) error {
	return nil
}
