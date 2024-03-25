package main

import (
	"errors"
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

func (app *application) getExerciseName(w http.ResponseWriter, r *http.Request) {
	id, err := app.ReadIDParam(r)

	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	exercise := &data.ExerciseName{
		Id: id,
	}
	err = app.models.Exercises.GetExerciseName(exercise)

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

	err = app.writeJSON(w, http.StatusOK, envelope{"ex_name": exercise}, nil)

	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}

func (app *application) deleteExerciseName(w http.ResponseWriter, r *http.Request) {
	id, err := app.ReadIDParam(r)

	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	err = app.models.Exercises.DeleteExerciseName(id)

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

	w.WriteHeader(http.StatusNoContent)

}

func (app *application) addExercise(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name   int64 `json:"name"`
		Sets   int   `json:"sets"`
		Reps   int   `json:"reps"`
		Weight int   `json:"weight"`
	}

	err := app.ReadJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	exercise := &data.ExerciseFK{
		Name:   input.Name,
		Sets:   input.Sets,
		Reps:   input.Reps,
		Weight: input.Weight,
	}

	user := app.contextGetUser(r)

	exercise.Coach = user.Id

	v := validator.New()
	data.ValidateExercise(v, *exercise)

	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Exercises.InsertExercise(exercise)
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

	err = app.writeJSON(w, http.StatusCreated, envelope{"exercise": exercise}, nil)

	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

func (app *application) listExercises(w http.ResponseWriter, r *http.Request) {
	var filter data.Filters
	v := validator.New()
	filter.PageSize = app.readInt(r.URL.Query(), "page_size", 10, v)
	filter.Page = app.readInt(r.URL.Query(), "page_number", 1, v)

	if data.ValidateFilters(v, filter); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	user := app.contextGetUser(r)

	exercises, metadata, err := app.models.Exercises.ListExercises(user.Id, filter)

	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"exercises": exercises, "metadata": metadata}, nil)

	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

func (app *application) getExercise(w http.ResponseWriter, r *http.Request) {
	id, err := app.ReadIDParam(r)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	user := app.contextGetUser(r)

	exercise := &data.Exercise{
		Id: id,
	}

	err = app.models.Exercises.GetExercise(exercise, user.Id)

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

	err = app.writeJSON(w, http.StatusOK, envelope{"exercise": exercise}, nil)

	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

func (app *application) deleteExercise(w http.ResponseWriter, r *http.Request) {
	id, err := app.ReadIDParam(r)

	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	coach := app.contextGetUser(r)

	err = app.models.Exercises.DeleteExercise(id, coach.Id)

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

	w.WriteHeader(http.StatusNoContent)

}

////////////////////////////////////////////////

func (app *application) addExerciseDay(w http.ResponseWriter, r *http.Request) {

	var input struct {
		Name      string  `json:"name"`
		Exercises []int64 `json:"exercises"`
	}

	err := app.ReadJSON(w, r, &input)

	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	exercise_day := &data.ExerciseDayFK{
		Name:      input.Name,
		Exercises: input.Exercises,
	}

	user := app.contextGetUser(r)

	exercise_day.Coach = user.Id
	v := validator.New()

	if data.ValidateExerciseDay(v, *exercise_day); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Exercises.InsertExerciseDay(exercise_day)

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

	err = app.writeJSON(w, http.StatusCreated, envelope{"exercise_day": exercise_day}, nil)

	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) updateExcerciseDay(w http.ResponseWriter, r *http.Request) {

	id, err := app.ReadIDParam(r)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	var input struct {
		Name      *string  `json:"name"`
		Exercises *[]int64 `json:"exercises"`
	}

	err = app.ReadJSON(w, r, &input)

	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	user := app.contextGetUser(r)

	exercise_day := &data.ExerciseDayFK{
		Id:    id,
		Coach: user.Id,
	}

	err = app.models.Exercises.GetExerciseDayIdFk(exercise_day)

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

	if input.Name != nil {
		exercise_day.Name = *input.Name
	}

	if input.Exercises != nil {
		exercise_day.Exercises = *input.Exercises
	}

	v := validator.New()

	if data.ValidateExerciseDay(v, *exercise_day); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Exercises.UpdateExcerciseDay(exercise_day)

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

	err = app.writeJSON(w, http.StatusOK, envelope{"ex_day": exercise_day}, nil)

	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) listExerciseDays(w http.ResponseWriter, r *http.Request) {
	var filter data.Filters
	v := validator.New()
	filter.PageSize = app.readInt(r.URL.Query(), "page_size", 10, v)
	filter.Page = app.readInt(r.URL.Query(), "page_number", 1, v)

	if data.ValidateFilters(v, filter); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	user := app.contextGetUser(r)

	exerciseDays, metadata, err := app.models.Exercises.ListExerciseDays(user.Id, filter)

	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"ex_days": exerciseDays, "metadata": metadata}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}

func (app *application) getExerciseDay(w http.ResponseWriter, r *http.Request) {

	user := app.contextGetUser(r)

	id, err := app.ReadIDParam(r)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	day := &data.ExerciseDay{Id: id}

	err = app.models.Exercises.GetExerciseDay(day, user.Id)

	if err != nil {
		if errors.Is(err, data.ErrRcordNotFound) {
			app.notFoundResponse(w, r)
			return
		}
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"ex_day": day}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}

func (app *application) deleteExerciseDay(w http.ResponseWriter, r *http.Request) {
	id, err := app.ReadIDParam(r)

	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	coach := app.contextGetUser(r)

	err = app.models.Exercises.DeleteExerciseDay(id, coach.Id)

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

	w.WriteHeader(http.StatusNoContent)

}
