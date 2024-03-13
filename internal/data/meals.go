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

func (m MealsModel) DeleteFood(id int64) error {
	if id < 1 {
		return ErrRcordNotFound
	}

	tx, err := m.DB.Begin()
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	query1 := `
        DELETE FROM breakfast_food 
        WHERE food_id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err = tx.ExecContext(ctx, query1, id)

	if err != nil {
		return err
	}

	query := `
        DELETE FROM food
        WHERE id = $1`

	result, err := tx.ExecContext(ctx, query, id)

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrRcordNotFound
	}

	tx.Commit()

	return nil
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
		switch {
		case err.Error() == `pq: insert or update on table "breakfast_food" violates foreign key constraint "breakfast_food_food_id_fkey"`:
			return ErrWrongForeignKey
		default:
			return err

		}
	}
	tx.Commit()

	return nil
}

func (m *MealsModel) GetAllBreakfastFoodsId(breakfast *Breakfast) error {
	stmt := `SELECT breakfast.calories,food.id,food_name,food.serving,food.calories FROM breakfast INNER JOIN breakfast_food ON breakfast.id = breakfast_food.breakfast_id
	 INNER JOIN food ON breakfast_food.food_id = food.id
	 WHERE breakfast.id = $1
	 `

	context, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	defer cancel()

	rows, err := m.DB.QueryContext(context, stmt, breakfast.Id)

	if err != nil {
		return err
	}

	defer rows.Close()

	for rows.Next() {
		food := Food{}
		err = rows.Scan(
			&breakfast.Calories,
			&food.Id,
			&food.FoodName,
			&food.Serving,
			&food.Calories,
		)

		if err != nil {
			return err
		}
		breakfast.Food = append(breakfast.Food, food)
	}

	if err = rows.Err(); err != nil {
		return err
	}

	if breakfast.Calories == 0 {
		return ErrRcordNotFound
	}

	return nil
}

func (m *MealsModel) GetAllBreakfastFoods(filter Filters) ([]*Breakfast, Metadata, error) {
	stmt := `SELECT count(*) OVER(), breakfast.id, breakfast.calories,food.id,food_name,food.serving,food.calories FROM breakfast INNER JOIN breakfast_food ON breakfast.id = breakfast_food.breakfast_id
	 INNER JOIN food ON breakfast_food.food_id = food.id
	 LIMIT $1 OFFSET $2
	 `

	context, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	defer cancel()

	rows, err := m.DB.QueryContext(context, stmt, filter.limit(), filter.offset())

	if err != nil {
		return nil, Metadata{}, err
	}

	defer rows.Close()

	ids_to_breakfast := make(map[int]*Breakfast)
	var breakfasts []*Breakfast
	totalRecords := 0

	for rows.Next() {

		var breakfast_id int
		var breakfast_calories int
		food := Food{}
		err = rows.Scan(
			&totalRecords,
			&breakfast_id,
			&breakfast_calories,
			&food.Id,
			&food.FoodName,
			&food.Serving,
			&food.Calories,
		)

		if err != nil {
			return nil, Metadata{}, err
		}
		if _, ok := ids_to_breakfast[breakfast_id]; !ok {

			breakfast := Breakfast{Id: int64(breakfast_id), Calories: breakfast_calories}
			ids_to_breakfast[breakfast_id] = &breakfast

			ids_to_breakfast[breakfast_id].Food = append(ids_to_breakfast[breakfast_id].Food, food)
			breakfasts = append(breakfasts, ids_to_breakfast[breakfast_id])
		} else {
			ids_to_breakfast[breakfast_id].Food = append(ids_to_breakfast[breakfast_id].Food, food)
		}

	}

	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetadata(totalRecords, filter.Page, filter.PageSize)
	return breakfasts, metadata, nil

}

func (m MealsModel) DeleteBreakfast(id int64) error {
	if id < 1 {
		return ErrRcordNotFound
	}

	tx, err := m.DB.Begin()
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	query := `
        DELETE FROM breakfast_food 
        WHERE breakfast_id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err = tx.ExecContext(ctx, query, id)

	if err != nil {
		return err
	}

	query1 := `
        DELETE FROM breakfast
        WHERE id = $1`

	result, err := tx.ExecContext(ctx, query1, id)

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrRcordNotFound
	}
	tx.Commit()

	return nil
}

/////////////////////////////////////////////////////////////////////////////////

func ValidateAmSnack(v *validator.Validator, amSnack AmSnack) {
	v.Check(len(amSnack.Food) > 0, "food", "must send more than 0 foods")
	var ids []int
	for _, food := range amSnack.Food {
		ValidateFood(v, food)
		ids = append(ids, int(food.Id))
	}
	v.Check(validator.Unique(ids), "food", "You must not send the same food twice")
}

func (m MealsModel) CreateAmSnack(amSnack *AmSnack) error {
	calories := 0
	for i := range amSnack.Food {
		calories += amSnack.Food[i].Calories
	}
	amSnack.Calories = calories
	tx, err := m.DB.Begin()
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()
	stmt := `INSERT INTO am_snack (calories) VALUES ($1) RETURNING id`

	context, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err = tx.QueryRowContext(context, stmt, amSnack.Calories).Scan(&amSnack.Id)

	if err != nil {
		return err
	}

	var bulkInsertValues []interface{}
	bulkInsertStrings := make([]string, 0)
	i := 1
	for _, food := range amSnack.Food {
		bulkInsertStrings = append(bulkInsertStrings, fmt.Sprintf("($%d,$%d)", i, i+1))
		bulkInsertValues = append(bulkInsertValues, amSnack.Id, food.Id)
		i += 2
	}

	stmt1 := fmt.Sprintf(`INSERT INTO am_snack_food (am_snack_id, food_id) VALUES %s`, strings.Join(bulkInsertStrings, ","))
	_, err = tx.ExecContext(context, stmt1, bulkInsertValues...)
	fmt.Println(err.Error() == `pq: insert or update on table "am_snack_food" violates foreign key constraint "am_snack_food_food_id_fkey"`)
	if err != nil {
		if err.Error() == `pq: insert or update on table "am_snack_food" violates foreign key constraint "am_snack_food_food_id_fkey"` {
			return ErrWrongForeignKey
		}
		return err
	}
	tx.Commit()

	return nil
}

func (m *MealsModel) UpdateAmSnack(am_snack *AmSnack) error {
	calories := 0
	for i := range am_snack.Food {
		calories += am_snack.Food[i].Calories
	}
	am_snack.Calories = calories

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

	stmt2 := `UPDATE am_snack SET calories = $1 WHERE id = $2`

	_, err = tx.ExecContext(context, stmt2, am_snack.Calories, am_snack.Id)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrRcordNotFound
		default:
			return err
		}
	}

	stmt := `DELETE FROM am_snack_food WHERE am_snack_id = $1`

	_, err = tx.ExecContext(context, stmt, am_snack.Id)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrRcordNotFound
		default:
			return err
		}
	}

	var bulkInsertValues []interface{}
	bulkInsertStrings := make([]string, 0)
	i := 1
	for _, food := range am_snack.Food {
		bulkInsertStrings = append(bulkInsertStrings, fmt.Sprintf("($%d,$%d)", i, i+1))
		bulkInsertValues = append(bulkInsertValues, am_snack.Id, food.Id)
		i += 2
	}

	stmt1 := fmt.Sprintf(`INSERT INTO am_snack_food (am_snack_id, food_id) VALUES %s`, strings.Join(bulkInsertStrings, ","))
	_, err = tx.ExecContext(context, stmt1, bulkInsertValues...)
	if err != nil {
		switch {
		case err.Error() == `pq: insert or update on table "am_snack_food" violates foreign key constraint "am_snack_food_food_id_fkey"`:
			return ErrWrongForeignKey
		default:
			return err
		}
	}
	tx.Commit()

	return nil
}

func (m *MealsModel) GetAllAmSnackID(am_snack *AmSnack) error {
	stmt := `SELECT am_snack.calories,food.id,food_name,food.serving,food.calories FROM am_snack INNER JOIN am_snack_food ON am_snack.id = am_snack_food.am_snack_id
	 INNER JOIN food ON am_snack_food.food_id = food.id
	 WHERE am_snack.id = $1
	 `

	context, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	defer cancel()

	rows, err := m.DB.QueryContext(context, stmt, am_snack.Id)

	if err != nil {
		return err
	}

	defer rows.Close()

	for rows.Next() {
		food := Food{}
		err = rows.Scan(
			&am_snack.Calories,
			&food.Id,
			&food.FoodName,
			&food.Serving,
			&food.Calories,
		)

		if err != nil {
			return err
		}
		am_snack.Food = append(am_snack.Food, food)
	}

	if err = rows.Err(); err != nil {
		return err
	}

	if am_snack.Calories == 0 {
		return ErrRcordNotFound
	}

	return nil
}

func (m *MealsModel) GetAllAmSnacks(filter Filters) ([]*AmSnack, Metadata, error) {
	stmt := `SELECT count(*) OVER(), am_snack.id, am_snack.calories,food.id,food_name,food.serving,food.calories FROM am_snack INNER JOIN am_snack_food ON am_snack.id = am_snack_food.am_snack_id
	 INNER JOIN food ON am_snack_food.food_id = food.id
	 LIMIT $1 OFFSET $2
	 `

	context, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	defer cancel()

	rows, err := m.DB.QueryContext(context, stmt, filter.limit(), filter.offset())

	if err != nil {
		return nil, Metadata{}, err
	}

	defer rows.Close()

	ids_to_amsnacks := make(map[int]*AmSnack)
	var am_snacks []*AmSnack
	totalRecords := 0

	for rows.Next() {

		var am_snack_id int
		var am_snack_calories int
		food := Food{}
		err = rows.Scan(
			&totalRecords,
			&am_snack_id,
			&am_snack_calories,
			&food.Id,
			&food.FoodName,
			&food.Serving,
			&food.Calories,
		)

		if err != nil {
			return nil, Metadata{}, err
		}
		if _, ok := ids_to_amsnacks[am_snack_id]; !ok {

			am_snack := AmSnack{Id: int64(am_snack_id), Calories: am_snack_calories}
			ids_to_amsnacks[am_snack_id] = &am_snack

			ids_to_amsnacks[am_snack_id].Food = append(ids_to_amsnacks[am_snack_id].Food, food)
			am_snacks = append(am_snacks, ids_to_amsnacks[am_snack_id])
		} else {
			ids_to_amsnacks[am_snack_id].Food = append(ids_to_amsnacks[am_snack_id].Food, food)
		}

	}

	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetadata(totalRecords, filter.Page, filter.PageSize)
	return am_snacks, metadata, nil

}

func (m MealsModel) DeleteAmSnack(id int64) error {
	if id < 1 {
		return ErrRcordNotFound
	}

	tx, err := m.DB.Begin()
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	query := `
        DELETE FROM am_snack_food 
        WHERE am_snack_id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err = tx.ExecContext(ctx, query, id)

	if err != nil {
		return err
	}

	query1 := `
        DELETE FROM am_snack
        WHERE id = $1`

	result, err := tx.ExecContext(ctx, query1, id)

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrRcordNotFound
	}
	tx.Commit()

	return nil
}

///////////////////////////////////////////////////////////////////////////////////

// func ValidateBreakfast(v *validator.Validator, breakfast Breakfast) {
// 	v.Check(len(breakfast.Food) > 0, "food", "must send more than 0 foods")
// 	var ids []int
// 	for _, food := range breakfast.Food {
// 		ValidateFood(v, food)
// 		ids = append(ids, int(food.Id))
// 	}
// 	v.Check(validator.Unique(ids), "food", "You munst send not send the same food twice")
// }

// func (m MealsModel) CreateBreakfast(breakfast *Breakfast) error {
// 	calories := 0
// 	for i := range breakfast.Food {
// 		calories += breakfast.Food[i].Calories
// 	}
// 	breakfast.Calories = calories
// 	tx, err := m.DB.Begin()
// 	if err != nil {
// 		return err
// 	}

// 	defer func() {
// 		if err != nil {
// 			tx.Rollback()
// 		}
// 	}()
// 	stmt := `INSERT INTO breakfast (calories) VALUES ($1) RETURNING id`

// 	context, cancel := context.WithTimeout(context.Background(), 3*time.Second)
// 	defer cancel()

// 	err = tx.QueryRowContext(context, stmt, breakfast.Calories).Scan(&breakfast.Id)

// 	if err != nil {
// 		return err
// 	}

// 	var bulkInsertValues []interface{}
// 	bulkInsertStrings := make([]string, 0)
// 	i := 1
// 	for _, food := range breakfast.Food {
// 		bulkInsertStrings = append(bulkInsertStrings, fmt.Sprintf("($%d,$%d)", i, i+1))
// 		bulkInsertValues = append(bulkInsertValues, breakfast.Id, food.Id)
// 		i += 2
// 	}

// 	stmt1 := fmt.Sprintf(`INSERT INTO breakfast_food (breakfast_id, food_id) VALUES %s`, strings.Join(bulkInsertStrings, ","))
// 	_, err = tx.ExecContext(context, stmt1, bulkInsertValues...)
// 	if err != nil {
// 		if err.Error() == `pq: insert or update on table "breakfast_food" violates foreign key constraint "breakfast_food_food_id_fkey"` {
// 			return ErrWrongForeignKey
// 		}
// 		return err
// 	}
// 	tx.Commit()

// 	return nil
// }

// func (m *MealsModel) UpdateBreakfast(breakfast *Breakfast) error {
// 	calories := 0
// 	for i := range breakfast.Food {
// 		calories += breakfast.Food[i].Calories
// 	}
// 	breakfast.Calories = calories

// 	tx, err := m.DB.Begin()
// 	if err != nil {
// 		return err
// 	}

// 	defer func() {
// 		if err != nil {
// 			tx.Rollback()
// 		}
// 	}()

// 	context, cancel := context.WithTimeout(context.Background(), 3*time.Second)
// 	defer cancel()

// 	stmt2 := `UPDATE breakfast SET calories = $1 WHERE id = $2`

// 	_, err = tx.ExecContext(context, stmt2, breakfast.Calories, breakfast.Id)

// 	if err != nil {
// 		switch {
// 		case errors.Is(err, sql.ErrNoRows):
// 			return ErrRcordNotFound
// 		default:
// 			return err
// 		}
// 	}

// 	stmt := `DELETE FROM breakfast_food WHERE breakfast_id = $1`

// 	_, err = tx.ExecContext(context, stmt, breakfast.Id)

// 	if err != nil {
// 		switch {
// 		case errors.Is(err, sql.ErrNoRows):
// 			return ErrRcordNotFound
// 		case err.Error() == `pq: insert or update on table "breakfast_food" violates foreign key constraint "breakfast_food_food_id_fkey"`:
// 			return ErrWrongForeignKey

// 		default:
// 			return err
// 		}
// 	}

// 	var bulkInsertValues []interface{}
// 	bulkInsertStrings := make([]string, 0)
// 	i := 1
// 	for _, food := range breakfast.Food {
// 		bulkInsertStrings = append(bulkInsertStrings, fmt.Sprintf("($%d,$%d)", i, i+1))
// 		bulkInsertValues = append(bulkInsertValues, breakfast.Id, food.Id)
// 		i += 2
// 	}

// 	stmt1 := fmt.Sprintf(`INSERT INTO breakfast_food (breakfast_id, food_id) VALUES %s`, strings.Join(bulkInsertStrings, ","))
// 	_, err = tx.ExecContext(context, stmt1, bulkInsertValues...)
// 	if err != nil {
// 		return err
// 	}
// 	tx.Commit()

// 	return nil
// }

// func (m *MealsModel) GetAllBreakfastFoodsId(breakfast *Breakfast) error {
// 	stmt := `SELECT breakfast.calories,food.id,food_name,food.serving,food.calories FROM breakfast INNER JOIN breakfast_food ON breakfast.id = breakfast_food.breakfast_id
// 	 INNER JOIN food ON breakfast_food.food_id = food.id
// 	 WHERE breakfast.id = $1
// 	 `

// 	context, cancel := context.WithTimeout(context.Background(), 3*time.Second)

// 	defer cancel()

// 	rows, err := m.DB.QueryContext(context, stmt, breakfast.Id)

// 	if err != nil {
// 		return err
// 	}

// 	defer rows.Close()

// 	for rows.Next() {
// 		food := Food{}
// 		err = rows.Scan(
// 			&breakfast.Calories,
// 			&food.Id,
// 			&food.FoodName,
// 			&food.Serving,
// 			&food.Calories,
// 		)

// 		if err != nil {
// 			return err
// 		}
// 		breakfast.Food = append(breakfast.Food, food)
// 	}

// 	if err = rows.Err(); err != nil {
// 		return err
// 	}

// 	if breakfast.Calories == 0 {
// 		return ErrRcordNotFound
// 	}

// 	return nil
// }

// func (m *MealsModel) GetAllBreakfastFoods(filter Filters) ([]*Breakfast, Metadata, error) {
// 	stmt := `SELECT count(*) OVER(), breakfast.id, breakfast.calories,food.id,food_name,food.serving,food.calories FROM breakfast INNER JOIN breakfast_food ON breakfast.id = breakfast_food.breakfast_id
// 	 INNER JOIN food ON breakfast_food.food_id = food.id
// 	 LIMIT $1 OFFSET $2
// 	 `

// 	context, cancel := context.WithTimeout(context.Background(), 3*time.Second)

// 	defer cancel()

// 	rows, err := m.DB.QueryContext(context, stmt, filter.limit(), filter.offset())

// 	if err != nil {
// 		return nil, Metadata{}, err
// 	}

// 	defer rows.Close()

// 	ids_to_breakfast := make(map[int]*Breakfast)
// 	var breakfasts []*Breakfast
// 	totalRecords := 0

// 	for rows.Next() {

// 		var breakfast_id int
// 		var breakfast_calories int
// 		food := Food{}
// 		err = rows.Scan(
// 			&totalRecords,
// 			&breakfast_id,
// 			&breakfast_calories,
// 			&food.Id,
// 			&food.FoodName,
// 			&food.Serving,
// 			&food.Calories,
// 		)

// 		if err != nil {
// 			return nil, Metadata{}, err
// 		}
// 		if _, ok := ids_to_breakfast[breakfast_id]; !ok {

// 			breakfast := Breakfast{Id: int64(breakfast_id), Calories: breakfast_calories}
// 			ids_to_breakfast[breakfast_id] = &breakfast

// 			ids_to_breakfast[breakfast_id].Food = append(ids_to_breakfast[breakfast_id].Food, food)
// 			breakfasts = append(breakfasts, ids_to_breakfast[breakfast_id])
// 		} else {
// 			ids_to_breakfast[breakfast_id].Food = append(ids_to_breakfast[breakfast_id].Food, food)
// 		}

// 	}

// 	if err = rows.Err(); err != nil {
// 		return nil, Metadata{}, err
// 	}

// 	metadata := calculateMetadata(totalRecords, filter.Page, filter.PageSize)
// 	return breakfasts, metadata, nil

// }

// func (m MealsModel) DeleteBreakfast(id int64) error {
// 	if id < 1 {
// 		return ErrRcordNotFound
// 	}

// 	tx, err := m.DB.Begin()
// 	if err != nil {
// 		return err
// 	}

// 	defer func() {
// 		if err != nil {
// 			tx.Rollback()
// 		}
// 	}()

// 	query := `
//         DELETE FROM breakfast_food
//         WHERE breakfast_id = $1`

// 	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
// 	defer cancel()

// 	_, err = tx.ExecContext(ctx, query, id)

// 	if err != nil {
// 		return err
// 	}

// 	query1 := `
//         DELETE FROM breakfast
//         WHERE id = $1`

// 	result, err := tx.ExecContext(ctx, query1, id)

// 	if err != nil {
// 		return err
// 	}

// 	rowsAffected, err := result.RowsAffected()
// 	if err != nil {
// 		return err
// 	}

// 	if rowsAffected == 0 {
// 		return ErrRcordNotFound
// 	}
// 	tx.Commit()

// 	return nil
// }

// ////////////////////////////////////////////////////////////////////////////////////

// func ValidateBreakfast(v *validator.Validator, breakfast Breakfast) {
// 	v.Check(len(breakfast.Food) > 0, "food", "must send more than 0 foods")
// 	var ids []int
// 	for _, food := range breakfast.Food {
// 		ValidateFood(v, food)
// 		ids = append(ids, int(food.Id))
// 	}
// 	v.Check(validator.Unique(ids), "food", "You munst send not send the same food twice")
// }

// func (m MealsModel) CreateBreakfast(breakfast *Breakfast) error {
// 	calories := 0
// 	for i := range breakfast.Food {
// 		calories += breakfast.Food[i].Calories
// 	}
// 	breakfast.Calories = calories
// 	tx, err := m.DB.Begin()
// 	if err != nil {
// 		return err
// 	}

// 	defer func() {
// 		if err != nil {
// 			tx.Rollback()
// 		}
// 	}()
// 	stmt := `INSERT INTO breakfast (calories) VALUES ($1) RETURNING id`

// 	context, cancel := context.WithTimeout(context.Background(), 3*time.Second)
// 	defer cancel()

// 	err = tx.QueryRowContext(context, stmt, breakfast.Calories).Scan(&breakfast.Id)

// 	if err != nil {
// 		return err
// 	}

// 	var bulkInsertValues []interface{}
// 	bulkInsertStrings := make([]string, 0)
// 	i := 1
// 	for _, food := range breakfast.Food {
// 		bulkInsertStrings = append(bulkInsertStrings, fmt.Sprintf("($%d,$%d)", i, i+1))
// 		bulkInsertValues = append(bulkInsertValues, breakfast.Id, food.Id)
// 		i += 2
// 	}

// 	stmt1 := fmt.Sprintf(`INSERT INTO breakfast_food (breakfast_id, food_id) VALUES %s`, strings.Join(bulkInsertStrings, ","))
// 	_, err = tx.ExecContext(context, stmt1, bulkInsertValues...)
// 	if err != nil {
// 		if err.Error() == `pq: insert or update on table "breakfast_food" violates foreign key constraint "breakfast_food_food_id_fkey"` {
// 			return ErrWrongForeignKey
// 		}
// 		return err
// 	}
// 	tx.Commit()

// 	return nil
// }

// func (m *MealsModel) UpdateBreakfast(breakfast *Breakfast) error {
// 	calories := 0
// 	for i := range breakfast.Food {
// 		calories += breakfast.Food[i].Calories
// 	}
// 	breakfast.Calories = calories

// 	tx, err := m.DB.Begin()
// 	if err != nil {
// 		return err
// 	}

// 	defer func() {
// 		if err != nil {
// 			tx.Rollback()
// 		}
// 	}()

// 	context, cancel := context.WithTimeout(context.Background(), 3*time.Second)
// 	defer cancel()

// 	stmt2 := `UPDATE breakfast SET calories = $1 WHERE id = $2`

// 	_, err = tx.ExecContext(context, stmt2, breakfast.Calories, breakfast.Id)

// 	if err != nil {
// 		switch {
// 		case errors.Is(err, sql.ErrNoRows):
// 			return ErrRcordNotFound
// 		default:
// 			return err
// 		}
// 	}

// 	stmt := `DELETE FROM breakfast_food WHERE breakfast_id = $1`

// 	_, err = tx.ExecContext(context, stmt, breakfast.Id)

// 	if err != nil {
// 		switch {
// 		case errors.Is(err, sql.ErrNoRows):
// 			return ErrRcordNotFound
// 		case err.Error() == `pq: insert or update on table "breakfast_food" violates foreign key constraint "breakfast_food_food_id_fkey"`:
// 			return ErrWrongForeignKey

// 		default:
// 			return err
// 		}
// 	}

// 	var bulkInsertValues []interface{}
// 	bulkInsertStrings := make([]string, 0)
// 	i := 1
// 	for _, food := range breakfast.Food {
// 		bulkInsertStrings = append(bulkInsertStrings, fmt.Sprintf("($%d,$%d)", i, i+1))
// 		bulkInsertValues = append(bulkInsertValues, breakfast.Id, food.Id)
// 		i += 2
// 	}

// 	stmt1 := fmt.Sprintf(`INSERT INTO breakfast_food (breakfast_id, food_id) VALUES %s`, strings.Join(bulkInsertStrings, ","))
// 	_, err = tx.ExecContext(context, stmt1, bulkInsertValues...)
// 	if err != nil {
// 		return err
// 	}
// 	tx.Commit()

// 	return nil
// }

// func (m *MealsModel) GetAllBreakfastFoodsId(breakfast *Breakfast) error {
// 	stmt := `SELECT breakfast.calories,food.id,food_name,food.serving,food.calories FROM breakfast INNER JOIN breakfast_food ON breakfast.id = breakfast_food.breakfast_id
// 	 INNER JOIN food ON breakfast_food.food_id = food.id
// 	 WHERE breakfast.id = $1
// 	 `

// 	context, cancel := context.WithTimeout(context.Background(), 3*time.Second)

// 	defer cancel()

// 	rows, err := m.DB.QueryContext(context, stmt, breakfast.Id)

// 	if err != nil {
// 		return err
// 	}

// 	defer rows.Close()

// 	for rows.Next() {
// 		food := Food{}
// 		err = rows.Scan(
// 			&breakfast.Calories,
// 			&food.Id,
// 			&food.FoodName,
// 			&food.Serving,
// 			&food.Calories,
// 		)

// 		if err != nil {
// 			return err
// 		}
// 		breakfast.Food = append(breakfast.Food, food)
// 	}

// 	if err = rows.Err(); err != nil {
// 		return err
// 	}

// 	if breakfast.Calories == 0 {
// 		return ErrRcordNotFound
// 	}

// 	return nil
// }

// func (m *MealsModel) GetAllBreakfastFoods(filter Filters) ([]*Breakfast, Metadata, error) {
// 	stmt := `SELECT count(*) OVER(), breakfast.id, breakfast.calories,food.id,food_name,food.serving,food.calories FROM breakfast INNER JOIN breakfast_food ON breakfast.id = breakfast_food.breakfast_id
// 	 INNER JOIN food ON breakfast_food.food_id = food.id
// 	 LIMIT $1 OFFSET $2
// 	 `

// 	context, cancel := context.WithTimeout(context.Background(), 3*time.Second)

// 	defer cancel()

// 	rows, err := m.DB.QueryContext(context, stmt, filter.limit(), filter.offset())

// 	if err != nil {
// 		return nil, Metadata{}, err
// 	}

// 	defer rows.Close()

// 	ids_to_breakfast := make(map[int]*Breakfast)
// 	var breakfasts []*Breakfast
// 	totalRecords := 0

// 	for rows.Next() {

// 		var breakfast_id int
// 		var breakfast_calories int
// 		food := Food{}
// 		err = rows.Scan(
// 			&totalRecords,
// 			&breakfast_id,
// 			&breakfast_calories,
// 			&food.Id,
// 			&food.FoodName,
// 			&food.Serving,
// 			&food.Calories,
// 		)

// 		if err != nil {
// 			return nil, Metadata{}, err
// 		}
// 		if _, ok := ids_to_breakfast[breakfast_id]; !ok {

// 			breakfast := Breakfast{Id: int64(breakfast_id), Calories: breakfast_calories}
// 			ids_to_breakfast[breakfast_id] = &breakfast

// 			ids_to_breakfast[breakfast_id].Food = append(ids_to_breakfast[breakfast_id].Food, food)
// 			breakfasts = append(breakfasts, ids_to_breakfast[breakfast_id])
// 		} else {
// 			ids_to_breakfast[breakfast_id].Food = append(ids_to_breakfast[breakfast_id].Food, food)
// 		}

// 	}

// 	if err = rows.Err(); err != nil {
// 		return nil, Metadata{}, err
// 	}

// 	metadata := calculateMetadata(totalRecords, filter.Page, filter.PageSize)
// 	return breakfasts, metadata, nil

// }

// func (m MealsModel) DeleteBreakfast(id int64) error {
// 	if id < 1 {
// 		return ErrRcordNotFound
// 	}

// 	tx, err := m.DB.Begin()
// 	if err != nil {
// 		return err
// 	}

// 	defer func() {
// 		if err != nil {
// 			tx.Rollback()
// 		}
// 	}()

// 	query := `
//         DELETE FROM breakfast_food
//         WHERE breakfast_id = $1`

// 	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
// 	defer cancel()

// 	_, err = tx.ExecContext(ctx, query, id)

// 	if err != nil {
// 		return err
// 	}

// 	query1 := `
//         DELETE FROM breakfast
//         WHERE id = $1`

// 	result, err := tx.ExecContext(ctx, query1, id)

// 	if err != nil {
// 		return err
// 	}

// 	rowsAffected, err := result.RowsAffected()
// 	if err != nil {
// 		return err
// 	}

// 	if rowsAffected == 0 {
// 		return ErrRcordNotFound
// 	}
// 	tx.Commit()

// 	return nil
// }

// ////////////////////////////////////////////////////////////////////////////

// func ValidateBreakfast(v *validator.Validator, breakfast Breakfast) {
// 	v.Check(len(breakfast.Food) > 0, "food", "must send more than 0 foods")
// 	var ids []int
// 	for _, food := range breakfast.Food {
// 		ValidateFood(v, food)
// 		ids = append(ids, int(food.Id))
// 	}
// 	v.Check(validator.Unique(ids), "food", "You munst send not send the same food twice")
// }

// func (m MealsModel) CreateBreakfast(breakfast *Breakfast) error {
// 	calories := 0
// 	for i := range breakfast.Food {
// 		calories += breakfast.Food[i].Calories
// 	}
// 	breakfast.Calories = calories
// 	tx, err := m.DB.Begin()
// 	if err != nil {
// 		return err
// 	}

// 	defer func() {
// 		if err != nil {
// 			tx.Rollback()
// 		}
// 	}()
// 	stmt := `INSERT INTO breakfast (calories) VALUES ($1) RETURNING id`

// 	context, cancel := context.WithTimeout(context.Background(), 3*time.Second)
// 	defer cancel()

// 	err = tx.QueryRowContext(context, stmt, breakfast.Calories).Scan(&breakfast.Id)

// 	if err != nil {
// 		return err
// 	}

// 	var bulkInsertValues []interface{}
// 	bulkInsertStrings := make([]string, 0)
// 	i := 1
// 	for _, food := range breakfast.Food {
// 		bulkInsertStrings = append(bulkInsertStrings, fmt.Sprintf("($%d,$%d)", i, i+1))
// 		bulkInsertValues = append(bulkInsertValues, breakfast.Id, food.Id)
// 		i += 2
// 	}

// 	stmt1 := fmt.Sprintf(`INSERT INTO breakfast_food (breakfast_id, food_id) VALUES %s`, strings.Join(bulkInsertStrings, ","))
// 	_, err = tx.ExecContext(context, stmt1, bulkInsertValues...)
// 	if err != nil {
// 		if err.Error() == `pq: insert or update on table "breakfast_food" violates foreign key constraint "breakfast_food_food_id_fkey"` {
// 			return ErrWrongForeignKey
// 		}
// 		return err
// 	}
// 	tx.Commit()

// 	return nil
// }

// func (m *MealsModel) UpdateBreakfast(breakfast *Breakfast) error {
// 	calories := 0
// 	for i := range breakfast.Food {
// 		calories += breakfast.Food[i].Calories
// 	}
// 	breakfast.Calories = calories

// 	tx, err := m.DB.Begin()
// 	if err != nil {
// 		return err
// 	}

// 	defer func() {
// 		if err != nil {
// 			tx.Rollback()
// 		}
// 	}()

// 	context, cancel := context.WithTimeout(context.Background(), 3*time.Second)
// 	defer cancel()

// 	stmt2 := `UPDATE breakfast SET calories = $1 WHERE id = $2`

// 	_, err = tx.ExecContext(context, stmt2, breakfast.Calories, breakfast.Id)

// 	if err != nil {
// 		switch {
// 		case errors.Is(err, sql.ErrNoRows):
// 			return ErrRcordNotFound
// 		default:
// 			return err
// 		}
// 	}

// 	stmt := `DELETE FROM breakfast_food WHERE breakfast_id = $1`

// 	_, err = tx.ExecContext(context, stmt, breakfast.Id)

// 	if err != nil {
// 		switch {
// 		case errors.Is(err, sql.ErrNoRows):
// 			return ErrRcordNotFound
// 		case err.Error() == `pq: insert or update on table "breakfast_food" violates foreign key constraint "breakfast_food_food_id_fkey"`:
// 			return ErrWrongForeignKey

// 		default:
// 			return err
// 		}
// 	}

// 	var bulkInsertValues []interface{}
// 	bulkInsertStrings := make([]string, 0)
// 	i := 1
// 	for _, food := range breakfast.Food {
// 		bulkInsertStrings = append(bulkInsertStrings, fmt.Sprintf("($%d,$%d)", i, i+1))
// 		bulkInsertValues = append(bulkInsertValues, breakfast.Id, food.Id)
// 		i += 2
// 	}

// 	stmt1 := fmt.Sprintf(`INSERT INTO breakfast_food (breakfast_id, food_id) VALUES %s`, strings.Join(bulkInsertStrings, ","))
// 	_, err = tx.ExecContext(context, stmt1, bulkInsertValues...)
// 	if err != nil {
// 		return err
// 	}
// 	tx.Commit()

// 	return nil
// }

// func (m *MealsModel) GetAllBreakfastFoodsId(breakfast *Breakfast) error {
// 	stmt := `SELECT breakfast.calories,food.id,food_name,food.serving,food.calories FROM breakfast INNER JOIN breakfast_food ON breakfast.id = breakfast_food.breakfast_id
// 	 INNER JOIN food ON breakfast_food.food_id = food.id
// 	 WHERE breakfast.id = $1
// 	 `

// 	context, cancel := context.WithTimeout(context.Background(), 3*time.Second)

// 	defer cancel()

// 	rows, err := m.DB.QueryContext(context, stmt, breakfast.Id)

// 	if err != nil {
// 		return err
// 	}

// 	defer rows.Close()

// 	for rows.Next() {
// 		food := Food{}
// 		err = rows.Scan(
// 			&breakfast.Calories,
// 			&food.Id,
// 			&food.FoodName,
// 			&food.Serving,
// 			&food.Calories,
// 		)

// 		if err != nil {
// 			return err
// 		}
// 		breakfast.Food = append(breakfast.Food, food)
// 	}

// 	if err = rows.Err(); err != nil {
// 		return err
// 	}

// 	if breakfast.Calories == 0 {
// 		return ErrRcordNotFound
// 	}

// 	return nil
// }

// func (m *MealsModel) GetAllBreakfastFoods(filter Filters) ([]*Breakfast, Metadata, error) {
// 	stmt := `SELECT count(*) OVER(), breakfast.id, breakfast.calories,food.id,food_name,food.serving,food.calories FROM breakfast INNER JOIN breakfast_food ON breakfast.id = breakfast_food.breakfast_id
// 	 INNER JOIN food ON breakfast_food.food_id = food.id
// 	 LIMIT $1 OFFSET $2
// 	 `

// 	context, cancel := context.WithTimeout(context.Background(), 3*time.Second)

// 	defer cancel()

// 	rows, err := m.DB.QueryContext(context, stmt, filter.limit(), filter.offset())

// 	if err != nil {
// 		return nil, Metadata{}, err
// 	}

// 	defer rows.Close()

// 	ids_to_breakfast := make(map[int]*Breakfast)
// 	var breakfasts []*Breakfast
// 	totalRecords := 0

// 	for rows.Next() {

// 		var breakfast_id int
// 		var breakfast_calories int
// 		food := Food{}
// 		err = rows.Scan(
// 			&totalRecords,
// 			&breakfast_id,
// 			&breakfast_calories,
// 			&food.Id,
// 			&food.FoodName,
// 			&food.Serving,
// 			&food.Calories,
// 		)

// 		if err != nil {
// 			return nil, Metadata{}, err
// 		}
// 		if _, ok := ids_to_breakfast[breakfast_id]; !ok {

// 			breakfast := Breakfast{Id: int64(breakfast_id), Calories: breakfast_calories}
// 			ids_to_breakfast[breakfast_id] = &breakfast

// 			ids_to_breakfast[breakfast_id].Food = append(ids_to_breakfast[breakfast_id].Food, food)
// 			breakfasts = append(breakfasts, ids_to_breakfast[breakfast_id])
// 		} else {
// 			ids_to_breakfast[breakfast_id].Food = append(ids_to_breakfast[breakfast_id].Food, food)
// 		}

// 	}

// 	if err = rows.Err(); err != nil {
// 		return nil, Metadata{}, err
// 	}

// 	metadata := calculateMetadata(totalRecords, filter.Page, filter.PageSize)
// 	return breakfasts, metadata, nil

// }

// func (m MealsModel) DeleteBreakfast(id int64) error {
// 	if id < 1 {
// 		return ErrRcordNotFound
// 	}

// 	tx, err := m.DB.Begin()
// 	if err != nil {
// 		return err
// 	}

// 	defer func() {
// 		if err != nil {
// 			tx.Rollback()
// 		}
// 	}()

// 	query := `
//         DELETE FROM breakfast_food
//         WHERE breakfast_id = $1`

// 	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
// 	defer cancel()

// 	_, err = tx.ExecContext(ctx, query, id)

// 	if err != nil {
// 		return err
// 	}

// 	query1 := `
//         DELETE FROM breakfast
//         WHERE id = $1`

// 	result, err := tx.ExecContext(ctx, query1, id)

// 	if err != nil {
// 		return err
// 	}

// 	rowsAffected, err := result.RowsAffected()
// 	if err != nil {
// 		return err
// 	}

// 	if rowsAffected == 0 {
// 		return ErrRcordNotFound
// 	}
// 	tx.Commit()

// 	return nil
// }
