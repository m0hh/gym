package main

import (
	"errors"
	"net/http"

	"github.com/m0hh/smart-logitics/internal/data"
	"github.com/m0hh/smart-logitics/internal/validator"
)

func (app *application) addDay(w http.ResponseWriter, r *http.Request) {

	var input_day struct {
		Breakfast int64  `json:"breakfast"`
		AmSnack   int64  `json:"am_snack"`
		Lunch     int64  `json:"lunch"`
		PmSnack   int64  `json:"pm_snack"`
		Dinner    int64  `json:"dinner"`
		Coach     int64  `json:"coach"`
		Name      string `json:"name"`
	}

	err := app.ReadJSON(w, r, &input_day)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	day := &data.Day{
		Breakfast: input_day.Breakfast,
		AmSnack:   input_day.AmSnack,
		Lunch:     input_day.Lunch,
		PmSnack:   input_day.PmSnack,
		Dinner:    input_day.Dinner,
		Name:      input_day.Name,
	}
	v := validator.New()

	if data.ValidateDay(v, *day); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	user := app.contextGetUser(r)

	err = app.models.Plans.InsertDay(day, *user)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRcordNotFound):
			app.notFoundResponse(w, r)
			return
		default:
			app.serverErrorResponse(w, r, err)
			return
		}

	}
	app.writeJSON(w, 200, envelope{"input": day}, nil)
}

func (app *application) retrieveDay(w http.ResponseWriter, r *http.Request) {
	id, err := app.ReadIDParam(r)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	dayFull, err := app.models.Plans.GetDayById(id, &app.models)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRcordNotFound):
			app.notFoundResponse(w, r)
			return
		default:
			app.serverErrorResponse(w, r, err)
			return
		}
	}

	app.writeJSON(w, http.StatusOK, envelope{"day": dayFull}, nil)
}

func (app *application) listDays(w http.ResponseWriter, r *http.Request) {
	user := app.contextGetUser(r)
	var input struct {
		data.Filters
	}

	v := validator.New()
	input.Filters.PageSize = app.readInt(r.URL.Query(), "page_size", 10, v)
	input.Filters.Page = app.readInt(r.URL.Query(), "page_number", 1, v)

	if data.ValidateFilters(v, input.Filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	days, metadata, err := app.models.Plans.GetAllCoachDays(*user, input.Filters)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	app.writeJSON(w, http.StatusOK, envelope{"days": days, "metadata": metadata}, nil)
}

func (app *application) addPlanMeal(w http.ResponseWriter, r *http.Request) {

	var input_plan struct {
		FirstDay   int64 `json:"first_day"`
		SecondDay  int64 `json:"second_day"`
		ThirdDay   int64 `json:"third_day"`
		FourthDay  int64 `json:"fourth_day"`
		FifthDay   int64 `json:"fifth_day"`
		SixthDay   int64 `json:"sixth_day"`
		SeventhDay int64 `json:"seventh_day"`
	}

	err := app.ReadJSON(w, r, &input_plan)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	user := app.contextGetUser(r)

	plan := &data.PlanMeal{
		FirstDay:   input_plan.FirstDay,
		SecondDay:  input_plan.SecondDay,
		ThirdDay:   input_plan.ThirdDay,
		FourthDay:  input_plan.FourthDay,
		FifthDay:   input_plan.FifthDay,
		SixthDay:   input_plan.SixthDay,
		SeventhDay: input_plan.SeventhDay,
		Coach:      user.Id,
	}
	v := validator.New()

	if data.ValidatePlanMeal(v, *plan); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Plans.InsertPlanMeal(plan)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRcordNotFound):
			app.notFoundResponse(w, r)
			return
		default:
			app.serverErrorResponse(w, r, err)
			return
		}

	}
	app.writeJSON(w, 200, envelope{"plan": plan}, nil)
}

func (app *application) listPlans(w http.ResponseWriter, r *http.Request) {
	user := app.contextGetUser(r)
	var input struct {
		data.Filters
	}

	v := validator.New()
	input.Filters.PageSize = app.readInt(r.URL.Query(), "page_size", 10, v)
	input.Filters.Page = app.readInt(r.URL.Query(), "page_number", 1, v)

	if data.ValidateFilters(v, input.Filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	days, metadata, err := app.models.Plans.GetAllPlansCoach(*user, input.Filters)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	app.writeJSON(w, http.StatusOK, envelope{"plans": days, "metadata": metadata}, nil)
}

func (app *application) addPlantoUser(w http.ResponseWriter, r *http.Request) {
	var input struct {
		NewPlan int64 `json:"new_plan"`
		Owner   int64 `json:"owner"`
	}

	err := app.ReadJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	coach := app.contextGetUser(r)

	user_card := &data.UserCard{
		Owner:       input.Owner,
		CurrentPlan: input.NewPlan,
	}

	v := validator.New()

	v.Check(user_card.Owner > 0, "owner", "must provide a valid owner")
	v.Check(user_card.CurrentPlan > 0, "new_plan", "must provide a valid new_plan")

	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Plans.AddPlantoUser(*coach, user_card, app.models.Users)

	if err != nil {
		switch {
		case errors.Is(err, data.ErrRcordNotFound):
			app.notFoundResponse(w, r)
			return
		case errors.Is(err, data.ErrWrongCredentials):
			app.notPermittedResponse(w, r)
			return
		default:
			app.serverErrorResponse(w, r, err)
			return
		}

	}

	w.WriteHeader(http.StatusNoContent)

}
