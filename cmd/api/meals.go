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
	user := app.contextGetUser(r)

	food := &data.Food{
		FoodName: input.FoodName,
		Serving:  input.Serving,
		Calories: input.Calories,
		Coach:    user.Id,
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

func (app *application) retriveFood(w http.ResponseWriter, r *http.Request) {
	id, err := app.ReadIDParam(r)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	user := app.contextGetUser(r)

	food := &data.Food{
		Id:    id,
		Coach: user.Id,
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
	user := app.contextGetUser(r)

	food := &data.Food{
		Id:    id,
		Coach: user.Id,
	}
	err = app.models.Meals.GetById(food)

	if food.Coach != user.Id {
		app.notPermittedResponse(w, r)
		return
	}

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

	user := app.contextGetUser(r)

	if data.ValidateFilters(v, input.Filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	foods, metadata, err := app.models.Meals.GetAllFood(input.Food_name, input.Serving, input.Filters, user.Id)

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

	user := app.contextGetUser(r)

	breakfast := &data.Breakfast{
		Food:  breakfast_input.Foods,
		Coach: user.Id,
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

	user := app.contextGetUser(r)

	breakfast := &data.Breakfast{
		Id:    id,
		Food:  breakfast_input.Foods,
		Coach: user.Id,
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

func (app *application) getBreakfastFoodById(w http.ResponseWriter, r *http.Request) {
	id, err := app.ReadIDParam(r)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	user := app.contextGetUser(r)
	breakfast := &data.Breakfast{
		Id:    id,
		Coach: user.Id,
	}

	err = app.models.Meals.GetAllBreakfastFoodsId(breakfast)

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

	err = app.writeJSON(w, http.StatusCreated, envelope{"breakfast": breakfast}, nil)

	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) listBreakfasts(w http.ResponseWriter, r *http.Request) {

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
	user := app.contextGetUser(r)

	breakfasts, metadata, err := app.models.Meals.GetAllBreakfastFoods(input.Filters, user.Id)

	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"breakfasts": breakfasts, "metadata": metadata}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

func (app *application) deleteFoodHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.ReadIDParam(r)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	user := app.contextGetUser(r)

	err = app.models.Meals.DeleteFood(id, user.Id)

	if err != nil {
		switch {
		case errors.Is(err, data.ErrRcordNotFound):
			app.notFoundResponse(w, r)
			return
		case errors.Is(err, data.ErrFKConflict):
			app.conflictResponse(w, r)
			return
		default:
			app.serverErrorResponse(w, r, err)
			return
		}
	}

	w.WriteHeader(http.StatusNoContent)
}

func (app *application) deleteBreakfastHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.ReadIDParam(r)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	user := app.contextGetUser(r)
	err = app.models.Meals.DeleteBreakfast(id, user.Id)

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

///////////////////////

func (app *application) addAmSnack(w http.ResponseWriter, r *http.Request) {

	var am_snack_input struct {
		Foods []data.Food `json:"foods"`
	}

	err := app.ReadJSON(w, r, &am_snack_input)

	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	user := app.contextGetUser(r)
	am_snack := &data.AmSnack{
		Food:  am_snack_input.Foods,
		Coach: user.Id,
	}

	v := validator.New()

	if data.ValidateAmSnack(v, *am_snack); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Meals.CreateAmSnack(am_snack)

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

	err = app.writeJSON(w, http.StatusCreated, envelope{"am_snack": am_snack}, nil)

	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) updateAmSnack(w http.ResponseWriter, r *http.Request) {

	id, err := app.ReadIDParam(r)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	var am_snack_input struct {
		Foods []data.Food `json:"foods"`
	}

	err = app.ReadJSON(w, r, &am_snack_input)

	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	user := app.contextGetUser(r)

	am_snack := &data.AmSnack{
		Id:    id,
		Food:  am_snack_input.Foods,
		Coach: user.Id,
	}

	v := validator.New()

	if data.ValidateAmSnack(v, *am_snack); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Meals.UpdateAmSnack(am_snack)

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

	err = app.writeJSON(w, http.StatusOK, envelope{"am_snack": am_snack}, nil)

	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) getAmSnackFoodById(w http.ResponseWriter, r *http.Request) {
	id, err := app.ReadIDParam(r)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	user := app.contextGetUser(r)

	am_snack := &data.AmSnack{
		Id:    id,
		Coach: user.Id,
	}

	err = app.models.Meals.GetAllAmSnackID(am_snack)

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

	err = app.writeJSON(w, http.StatusCreated, envelope{"am_snack": am_snack}, nil)

	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) listAmSnacks(w http.ResponseWriter, r *http.Request) {

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

	user := app.contextGetUser(r)

	am_snacks, metadata, err := app.models.Meals.GetAllAmSnacks(input.Filters, user.Id)

	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"am_snacks": am_snacks, "metadata": metadata}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

func (app *application) deleteAmSnackHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.ReadIDParam(r)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	user := app.contextGetUser(r)

	err = app.models.Meals.DeleteAmSnack(id, user.Id)

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

//////////////////////////////

func (app *application) addLunch(w http.ResponseWriter, r *http.Request) {

	var lunch_input struct {
		Foods []data.Food `json:"foods"`
	}

	err := app.ReadJSON(w, r, &lunch_input)

	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	user := app.contextGetUser(r)

	lunch := &data.Lunch{
		Food:  lunch_input.Foods,
		Coach: user.Id,
	}

	v := validator.New()

	if data.ValidateLunch(v, *lunch); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Meals.CreateLunch(lunch)

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

	err = app.writeJSON(w, http.StatusCreated, envelope{"lunch": lunch}, nil)

	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) updateLunch(w http.ResponseWriter, r *http.Request) {

	id, err := app.ReadIDParam(r)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	var lunch_input struct {
		Foods []data.Food `json:"foods"`
	}

	err = app.ReadJSON(w, r, &lunch_input)

	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	user := app.contextGetUser(r)

	lunch := &data.Lunch{
		Id:    id,
		Food:  lunch_input.Foods,
		Coach: user.Id,
	}

	v := validator.New()

	if data.ValidateLunch(v, *lunch); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Meals.UpdateLunch(lunch)

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

	err = app.writeJSON(w, http.StatusOK, envelope{"lunch": lunch}, nil)

	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) getLunchById(w http.ResponseWriter, r *http.Request) {
	id, err := app.ReadIDParam(r)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	user := app.contextGetUser(r)

	lunch := &data.Lunch{
		Id:    id,
		Coach: user.Id,
	}

	err = app.models.Meals.GetAllLunchFoodsId(lunch)

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

	err = app.writeJSON(w, http.StatusCreated, envelope{"lunch": lunch}, nil)

	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) listlunches(w http.ResponseWriter, r *http.Request) {

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

	user := app.contextGetUser(r)

	lunches, metadata, err := app.models.Meals.GetAllLunches(input.Filters, user.Id)

	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"lunches": lunches, "metadata": metadata}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

func (app *application) deleteLunchHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.ReadIDParam(r)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	user := app.contextGetUser(r)

	err = app.models.Meals.DeleteLunch(id, user.Id)

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

//////////////////////////////////////

func (app *application) addPmSnack(w http.ResponseWriter, r *http.Request) {

	var pm_snack_input struct {
		Foods []data.Food `json:"foods"`
	}

	err := app.ReadJSON(w, r, &pm_snack_input)

	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	user := app.contextGetUser(r)

	pm_snack := &data.PmSnack{
		Food:  pm_snack_input.Foods,
		Coach: user.Id,
	}

	v := validator.New()

	if data.ValidatePmSnack(v, *pm_snack); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Meals.CreatePmSnack(pm_snack)

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

	err = app.writeJSON(w, http.StatusCreated, envelope{"pm_snack": pm_snack}, nil)

	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) updatePmSnack(w http.ResponseWriter, r *http.Request) {

	id, err := app.ReadIDParam(r)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	var pm_snack_input struct {
		Foods []data.Food `json:"foods"`
	}

	err = app.ReadJSON(w, r, &pm_snack_input)

	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	user := app.contextGetUser(r)

	pm_snack := &data.PmSnack{
		Id:    id,
		Food:  pm_snack_input.Foods,
		Coach: user.Id,
	}

	v := validator.New()

	if data.ValidatePmSnack(v, *pm_snack); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Meals.UpdatePmSnack(pm_snack)

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

	err = app.writeJSON(w, http.StatusOK, envelope{"pm_snack": pm_snack}, nil)

	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) getPmSnackById(w http.ResponseWriter, r *http.Request) {
	id, err := app.ReadIDParam(r)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	user := app.contextGetUser(r)

	pm_snack := &data.PmSnack{
		Id:    id,
		Coach: user.Id,
	}

	err = app.models.Meals.GetAllPmSnackID(pm_snack)

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

	err = app.writeJSON(w, http.StatusCreated, envelope{"pm_snack": pm_snack}, nil)

	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) listPmsnacks(w http.ResponseWriter, r *http.Request) {

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

	user := app.contextGetUser(r)

	pm_snacks, metadata, err := app.models.Meals.GetAllPmSnacks(input.Filters, user.Id)

	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"pm_snacks": pm_snacks, "metadata": metadata}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

func (app *application) deletePmSnackHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.ReadIDParam(r)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	user := app.contextGetUser(r)

	err = app.models.Meals.DeletePmSnack(id, user.Id)

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

///////////////////////////////////

func (app *application) addDinner(w http.ResponseWriter, r *http.Request) {

	var dinner_input struct {
		Foods []data.Food `json:"foods"`
	}

	err := app.ReadJSON(w, r, &dinner_input)

	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	user := app.contextGetUser(r)

	dinner := &data.Dinner{
		Food:  dinner_input.Foods,
		Coach: user.Id,
	}

	v := validator.New()

	if data.ValidateDinner(v, *dinner); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Meals.CreateDinner(dinner)

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

	err = app.writeJSON(w, http.StatusCreated, envelope{"dinner": dinner}, nil)

	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) updateDinner(w http.ResponseWriter, r *http.Request) {

	id, err := app.ReadIDParam(r)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	var dinner_input struct {
		Foods []data.Food `json:"foods"`
	}

	err = app.ReadJSON(w, r, &dinner_input)

	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	user := app.contextGetUser(r)

	dinner := &data.Dinner{
		Id:    id,
		Food:  dinner_input.Foods,
		Coach: user.Id,
	}

	v := validator.New()

	if data.ValidateDinner(v, *dinner); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Meals.UpdateDinner(dinner)

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

	err = app.writeJSON(w, http.StatusOK, envelope{"dinner": dinner}, nil)

	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) getDinnerById(w http.ResponseWriter, r *http.Request) {
	id, err := app.ReadIDParam(r)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	user := app.contextGetUser(r)

	dinner := &data.Dinner{
		Id:    id,
		Coach: user.Id,
	}

	err = app.models.Meals.GetAllDinnerID(dinner)

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

	err = app.writeJSON(w, http.StatusCreated, envelope{"dinner": dinner}, nil)

	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) listDinners(w http.ResponseWriter, r *http.Request) {

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

	user := app.contextGetUser(r)

	dinners, metadata, err := app.models.Meals.GetAllDinners(input.Filters, user.Id)

	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"dinners": dinners, "metadata": metadata}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

func (app *application) deleteDinnerHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.ReadIDParam(r)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	user := app.contextGetUser(r)
	err = app.models.Meals.DeleteDinner(id, user.Id)

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
