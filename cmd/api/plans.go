package main

import (
	"errors"
	"net/http"

	"github.com/m0hh/smart-logitics/internal/data"
	"github.com/m0hh/smart-logitics/internal/validator"
)

func (app *application) addDay(w http.ResponseWriter, r *http.Request) {

	var input_day struct {
		Breakfast int64 `json:"breakfast"`
		AmSnack   int64 `json:"am_snack"`
		Lunch     int64 `json:"lunch"`
		PmSnack   int64 `json:"pm_snack"`
		Dinner    int64 `json:"dinner"`
		Coach     int64 `json:"coach"`
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
		Coach:     input_day.Coach,
	}
	v := validator.New()

	if data.ValidateDay(v, *day); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Plans.InsertDay(day)
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
