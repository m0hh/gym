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
		Calories int    `json:"calories"`
	}

	err := app.ReadJSON(w, r, &input)

	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	food := &data.Food{
		FoodName: input.FoodName,
		Serving:  input.Serving,
		Calories: input.Calories,
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

func (app *application) updateFood(w http.ResponseWriter, r *http.Request) {
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
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	var input struct {
		FoodName *string `json:"food_name"`
		Serving  *string `json:"serving"`
		Calories *int    `json:"calories"`
	}

	err = app.ReadJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if input.FoodName != nil {
		food.FoodName = *input.FoodName
	}

	if input.Serving != nil {
		food.Serving = *input.Serving
	}
	if input.Calories != nil {
		food.Calories = *input.Calories
	}
	v := validator.New()
	if data.ValidateFood(v, *food); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	err = app.models.Meals.UpdateFood(food)

	if err != nil {
		switch {
		case errors.Is(err, data.ErrRcordNotFound):
			app.notFoundResponse(w, r)
			return

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

func (app *application) listFoods(w http.ResponseWriter, r *http.Request) {

	var input struct {
		Food_name string
		Serving   string
		data.Filters
	}

	v := validator.New()
	input.Food_name = app.readString(r.URL.Query(), "food_name", "")
	input.Serving = app.readString(r.URL.Query(), "serving", "")
	input.Filters.PageSize = app.readInt(r.URL.Query(), "page_size", 10, v)
	input.Filters.Page = app.readInt(r.URL.Query(), "page_number", 1, v)

	if data.ValidateFilters(v, input.Filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	foods, metadata, err := app.models.Meals.GetAllFood(input.Food_name, input.Serving, input.Filters)

	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"foods": foods, "metadata": metadata}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

func (app *application) addBreakfast(w http.ResponseWriter, r *http.Request) {

	var breakfast_input struct {
		Foods []data.Food `json:"foods"`
	}

	err := app.ReadJSON(w, r, &breakfast_input)

	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	breakfast := &data.Breakfast{
		Food: breakfast_input.Foods,
	}

	v := validator.New()

	if data.ValidateBreakfast(v, *breakfast); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Meals.CreateBreakfast(breakfast)

	if err != nil {
		switch {
		case errors.Is(err, data.ErrWrongForeignKey):
			app.notFoundResponse(w, r)
			return
		default:
			app.serverErrorResponse(w, r, err)
			return
		}

	}

	err = app.writeJSON(w, http.StatusCreated, envelope{"breakfast": breakfast}, nil)

	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) updateBreakfast(w http.ResponseWriter, r *http.Request) {

	id, err := app.ReadIDParam(r)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	var breakfast_input struct {
		Foods []data.Food `json:"foods"`
	}

	err = app.ReadJSON(w, r, &breakfast_input)

	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	breakfast := &data.Breakfast{
		Id:   id,
		Food: breakfast_input.Foods,
	}

	v := validator.New()

	if data.ValidateBreakfast(v, *breakfast); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Meals.UpdateBreakfast(breakfast)

	if err != nil {
		switch {
		case errors.Is(err, data.ErrRcordNotFound):
			app.notFoundResponse(w, r)
			return
		case errors.Is(err, data.ErrWrongForeignKey):
			app.notFoundResponse(w, r)
			return
		default:
			app.serverErrorResponse(w, r, err)
			return
		}

	}

	err = app.writeJSON(w, http.StatusOK, envelope{"breakfast": breakfast}, nil)

	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
