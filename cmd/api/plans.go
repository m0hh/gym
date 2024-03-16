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
