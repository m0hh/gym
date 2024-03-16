package data

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/m0hh/smart-logitics/internal/validator"
)

type Day struct {
	Id        int64  `json:"id"`
	Breakfast int64  `json:"breakfast"`
	AmSnack   int64  `json:"am_snack"`
	Lunch     int64  `json:"lunch"`
	PmSnack   int64  `json:"pm_snack"`
	Dinner    int64  `json:"dinner"`
	Coach     int64  `json:"coach"`
	Name      string `json:"name"`
}

type CoachDays struct {
	Id   int64  `json:"id"`
	Name string `json:"name"`
}

type DayFull struct {
	Id        int64     `json:"id"`
	Breakfast Breakfast `json:"breakfast,omitempty"`
	AmSnack   AmSnack   `json:"am_snack,omitempty"`
	Lunch     Lunch     `json:"lunch,omitempty"`
	PmSnack   PmSnack   `json:"pm_snack,omitempty"`
	Dinner    Dinner    `json:"dinner,omitempty"`
	Coach     User      `json:"coach,omitempty"`
	Name      string    `json:"name"`
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
	v.Check(day.Name != "", "name", "must provide a valid name")

}

func (m PlanModel) InsertDay(day *Day, user User) error {

	stmt := ` INSERT INTO day (breakfast_id, am_snack_id,lunch_id , pm_snack_id, dinner_id, coach, name)
	VALUES ($1, CASE WHEN $2 = 0 THEN NULL ELSE $2 END, 
        $3,CASE WHEN $4 = 0 THEN NULL ELSE $4 END,$5, $6, $7) 
	RETURNING id`

	context, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []interface{}{day.Breakfast, day.AmSnack, day.Lunch, day.PmSnack, day.Dinner, user.Id, day.Name}

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

func (m PlanModel) GetDayById(id int64, models *Models) (*DayFull, error) {
	stmt := `SELECT breakfast_id,am_snack_id,lunch_id,pm_snack_id,dinner_id,coach,name FROM day WHERE id = $1`

	context, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	var day Day
	var am_snack_id any
	var pm_snack_id any

	err := m.DB.QueryRowContext(context, stmt, id).Scan(&day.Breakfast, &am_snack_id, &day.Lunch, &pm_snack_id, &day.Dinner, &day.Coach, &day.Name)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRcordNotFound
		default:
			return nil, err
		}
	}

	var snackId int64
	snackId, ok := am_snack_id.(int64)
	if ok {
		day.AmSnack = snackId
	}

	snackId, ok = pm_snack_id.(int64)
	if ok {
		day.PmSnack = snackId
	}

	var day_full DayFull

	var breakfast Breakfast
	breakfast.Id = day.Breakfast
	models.Meals.GetAllBreakfastFoodsId(&breakfast)

	var am_snack AmSnack
	if day.AmSnack != 0 {
		am_snack.Id = day.AmSnack
		models.Meals.GetAllAmSnackID(&am_snack)
	}

	var lunch Lunch
	lunch.Id = day.Lunch
	models.Meals.GetAllLunchFoodsId(&lunch)

	var pm_snack PmSnack
	if day.PmSnack != 0 {
		pm_snack.Id = day.PmSnack
		models.Meals.GetAllPmSnackID(&pm_snack)
	}

	var dinner Dinner

	dinner.Id = day.Dinner
	models.Meals.GetAllDinnerID(&dinner)

	day_full.Breakfast = breakfast
	day_full.AmSnack = am_snack
	day_full.Lunch = lunch
	day_full.PmSnack = pm_snack
	day_full.Dinner = dinner
	day_full.Id = int64(id)
	day_full.Name = day.Name
	return &day_full, nil
}

func (m PlanModel) GetAllCoachDays(user User, filter Filters) ([]*CoachDays, Metadata, error) {
	stmt := `SELECT count(*) OVER(), id, name FROM day WHERE coach = $1 LIMIT $2 OFFSET $3`
	context, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	rows, err := m.DB.QueryContext(context, stmt, user.Id, filter.limit(), filter.offset())

	if err != nil {
		return nil, Metadata{}, err
	}

	defer rows.Close()

	var days []*CoachDays

	totalRecords := 0
	for rows.Next() {
		var day CoachDays
		err = rows.Scan(&totalRecords, &day.Id, &day.Name)
		if err != nil {
			return nil, Metadata{}, err
		}
		days = append(days, &day)
	}

	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetadata(totalRecords, filter.Page, filter.PageSize)

	return days, metadata, nil

}
