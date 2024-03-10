package data

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/m0hh/smart-logitics/internal/validator"
)

type Food struct {
	Id       int64  `json:"id"`
	FoodName string `json:"food_name"`
	Serving  string `json:"serving"`
}

type Breakfast struct {
	Id       int64 `json:"id"`
	Calories int
	Food     Food `json:"food"`
}

type AmSnack struct {
	Id       int64 `json:"id"`
	Calories int
	Food     Food `json:"food"`
}

type Lunch struct {
	Id       int64 `json:"id"`
	Calories int
	Food     Food `json:"food"`
}

type PmSnack struct {
	Id       int64 `json:"id"`
	Calories int
	Food     Food `json:"food"`
}

type Dinner struct {
	Id       int64 `json:"id"`
	Calories int
	Food     Food `json:"food"`
}

type MealsModel struct {
	DB *sql.DB
}

func ValidateFood(v *validator.Validator, food Food) {
	v.Check(food.FoodName != "", "food", "food name cannot be empty")
	v.Check(food.Serving != "", "food", "food serving cannot be empty")
}

func (m MealsModel) CreateFood(food *Food) error {
	stmt := ` INSERT INTO food (food_name, serving)
	VALUES ($1, $2) 
	RETURNING id
	`

	args := []interface{}{food.FoodName, food.Serving}

	context, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	defer cancel()

	err := m.DB.QueryRowContext(context, stmt, args...).Scan(&food.Id)
	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "unique_food_serving"`:
			return ErrUniqueFood
		default:
			return err
		}
	}
	return nil
}

func (m MealsModel) UpdateFood(food *Food) error {
	stmt := `UPDATE food 
	SET food_name = $1, serving = $2
	WHERE id = $3
	RETURNING food_name, serving`

	context, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	defer cancel()

	args := []interface{}{food.FoodName, food.Serving, food.Id}

	err := m.DB.QueryRowContext(context, stmt, args...).Scan(&food.FoodName, &food.Serving)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrRcordNotFound
		default:
			return err
		}
	}
	return nil
}

func (m MealsModel) GetById(food *Food) error {
	stmt := `SELECT food_name, serving FROM food
	WHERE id = $1`

	context, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	defer cancel()

	err := m.DB.QueryRowContext(context, stmt, food.Id).Scan(&food.FoodName, &food.Serving)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrRcordNotFound
		default:
			return err
		}
	}
	return nil
}
