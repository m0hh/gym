package data

import (
	"database/sql"
	"errors"
)

var (
	ErrRcordNotFound   = errors.New("record not found")
	ErrEditConflict    = errors.New("edit conflict")
	ErrUniqueFood      = errors.New("unique food with serving failed")
	ErrWrongForeignKey = errors.New("wrong Foreign key")
)

type Models struct {
	Tokens TokenModel
	Users  UserModel
	Meals  MealsModel
	Plans  PlanModel
}

func NewModels(db *sql.DB) Models {
	return Models{
		Tokens: TokenModel{DB: db},
		Users:  UserModel{DB: db},
		Meals:  MealsModel{DB: db},
		Plans:  PlanModel{DB: db},
	}
}
