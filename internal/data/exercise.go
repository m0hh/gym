package data

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

type ExerciseName struct {
	Id   int64  `json:"id"`
	Name string `json:"name"`
}

type Exercise struct {
	Id     int64        `json:"id"`
	Name   ExerciseName `json:"name"`
	Sets   int          `json:"sets"`
	Reps   int          `json:"reps"`
	Weight int          `json:"weight"`
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
}

type ExerciseDayFK struct {
	Id        int64   `json:"id"`
	Name      string  `json:"name"`
	Exercises []int64 `json:"exercises"`
}

type ExercisePlanFK struct {
	Id    int64   `json:"id"`
	Name  string  `json:"name"`
	HowTo string  `json:"how_to"`
	Days  []int64 `json:"days"`
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
