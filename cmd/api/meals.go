package main

import (
	"errors"
	"net/http"

	"github.com/m0hh/smart-logitics/internal/data"
	"github.com/m0hh/smart-logitics/internal/validator"
)

func (app *application) addFood(w http.ResponseWriter, r *http.Request) {
	var input struct {
		FoodName string `json:"food_name"`
		Serving  string `json:"serving"`
	}

	err := app.ReadJSON(w, r, &input)

	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	food := &data.Food{
		FoodName: input.FoodName,
		Serving:  input.Serving,
	}

	v := validator.New()

	if data.ValidateFood(v, *food); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Meals.CreateFood(food)

	if err != nil {
		switch {
		case errors.Is(err, data.ErrUniqueFood):
			v.AddError("food", "food name and serving must be unique")
			app.failedValidationResponse(w, r, v.Errors)
			return
		default:
			app.serverErrorResponse(w, r, err)
			return
		}
	}

	err = app.writeJSON(w, http.StatusCreated, envelope{"food": food}, nil)

	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) retriveFood(w http.ResponseWriter, r *http.Request) {
	id, err := app.ReadIDParam(r)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	food := &data.Food{
		Id: id,
	}

	err = app.models.Meals.GetById(food)

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

	err = app.writeJSON(w, http.StatusCreated, envelope{"food": food}, nil)

	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
