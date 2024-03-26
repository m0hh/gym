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

type ExerciseName struct {
	Id   int64  `json:"id"`
	Name string `json:"name"`
}

type Exercise struct {
	Id     int64  `json:"id"`
	Name   string `json:"name"`
	Sets   int    `json:"sets"`
	Reps   int    `json:"reps"`
	Weight int    `json:"weight"`
}

type ExerciseDay struct {
	Id        int64      `json:"id"`
	Name      string     `json:"name"`
	Exercises []Exercise `json:"exercises"`
}

type ExercisePlan struct {
	Id    int64    `json:"id"`
	Name  string   `json:"name"`
	HowTo string   `json:"how_to"`
	Days  []string `json:"days"`
}

type ExerciseFK struct {
	Id     int64 `json:"id"`
	Name   int64 `json:"name"`
	Sets   int   `json:"sets"`
	Reps   int   `json:"reps"`
	Weight int   `json:"weight"`
	Coach  int64 `json:"coach"`
}

type ExerciseDayFK struct {
	Id        int64   `json:"id"`
	Name      string  `json:"name"`
	Exercises []int64 `json:"exercises"`
	Coach     int64   `json:"coach"`
}

type ExercisePlanFK struct {
	Id    int64   `json:"id"`
	Name  string  `json:"name"`
	HowTo string  `json:"how_to"`
	Days  []int64 `json:"days"`
	Coach int64   `json:"coach"`
}

type ExerciseModel struct {
	DB *sql.DB
}

func (m ExerciseModel) InsertExerciseName(exercise *ExerciseName) error {
	stmt := `INSERT INTO exercise_name (name) VALUES ($1) RETURNING id`

	context, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(context, stmt, exercise.Name).Scan(&exercise.Id)

	if err != nil {
		return err
	}

	return nil
}

func (m ExerciseModel) ListExerciseName(filter Filters) ([]*ExerciseName, Metadata, error) {
	stmt := `SELECT count(*) OVER(), id , name FROM exercise_name LIMIT $1 OFFSET $2
	`

	context, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(context, stmt, filter.limit(), filter.offset())

	if err != nil {
		return nil, Metadata{}, err
	}
	defer rows.Close()

	var exercises []*ExerciseName

	total_records := 0
	for rows.Next() {
		var exercise ExerciseName
		err = rows.Scan(&total_records, &exercise.Id, &exercise.Name)
		if err != nil {
			return nil, Metadata{}, err
		}
		exercises = append(exercises, &exercise)
	}

	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetadata(total_records, filter.Page, filter.PageSize)

	return exercises, metadata, nil
}

func (m ExerciseModel) GetExerciseName(exercise *ExerciseName) error {
	stmt := `SELECT name FROM exercise_name WHERE id =  $1`

	context, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(context, stmt, exercise.Id).Scan(&exercise.Name)

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

func (m ExerciseModel) DeleteExerciseName(id int64) error {
	stmt := `DELETE FROM exercise_name WHERE id = $1`

	context, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := m.DB.ExecContext(context, stmt, id)

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

	return nil
}

///////////////////////////////////////////////

func ValidateExercise(v *validator.Validator, exercise ExerciseFK) {
	v.Check(exercise.Name > 0, "name", "You must enter a valid FK")
	v.Check(exercise.Reps > 0, "reps", "You must enter a valid number")
	v.Check(exercise.Weight > 0, "weight", "You must enter a valid number")
	v.Check(exercise.Sets > 0, "sets", "You must enter a valid number")

}

func (m ExerciseModel) InsertExercise(exercise *ExerciseFK) error {
	stmt := `INSERT INTO exercise (name, sets, reps, weight, coach) VALUES($1,$2,$3,$4,$5) RETURNING id`

	context, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	args := []interface{}{exercise.Name, exercise.Sets, exercise.Reps, exercise.Weight, exercise.Coach}

	err := m.DB.QueryRowContext(context, stmt, args...).Scan(&exercise.Id)

	if err != nil {
		switch {
		case strings.HasPrefix(err.Error(), `pq: insert or update on table "exercise" violates foreign key constraint`):
			return ErrRcordNotFound
		default:
			return err
		}
	}
	return nil

}

func (m ExerciseModel) ListExercises(user_id int64, filter Filters) ([]*Exercise, Metadata, error) {
	stmt := `SELECT count(*) OVER(), e.id, n.name, e.sets, e.reps, e.weight FROM exercise AS e
	JOIN exercise_name AS n ON e.name = n.id
	WHERE e.coach = $1
	LIMIT $2 OFFSET $3
	`

	context, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(context, stmt, user_id, filter.limit(), filter.offset())

	if err != nil {
		return nil, Metadata{}, err
	}

	defer rows.Close()

	var exercises []*Exercise
	total_records := 0

	for rows.Next() {
		var exercise Exercise

		err = rows.Scan(&total_records, &exercise.Id, &exercise.Name, &exercise.Sets, &exercise.Reps, &exercise.Weight)
		if err != nil {
			return nil, Metadata{}, err
		}

		exercises = append(exercises, &exercise)
	}

	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetadata(total_records, filter.Page, filter.PageSize)

	return exercises, metadata, nil

}

func (m ExerciseModel) GetExercise(exercise *Exercise, user_id int64) error {
	stmt := `SELECT n.name, e.sets, e.reps, e.weight FROM exercise AS e
	JOIN exercise_name AS n ON e.name = n.id
	WHERE e.id = $1 AND e.coach = $2
	`

	context, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(context, stmt, exercise.Id, user_id).Scan(&exercise.Name, &exercise.Sets, &exercise.Reps, &exercise.Weight)

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

func (m ExerciseModel) DeleteExercise(exercise_id int64, coach_id int64) error {
	stmt := `DELETE FROM exercise WHERE id =$1 AND coach= $2`

	context, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := m.DB.ExecContext(context, stmt, exercise_id, coach_id)

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

	return nil

}

// /////////////////////////////////////////////
func ValidateExerciseDay(v *validator.Validator, exercise ExerciseDayFK) {
	v.Check(exercise.Name != "", "name", "must send a name")
	v.Check(exercise.Coach > 0, "coach", "must send a valid coach id")

	v.Check(validator.Unique(exercise.Exercises), "exercies", "You munst send not send the same exercise twice")
	v.Check(len(exercise.Exercises) != 0, "exercises", "You must at least input one exercise")
}

func (m ExerciseModel) InsertExerciseDay(exercise_day *ExerciseDayFK) error {
	tx, err := m.DB.Begin()
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()
	stmt := `INSERT INTO exercise_day (name,coach) VALUES ($1,$2) RETURNING id`

	context, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err = tx.QueryRowContext(context, stmt, exercise_day.Name, exercise_day.Coach).Scan(&exercise_day.Id)

	if err != nil {
		return err
	}

	var bulkInsertValues []interface{}
	bulkInsertStrings := make([]string, 0)
	i := 1
	for _, exercie_name := range exercise_day.Exercises {
		bulkInsertStrings = append(bulkInsertStrings, fmt.Sprintf("($%d,$%d)", i, i+1))
		bulkInsertValues = append(bulkInsertValues, exercie_name, exercise_day.Id)
		i += 2
	}

	stmt1 := fmt.Sprintf(`INSERT INTO exercises_to_day (exercise_id, day_id) VALUES %s`, strings.Join(bulkInsertStrings, ","))
	_, err = tx.ExecContext(context, stmt1, bulkInsertValues...)
	if err != nil {
		if strings.HasPrefix(err.Error(), `pq: insert or update on table "exercises_to_day" violates foreign key constraint`) {
			return ErrRcordNotFound
		}
		return err
	}
	tx.Commit()

	return nil
}

func (m *ExerciseModel) UpdateExcerciseDay(exercise_day *ExerciseDayFK) error {
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

	stmt2 := `UPDATE exercise_day SET name =  $1 WHERE id = $2 AND coach = $3`

	result, err := tx.ExecContext(context, stmt2, exercise_day.Name, exercise_day.Id, exercise_day.Coach)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()

	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		err = ErrRcordNotFound
		return err
	}

	stmt := `DELETE FROM exercises_to_day WHERE day_id = $1`

	result, err = tx.ExecContext(context, stmt, exercise_day.Id)

	if err != nil {

		return err
	}

	rowsAffected, err = result.RowsAffected()

	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		err = ErrRcordNotFound
		return err
	}

	var bulkInsertValues []interface{}
	bulkInsertStrings := make([]string, 0)
	i := 1
	for _, exercies := range exercise_day.Exercises {
		bulkInsertStrings = append(bulkInsertStrings, fmt.Sprintf("($%d,$%d)", i, i+1))
		bulkInsertValues = append(bulkInsertValues, exercies, exercise_day.Id)
		i += 2
	}

	stmt1 := fmt.Sprintf(`INSERT INTO exercises_to_day (exercise_id, day_id) VALUES %s`, strings.Join(bulkInsertStrings, ","))
	_, err = tx.ExecContext(context, stmt1, bulkInsertValues...)
	if err != nil {
		if strings.HasPrefix(err.Error(), `pq: insert or update on table "exercises_to_day" violates foreign key constraint`) {
			return ErrRcordNotFound
		}
		return err
	}
	tx.Commit()

	return nil
}

func (m ExerciseModel) GetExerciseDayIdFk(exercise_day *ExerciseDayFK) error {
	stmt := `SELECT d.name, e.exercise_id FROM  exercise_day AS d JOIN exercises_to_day AS e ON d.id = e.day_id WHERE d.id = $1 AND d.coach = $2`
	context, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(context, stmt, exercise_day.Id, exercise_day.Coach)

	if err != nil {
		return err
	}
	defer rows.Close()

	var ids []int64
	for rows.Next() {
		var id int64
		err = rows.Scan(&exercise_day.Name, &id)
		if err != nil {
			return err
		}
		ids = append(ids, id)
	}

	err = rows.Err()
	if err != nil {
		return err
	}

	exercise_day.Exercises = ids

	return nil

}

func (m ExerciseModel) ListExerciseDays(coach_id int64, filter Filters) ([]*ExerciseDay, Metadata, error) {
	stmt := `SELECT count(*) OVER(), d.id, d.name, e.id,  e.name, e.sets, e.reps, e.weight
	FROM  exercise_day AS d JOIN exercises_to_day AS t ON d.id = t.day_id  JOIN exercise AS e ON t.exercise_id = e.id
	WHERE  d.coach = $1 LIMIT $2 OFFSET $3`

	context, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(context, stmt, coach_id, filter.limit(), filter.offset())

	if err != nil {
		return nil, Metadata{}, err
	}

	defer rows.Close()

	ids_to_exercise_day := make(map[int]*ExerciseDay)
	var exercie_days []*ExerciseDay
	totalRecords := 0

	for rows.Next() {

		var excercise_day_id int64
		var exercie_day_Name string
		exercies := Exercise{}
		err = rows.Scan(
			&totalRecords,
			&excercise_day_id,
			&exercie_day_Name,
			&exercies.Id,
			&exercies.Name,
			&exercies.Sets,
			&exercies.Reps,
			&exercies.Weight,
		)

		if err != nil {
			return nil, Metadata{}, err
		}
		if _, ok := ids_to_exercise_day[int(excercise_day_id)]; !ok {

			exercie_day := ExerciseDay{Id: excercise_day_id, Name: exercie_day_Name}
			ids_to_exercise_day[int(excercise_day_id)] = &exercie_day

			ids_to_exercise_day[int(excercise_day_id)].Exercises = append(ids_to_exercise_day[int(excercise_day_id)].Exercises, exercies)
			exercie_days = append(exercie_days, ids_to_exercise_day[int(excercise_day_id)])
		} else {
			ids_to_exercise_day[int(excercise_day_id)].Exercises = append(ids_to_exercise_day[int(excercise_day_id)].Exercises, exercies)
		}

	}

	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetadata(totalRecords, filter.Page, filter.PageSize)

	return exercie_days, metadata, nil

}

func (m ExerciseModel) GetExerciseDay(day *ExerciseDay, coach_id int64) error {
	stmt := `SELECT   d.name, e.id,  e.name, e.sets, e.reps, e.weight
	FROM  exercise_day AS d JOIN exercises_to_day AS t ON d.id = t.day_id  JOIN exercise AS e ON t.exercise_id = e.id
	WHERE  d.coach = $1 AND d.id = $2`

	context, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(context, stmt, coach_id, day.Id)

	if err != nil {
		return err
	}

	defer rows.Close()

	for rows.Next() {

		exercies := Exercise{}
		err = rows.Scan(
			&day.Name,
			&exercies.Id,
			&exercies.Name,
			&exercies.Sets,
			&exercies.Reps,
			&exercies.Weight,
		)

		if err != nil {
			return err
		}
		day.Exercises = append(day.Exercises, exercies)

	}

	if err = rows.Err(); err != nil {
		return err
	}
	if day.Name == "" {
		return ErrRcordNotFound
	}

	return nil

}

func (m ExerciseModel) DeleteExerciseDay(exercise_day_id, coach_id int64) error {
	tx, err := m.DB.Begin()
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	stmt := `DELETE FROM exercises_to_day WHERE day_id =$1`

	context, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := tx.ExecContext(context, stmt, exercise_day_id)

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		err = ErrRcordNotFound
		return err
	}

	stmt2 := `DELETE FROM exercise_day WHERE id =$1 AND coach = $2`

	result, err = tx.ExecContext(context, stmt2, exercise_day_id, coach_id)

	if err != nil {
		return err
	}

	rowsAffected, err = result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		err = ErrRcordNotFound
		return err
	}

	tx.Commit()
	return nil

}

//////////////////////////////

func ValidateExercisePlan(v *validator.Validator, exercise ExercisePlanFK) {
	v.Check(exercise.Name != "", "name", "must send a name")
	v.Check(exercise.Coach > 0, "coach", "must send a valid coach id")
	v.Check(exercise.HowTo != "", "how_to", "must send a how to")

	v.Check(validator.Unique(exercise.Days), "days", "You munst send not send the same day twice")
	v.Check(len(exercise.Days) != 0, "days", "You must at least input one day")
}

func (m ExerciseModel) InsertExercisePlan(exercise_plan *ExercisePlanFK) error {
	tx, err := m.DB.Begin()
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()
	stmt := `INSERT INTO exercise_plan (name,how_to,coach) VALUES ($1,$2,$3) RETURNING id`

	context, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err = tx.QueryRowContext(context, stmt, exercise_plan.Name, exercise_plan.HowTo, exercise_plan.Coach).Scan(&exercise_plan.Id)

	if err != nil {
		return err
	}

	var bulkInsertValues []interface{}
	bulkInsertStrings := make([]string, 0)
	i := 1
	for _, day_id := range exercise_plan.Days {
		bulkInsertStrings = append(bulkInsertStrings, fmt.Sprintf("($%d,$%d)", i, i+1))
		bulkInsertValues = append(bulkInsertValues, day_id, exercise_plan.Id)
		i += 2
	}

	stmt1 := fmt.Sprintf(`INSERT INTO days_to_plan (day_id, plan_id) VALUES %s`, strings.Join(bulkInsertStrings, ","))
	_, err = tx.ExecContext(context, stmt1, bulkInsertValues...)
	if err != nil {
		if strings.HasPrefix(err.Error(), `pq: insert or update on table "days_to_plan" violates foreign key constraint`) {
			err = ErrRcordNotFound
			return err
		}
		return err
	}
	tx.Commit()

	return nil
}

func (m *ExerciseModel) UpdateExcercisePlan(exercise_plan *ExercisePlanFK) error {
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

	stmt2 := `UPDATE exercise_plan SET name =  $1 , how_to = $2 WHERE id = $3 AND coach = $4`

	result, err := tx.ExecContext(context, stmt2, exercise_plan.Name, exercise_plan.HowTo, exercise_plan.Id, exercise_plan.Coach)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()

	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		err = ErrRcordNotFound
		return err
	}

	stmt := `DELETE FROM days_to_plan WHERE plan_id = $1`

	result, err = tx.ExecContext(context, stmt, exercise_plan.Id)

	if err != nil {

		return err
	}

	rowsAffected, err = result.RowsAffected()

	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		err = ErrRcordNotFound
		return err
	}

	var bulkInsertValues []interface{}
	bulkInsertStrings := make([]string, 0)
	i := 1
	for _, day := range exercise_plan.Days {
		bulkInsertStrings = append(bulkInsertStrings, fmt.Sprintf("($%d,$%d)", i, i+1))
		bulkInsertValues = append(bulkInsertValues, day, exercise_plan.Id)
		i += 2
	}

	stmt1 := fmt.Sprintf(`INSERT INTO days_to_plan (day_id, plan_id) VALUES %s`, strings.Join(bulkInsertStrings, ","))
	_, err = tx.ExecContext(context, stmt1, bulkInsertValues...)
	if err != nil {
		if strings.HasPrefix(err.Error(), `pq: insert or update on table "days_to_plan" violates foreign key constraint`) {
			err = ErrRcordNotFound
			return err
		}
		return err
	}
	tx.Commit()

	return nil
}

func (m ExerciseModel) GetExercisePlanIdFk(exercise_plan *ExercisePlanFK) error {
	stmt := `SELECT p.name,p.how_to, e.day_id FROM  exercise_plan AS p JOIN days_to_plan AS e ON p.id = e.plan_id WHERE p.id = $1 AND p.coach = $2`
	context, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(context, stmt, exercise_plan.Id, exercise_plan.Coach)

	if err != nil {
		return err
	}
	defer rows.Close()

	var ids []int64
	for rows.Next() {
		var id int64
		err = rows.Scan(&exercise_plan.Name, &exercise_plan.HowTo, &id)
		if err != nil {
			return err
		}
		ids = append(ids, id)
	}

	err = rows.Err()
	if err != nil {
		return err
	}

	exercise_plan.Days = ids

	return nil

}

func (m ExerciseModel) ListExercisePlans(coach_id int64, filter Filters) ([]*ExercisePlan, Metadata, error) {
	stmt := `SELECT count(*) OVER(),p.id, p.name,p.how_to, d.name
	FROM  exercise_plan AS p JOIN days_to_plan AS t ON p.id = t.plan_id  JOIN exercise_day AS d ON t.day_id = d.id
	WHERE  p.coach = $1 LIMIT $2 OFFSET $3`

	context, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(context, stmt, coach_id, filter.limit(), filter.offset())

	if err != nil {
		return nil, Metadata{}, err
	}

	defer rows.Close()

	ids_to_exercise_plan := make(map[int64]*ExercisePlan)
	var exercie_plans []*ExercisePlan
	totalRecords := 0

	for rows.Next() {

		var excercise_plan_id int64
		var exercie_plan_Name string
		var exercie_plan_how string

		var day_name string
		err = rows.Scan(
			&totalRecords,
			&excercise_plan_id,
			&exercie_plan_Name,
			&exercie_plan_how,
			&day_name,
		)

		if err != nil {
			return nil, Metadata{}, err
		}
		if _, ok := ids_to_exercise_plan[excercise_plan_id]; !ok {

			exercie_plan := ExercisePlan{Id: excercise_plan_id, Name: exercie_plan_Name, HowTo: exercie_plan_how}
			ids_to_exercise_plan[excercise_plan_id] = &exercie_plan

			ids_to_exercise_plan[excercise_plan_id].Days = append(ids_to_exercise_plan[excercise_plan_id].Days, day_name)
			exercie_plans = append(exercie_plans, ids_to_exercise_plan[excercise_plan_id])
		} else {
			ids_to_exercise_plan[excercise_plan_id].Days = append(ids_to_exercise_plan[excercise_plan_id].Days, day_name)
		}

	}

	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetadata(totalRecords, filter.Page, filter.PageSize)

	return exercie_plans, metadata, nil

}

func (m ExerciseModel) GetExercisePlan(plan *ExercisePlan, coach_id int64) error {
	stmt := `SELECT p.name,p.how_to, d.name
	FROM  exercise_plan AS p JOIN days_to_plan AS t ON p.id = t.plan_id  JOIN exercise_day AS d ON t.day_id = d.id
	WHERE  p.coach = $1 AND p.id = $2`

	context, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(context, stmt, coach_id, plan.Id)

	if err != nil {
		return err
	}

	defer rows.Close()

	for rows.Next() {

		var day_name string
		err = rows.Scan(
			&plan.Name,
			&plan.HowTo,
			&day_name,
		)

		if err != nil {
			return err
		}
		plan.Days = append(plan.Days, day_name)

	}

	if err = rows.Err(); err != nil {
		return err
	}
	if plan.Name == "" {
		return ErrRcordNotFound
	}

	return nil

}

func (m ExerciseModel) DeleteExercisePlan(exercise_plan_id, coach_id int64) error {
	tx, err := m.DB.Begin()
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	stmt := `DELETE FROM days_to_plan WHERE plan_id =$1`

	context, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := tx.ExecContext(context, stmt, exercise_plan_id)

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		err = ErrRcordNotFound
		return err
	}

	stmt2 := `DELETE FROM exercise_plan WHERE id =$1 AND coach = $2`

	result, err = tx.ExecContext(context, stmt2, exercise_plan_id, coach_id)

	if err != nil {
		return err
	}

	rowsAffected, err = result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		err = ErrRcordNotFound
		return err
	}

	tx.Commit()
	return nil

}
