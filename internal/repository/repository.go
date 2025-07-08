package repository

import "time"

type Class struct {
	ID        int       `db:"id"`
	Datetime  time.Time `db:"datetime"`
	Day       string    `db:"day"`
	Level     string    `db:"level"`
	Type      string    `db:"type"`
	SpotsLeft int       `db:"spotsLeft"`
	Place     string    `db:"place"`
}

type Classes interface {
	GetCurrentMonthClasses() ([]Class, error)
}
