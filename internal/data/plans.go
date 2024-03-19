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

type PlanMeal struct {
	Id         int64 `json:"id"`
	FirstDay   int64 `json:"first_day"`
	SecondDay  int64 `json:"second_day"`
	ThirdDay   int64 `json:"third_day"`
	FourthDay  int64 `json:"fourth_day"`
	FifthDay   int64 `json:"fifth_day"`
	SixthDay   int64 `json:"sixth_day"`
	SeventhDay int64 `json:"seventh_day"`
	Coach      int64 `json:"coach"`
}

type PlanMealCoachList struct {
	Id         int64  `json:"id"`
	FirstDay   string `json:"first_day"`
	SecondDay  string `json:"second_day"`
	ThirdDay   string `json:"third_day"`
	FourthDay  string `json:"fourth_day"`
	FifthDay   string `json:"fifth_day"`
	SixthDay   string `json:"sixth_day"`
	SeventhDay string `json:"seventh_day"`
}

type PlanMealFull struct {
	Id         int64   `json:"id"`
	FirstDay   DayFull `json:"first_day"`
	SecondDay  DayFull `json:"second_day"`
	ThirdDay   DayFull `json:"third_day"`
	FourthDay  DayFull `json:"fourth_day"`
	FifthDay   DayFull `json:"fifth_day"`
	SixthDay   DayFull `json:"sixth_day"`
	SeventhDay DayFull `json:"seventh_day"`
	Coach      User    `json:"coach"`
}

func ValidatePlanMeal(v *validator.Validator, plan PlanMeal) {
	v.Check(plan.FirstDay > 0, "first_day", "must provide a valid integer")
	v.Check(plan.SecondDay > 0, "second_day", "must provide a valid integer")
	v.Check(plan.ThirdDay > 0, "third_day", "must provide a valid integer")
	v.Check(plan.FourthDay > 0, "fourth_day", "must provide a valid integer")
	v.Check(plan.FifthDay > 0, "fifth_day", "must provide a valid integer")
	v.Check(plan.SixthDay > 0, "sixth_day", "must provide a valid integer")
	v.Check(plan.SeventhDay > 0, "seventh_day", "must provide a valid integer")
	v.Check(plan.Coach > 0, "coach", "must provide a valid integer")
}

func (m PlanModel) InsertPlanMeal(plan_meal *PlanMeal) error {
	stmt := ` INSERT INTO plan_meal (first_day, second_day, third_day, fourth_day, fifth_day, sixth_day,seventh_day, coach) 
	VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
	RETURNING id`

	context, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []interface{}{plan_meal.FirstDay, plan_meal.SecondDay, plan_meal.ThirdDay,
		plan_meal.FourthDay, plan_meal.FifthDay, plan_meal.SixthDay, plan_meal.SeventhDay, plan_meal.Coach}

	err := m.DB.QueryRowContext(context, stmt, args...).Scan(&plan_meal.Id)

	if err != nil {
		switch {
		case strings.HasPrefix(err.Error(), `pq: insert or update on table "plan_meal" violates foreign key constraint`):
			return ErrRcordNotFound
		default:
			return err
		}
	}

	return nil
}

func (m PlanModel) GetAllPlansCoach(user User, filter Filters) ([]*PlanMealCoachList, Metadata, error) {
	stmt := `
	SELECT 
	count(*) OVER(),
    pm.id ,
    d1.name ,
    d2.name ,
    d3.name ,
    d4.name ,
    d5.name ,
    d6.name ,
    d7.name 
	FROM 
		plan_meal AS pm
	JOIN 
		day AS d1 ON pm.first_day = d1.id
	JOIN 
		day AS d2 ON pm.second_day = d2.id
	JOIN 
		day AS d3 ON pm.third_day = d3.id
	JOIN 
		day AS d4 ON pm.fourth_day = d4.id
	JOIN 
		day AS d5 ON pm.fifth_day = d5.id
	JOIN 
		day AS d6 ON pm.sixth_day = d6.id
	JOIN 
		day AS d7 ON pm.seventh_day = d7.id
	WHERE pm.coach = $1
	LIMIT $2 OFFSET $3
	`
	context, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(context, stmt, user.Id, filter.limit(), filter.offset())

	if err != nil {
		return nil, Metadata{}, err
	}

	defer rows.Close()

	var plans []*PlanMealCoachList

	totalRecords := 0
	for rows.Next() {
		var plan PlanMealCoachList
		err = rows.Scan(&totalRecords, &plan.Id, &plan.FirstDay, &plan.SecondDay, &plan.ThirdDay, &plan.FourthDay, &plan.FifthDay, &plan.SixthDay, &plan.SeventhDay)
		if err != nil {
			return nil, Metadata{}, err
		}
		plans = append(plans, &plan)
	}

	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetadata(totalRecords, filter.Page, filter.PageSize)

	return plans, metadata, nil

}

func (m PlanModel) AddPlantoUser(coach User, user_card *UserCard, u UserModel) error {
	plan_id := user_card.CurrentPlan
	err := u.RetrieveUserCard(user_card)
	if err != nil {
		switch {
		case errors.Is(err, ErrRcordNotFound):
			return ErrRcordNotFound
		default:
			return err
		}
	}

	if user_card.Coach != coach.Id {
		return ErrWrongCredentials
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

	stmt1 := `UPDATE user_history SET to_date = $1, weight_finish = $2,  is_now = false WHERE is_now = true AND owner = $3`

	context, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err = tx.ExecContext(context, stmt1, time.Now(), user_card.Weight, user_card.Owner)

	if err != nil {
		return err
	}

	stmt := `UPDATE user_card SET current_plan = $1 WHERE owner = $2`

	_, err = tx.ExecContext(context, stmt, plan_id, user_card.Owner)

	if err != nil {
		switch {
		case strings.HasPrefix(err.Error(), `pq: insert or update on table "user_card" violates foreign key constraint`):
			return ErrRcordNotFound
		default:
			return err
		}
	}

	stmt2 := `INSERT INTO user_history (plan_done, weight_start, owner) VALUES($1,$2,$3)`

	_, err = tx.ExecContext(context, stmt2, plan_id, user_card.Weight, user_card.Owner)

	if err != nil {
		return err
	}

	tx.Commit()
	return nil
}
