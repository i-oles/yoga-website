package repository

import (
	"context"
	"time"
)

type Class struct {
	ID        int       `db:"id"`
	Datetime  time.Time `db:"datetime"`
	Day       string    `db:"day"`
	Level     string    `db:"level"`
	Type      string    `db:"type"`
	SpotsLeft int       `db:"spotsLeft"`
	Place     string    `db:"place"`
}

type UpdateClass struct {
	ID        int       `db:"id"`
	Datetime  time.Time `db:"datetime"`
	Day       string    `db:"day"`
	Level     string    `db:"level"`
	Type      string    `db:"type"`
	SpotsLeft int       `db:"spotsLeft"`
	Place     string    `db:"place"`
}

type Classes interface {
	GetAll() ([]Class, error)
}

type Practitioners interface {
	Insert(ctx context.Context, classID int, name, lastName, email string) error
}
