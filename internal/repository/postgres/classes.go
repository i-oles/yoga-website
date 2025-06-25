package postgres

import (
	"database/sql"
	"fmt"
	"time"
)

type Class struct {
	ID        int       `db:"id"`
	Date      string    `db:"datetime"`
	Day       time.Time `db:"day"`
	Level     string    `db:"level"`
	SpotsLeft int       `db:"spotsLeft"`
	Place     string    `db:"place"`
}

type Classes struct {
	db       *sql.DB
	collName string
}

func NewClasses(db *sql.DB) *Classes {
	return &Classes{
		db:       db,
		collName: "classes"}
}

func (c Classes) GetCurrentMonthClasses() ([]Class, error) {
	query := fmt.Sprintf("SELECT * FROM %s WHERE EXTRACT(YEAR FROM datetime) = EXTRACT(YEAR FROM current_date) AND EXTRACT(MONTH FROM datetime) = EXTRACT(MONTH FROM current_date);", c.collName)

	rows, err := c.db.Query(query)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	classes := make([]Class, 0)

	for rows.Next() {
		class := Class{}
		err = rows.Scan(&class.ID, &class.Date, &class.Day, &class.Level, &class.SpotsLeft, &class.Place)
		if err != nil {
			return nil, err
		}

		classes = append(classes, class)
	}

	return classes, nil
}
