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

/////////////////////////////////////////////////////////

func (app *application) addExercisePlan(w http.ResponseWriter, r *http.Request) {

	var input struct {
		Name  string  `json:"name"`
		HowTo string  `json:"how_to"`
		Days  []int64 `json:"days"`
	}

	err := app.ReadJSON(w, r, &input)

	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	exercise_plan := &data.ExercisePlanFK{
		Name:  input.Name,
		HowTo: input.HowTo,
		Days:  input.Days,
	}

	user := app.contextGetUser(r)

	exercise_plan.Coach = user.Id
	v := validator.New()

	if data.ValidateExercisePlan(v, *exercise_plan); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Exercises.InsertExercisePlan(exercise_plan)

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

	err = app.writeJSON(w, http.StatusCreated, envelope{"exercise_plan": exercise_plan}, nil)

	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) updateExcercisePlan(w http.ResponseWriter, r *http.Request) {

	id, err := app.ReadIDParam(r)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	var input struct {
		Name  *string  `json:"name"`
		HowTo *string  `json:"how_to"`
		Days  *[]int64 `json:"days"`
	}

	err = app.ReadJSON(w, r, &input)

	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	user := app.contextGetUser(r)

	exercise_plan := &data.ExercisePlanFK{
		Id:    id,
		Coach: user.Id,
	}

	err = app.models.Exercises.GetExercisePlanIdFk(exercise_plan)

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
		exercise_plan.Name = *input.Name
	}

	if input.HowTo != nil {
		exercise_plan.HowTo = *input.HowTo
	}

	if input.Days != nil {
		exercise_plan.Days = *input.Days
	}

	v := validator.New()

	if data.ValidateExercisePlan(v, *exercise_plan); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Exercises.UpdateExcercisePlan(exercise_plan)

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

	err = app.writeJSON(w, http.StatusOK, envelope{"ex_plan": exercise_plan}, nil)

	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) listExercisePlans(w http.ResponseWriter, r *http.Request) {
	var filter data.Filters
	v := validator.New()
	filter.PageSize = app.readInt(r.URL.Query(), "page_size", 10, v)
	filter.Page = app.readInt(r.URL.Query(), "page_number", 1, v)

	if data.ValidateFilters(v, filter); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	user := app.contextGetUser(r)

	exerciseDays, metadata, err := app.models.Exercises.ListExercisePlans(user.Id, filter)

	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"ex_plans": exerciseDays, "metadata": metadata}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}

func (app *application) getExercisePlan(w http.ResponseWriter, r *http.Request) {

	user := app.contextGetUser(r)

	id, err := app.ReadIDParam(r)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	plan := &data.ExercisePlan{Id: id}

	err = app.models.Exercises.GetExercisePlan(plan, user.Id)

	if err != nil {
		if errors.Is(err, data.ErrRcordNotFound) {
			app.notFoundResponse(w, r)
			return
		}
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"ex_plan": plan}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}

func (app *application) deleteExercisePlan(w http.ResponseWriter, r *http.Request) {
	id, err := app.ReadIDParam(r)

	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	coach := app.contextGetUser(r)

	err = app.models.Exercises.DeleteExercisePlan(id, coach.Id)

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

func (app *application) addExPlantoUser(w http.ResponseWriter, r *http.Request) {
	var input struct {
		NewExPlan int64 `json:"new_ex_plan"`
		Owner     int64 `json:"owner"`
	}

	err := app.ReadJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	coach := app.contextGetUser(r)

	user_card := &data.UserCard{
		Owner:         input.Owner,
		CurrentExPlan: input.NewExPlan,
	}

	v := validator.New()

	v.Check(user_card.Owner > 0, "owner", "must provide a valid owner")
	v.Check(user_card.CurrentExPlan > 0, "new_ex_plan", "must provide a valid new_plan")

	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Exercises.AddExPlantoUser(*coach, user_card, app.models.Users)

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
