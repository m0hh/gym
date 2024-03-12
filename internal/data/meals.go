package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/m0hh/smart-logitics/internal/validator"
)

type Food struct {
	Id       int64  `json:"id"`
	FoodName string `json:"food_name"`
	Serving  string `json:"serving"`
	Calories int    `json:"calories"`
}

type Breakfast struct {
	Id       int64  `json:"id"`
	Calories int    `json:"calories"`
	Food     []Food `json:"food"`
}

type AmSnack struct {
	Id       int64  `json:"id"`
	Calories int    `json:"calories"`
	Food     []Food `json:"food"`
}

type Lunch struct {
	Id       int64  `json:"id"`
	Calories int    `json:"calories"`
	Food     []Food `json:"food"`
}

type PmSnack struct {
	Id       int64  `json:"id"`
	Calories int    `json:"calories"`
	Food     []Food `json:"food"`
}

type Dinner struct {
	Id       int64  `json:"id"`
	Calories int    `json:"calories"`
	Food     []Food `json:"food"`
}

type MealsModel struct {
	DB *sql.DB
}

func ValidateFood(v *validator.Validator, food Food) {
	v.Check(food.FoodName != "", "name", "food name cannot be empty")
	v.Check(food.Serving != "", "serving", "food serving cannot be empty")
	v.Check(food.Calories > 0, "calories", "food calories cannot be empty")

}

func (m MealsModel) CreateFood(food *Food) error {
	stmt := ` INSERT INTO food (food_name, serving, calories)
	VALUES ($1, $2, $3) 
	RETURNING id
	`

	args := []interface{}{food.FoodName, food.Serving, food.Calories}

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
	SET food_name = $1, serving = $2, calories = $3 
	WHERE id = $4
	RETURNING food_name, serving, calories`

	context, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	defer cancel()

	args := []interface{}{food.FoodName, food.Serving, food.Calories, food.Id}

	err := m.DB.QueryRowContext(context, stmt, args...).Scan(&food.FoodName, &food.Serving, &food.Calories)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrRcordNotFound
		case err.Error() == `pq: duplicate key value violates unique constraint "unique_food_serving"`:
			return ErrUniqueFood
		default:
			return err
		}
	}
	return nil
}

func (m MealsModel) GetById(food *Food) error {
	stmt := `SELECT food_name, serving, calories FROM food
	WHERE id = $1`

	context, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	defer cancel()

	err := m.DB.QueryRowContext(context, stmt, food.Id).Scan(&food.FoodName, &food.Serving, &food.Calories)

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

func (m MealsModel) GetAllFood(food_name string, serving string, filter Filters) ([]*Food, Metadata, error) {
	query := `
	SELECT count(*) OVER(),id, food_name, serving, calories
	FROM food
	WHERE (to_tsvector('simple', food_name) @@ plainto_tsquery('simple', $1) OR $1 = '')
	AND (to_tsvector('simple', serving) @@ plainto_tsquery('simple', $2) OR $2 = '')
	LIMIT $3 OFFSET $4`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, food_name, serving, filter.limit(), filter.offset())

	if err != nil {
		return nil, Metadata{}, err
	}

	defer rows.Close()

	totalRecords := 0
	foods := []*Food{}

	for rows.Next() {
		var food Food
		err := rows.Scan(
			&totalRecords,
			&food.Id,
			&food.FoodName,
			&food.Serving,
			&food.Calories,
		)
		if err != nil {
			return nil, Metadata{}, err
		}

		foods = append(foods, &food)
	}

	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}
	metadata := calculateMetadata(totalRecords, filter.Page, filter.PageSize)
	return foods, metadata, nil
}

func ValidateBreakfast(v *validator.Validator, breakfast Breakfast) {
	v.Check(len(breakfast.Food) > 0, "food", "must send more than 0 foods")
	var ids []int
	for _, food := range breakfast.Food {
		ValidateFood(v, food)
		ids = append(ids, int(food.Id))
	}
	v.Check(validator.Unique(ids), "food", "You munst send not send the same food twice")
}

func (m MealsModel) CreateBreakfast(breakfast *Breakfast) error {
	calories := 0
	for i := range breakfast.Food {
		calories += breakfast.Food[i].Calories
	}
	breakfast.Calories = calories
	tx, err := m.DB.Begin()
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()
	stmt := `INSERT INTO breakfast (calories) VALUES ($1) RETURNING id`

	context, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err = tx.QueryRowContext(context, stmt, breakfast.Calories).Scan(&breakfast.Id)

	if err != nil {
		return err
	}

	var bulkInsertValues []interface{}
	bulkInsertStrings := make([]string, 0)
	i := 1
	for _, food := range breakfast.Food {
		bulkInsertStrings = append(bulkInsertStrings, fmt.Sprintf("($%d,$%d)", i, i+1))
		bulkInsertValues = append(bulkInsertValues, breakfast.Id, food.Id)
		i += 2
	}

	stmt1 := fmt.Sprintf(`INSERT INTO breakfast_food (breakfast_id, food_id) VALUES %s`, strings.Join(bulkInsertStrings, ","))
	_, err = tx.ExecContext(context, stmt1, bulkInsertValues...)
	if err != nil {
		if err.Error() == `pq: insert or update on table "breakfast_food" violates foreign key constraint "breakfast_food_food_id_fkey"` {
			return ErrWrongForeignKey
		}
		return err
	}
	tx.Commit()

	return nil
}

func (m *MealsModel) UpdateBreakfast(breakfast *Breakfast) error {
	calories := 0
	for i := range breakfast.Food {
		calories += breakfast.Food[i].Calories
	}
	breakfast.Calories = calories

	tx, err := m.DB.Begin()
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	context, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt2 := `UPDATE breakfast SET calories = $1 WHERE id = $2`

	_, err = tx.ExecContext(context, stmt2, breakfast.Calories, breakfast.Id)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrRcordNotFound
		default:
			return err
		}
	}

	stmt := `DELETE FROM breakfast_food WHERE breakfast_id = $1`

	_, err = tx.ExecContext(context, stmt, breakfast.Id)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrRcordNotFound
		case err.Error() == `pq: insert or update on table "breakfast_food" violates foreign key constraint "breakfast_food_food_id_fkey"`:
			return ErrWrongForeignKey

		default:
			return err
		}
	}

	var bulkInsertValues []interface{}
	bulkInsertStrings := make([]string, 0)
	i := 1
	for _, food := range breakfast.Food {
		bulkInsertStrings = append(bulkInsertStrings, fmt.Sprintf("($%d,$%d)", i, i+1))
		bulkInsertValues = append(bulkInsertValues, breakfast.Id, food.Id)
		i += 2
	}

	stmt1 := fmt.Sprintf(`INSERT INTO breakfast_food (breakfast_id, food_id) VALUES %s`, strings.Join(bulkInsertStrings, ","))
	_, err = tx.ExecContext(context, stmt1, bulkInsertValues...)
	if err != nil {
		return err
	}
	tx.Commit()

	return nil
}
