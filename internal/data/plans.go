package data

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"github.com/m0hh/smart-logitics/internal/validator"
)

type Day struct {
	Id        int64 `json:"id"`
	Breakfast int64 `json:"breakfast"`
	AmSnack   int64 `json:"am_snack"`
	Lunch     int64 `json:"lunch"`
	PmSnack   int64 `json:"pm_snack"`
	Dinner    int64 `json:"dinner"`
	Coach     int64 `json:"coach"`
}

type DayFull struct {
	Id        int64     `json:"id"`
	Breakfast Breakfast `json:"breakfast,omitempty"`
	AmSnack   AmSnack   `json:"am_snack,omitempty"`
	Lunch     Lunch     `json:"lunch,omitempty"`
	PmSnack   PmSnack   `json:"pm_snack,omitempty"`
	Dinner    Dinner    `json:"dinner,omitempty"`
	Coach     User      `json:"coach,omitempty"`
}

type PlanModel struct {
	DB *sql.DB
}

func ValidateDay(v *validator.Validator, day Day) {
	v.Check(day.Breakfast > 0, "breakfast", "must provide a valid breakfast id")
	v.Check(day.AmSnack > -1, "am_snack", "must provide a valid am_snack id")
	v.Check(day.Lunch > 0, "lunch", "must provide a valid lunch id")
	v.Check(day.PmSnack > -1, "pm_snack", "must provide a valid pm_snack id")
	v.Check(day.Dinner > 0, "dinner", "must provide a valid dinner id")
	v.Check(day.Coach > 0, "coach", "must provide a valid user id")

}

func (m PlanModel) InsertDay(day *Day) error {

	stmt := ` INSERT INTO day (breakfast_id, am_snack_id,lunch_id , pm_snack_id, dinner_id, coach)
	VALUES ($1, CASE WHEN $2 = 0 THEN NULL ELSE $2 END, 
        $3,CASE WHEN $4 = 0 THEN NULL ELSE $4 END,$5, $6) 
	RETURNING id`

	context, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []interface{}{day.Breakfast, day.AmSnack, day.Lunch, day.PmSnack, day.Dinner, day.Coach}

	err := m.DB.QueryRowContext(context, stmt, args...).Scan(&day.Id)

	if err != nil {
		switch {
		case strings.HasPrefix(err.Error(), `pq: insert or update on table "day" violates foreign key constraint`):
			return ErrRcordNotFound
		default:
			return err

		}
	}

	return nil

}
