package data

import (
	"database/sql"
	"errors"
)

var (
	ErrRcordNotFound = errors.New("record not found")
	ErrEditConflict  = errors.New("edit conflict")
)

type Models struct {
}

func NewModels(db *sql.DB) Models {
	return Models{}
}
