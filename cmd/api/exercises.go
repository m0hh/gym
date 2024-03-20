package main

import (
	"net/http"

	"github.com/m0hh/smart-logitics/internal/data"
	"github.com/m0hh/smart-logitics/internal/validator"
)

func (app *application) addExerciseName(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name string `json:"name"`
	}

	err := app.ReadJSON(w, r, &input)

	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	v := validator.New()

	v.Check(input.Name != "", "name", "must enter a valid name")

	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	exercise := &data.ExerciseName{
		Name: input.Name,
	}

	err = app.models.Exercises.InsertExerciseName(exercise)

	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusCreated, envelope{"ex_name": exercise}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
func (app *application) listExerciseNames(w http.ResponseWriter, r *http.Request) {
	var filter data.Filters
	v := validator.New()
	filter.PageSize = app.readInt(r.URL.Query(), "page_size", 10, v)
	filter.Page = app.readInt(r.URL.Query(), "page_number", 1, v)

	if data.ValidateFilters(v, filter); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	exercises, metadata, err := app.models.Exercises.ListExerciseName(filter)

	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"ex_names": exercises, "metadata": metadata}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}
