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
	Id    int64         `json:"id"`
	Name  string        `json:"name"`
	HowTo string        `json:"how_to"`
	Days  []ExerciseDay `json:"days"`
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
