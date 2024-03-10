package data

import (
	"database/sql"
	"errors"
)

var (
	ErrRcordNotFound = errors.New("record not found")
	ErrEditConflict  = errors.New("edit conflict")
	ErrUniqueFood    = errors.New("unique food with serving failed")
)

type Models struct {
	Tokens TokenModel
	Users  UserModel
	Meals  MealsModel
}

func NewModels(db *sql.DB) Models {
	return Models{
		Tokens: TokenModel{DB: db},
		Users:  UserModel{DB: db},
		Meals:  MealsModel{DB: db},
	}
}
