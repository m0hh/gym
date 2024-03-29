package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/m0hh/smart-logitics/internal/data"
	"github.com/m0hh/smart-logitics/internal/validator"
)

var AdminEmail = "mohamed.ehab.desoky.@gmail.com"

func (app *application) registerUserHandler(w http.ResponseWriter, r *http.Request) {
	coach := app.contextGetUser(r)
	var input struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := app.ReadJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	user := &data.User{
		Name:      input.Name,
		Email:     input.Email,
		Activated: true,
		Role:      data.TraineeRole,
	}

	err = user.Password.Set(input.Password)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

	v := validator.New()

	if data.ValidateUser(v, user); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Users.Insert(user)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrDuplicateEmail):
			v.AddError("email", "a user with this email address already exists")
			app.failedValidationResponse(w, r, v.Errors)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	app.background(func() {

		usercard := &data.UserCard{
			Owner: user.Id,
			Coach: coach.Id,
		}
		err = app.models.Users.CreateUserCardRegistration(*usercard)
		if err != nil {
			prop := make(map[string]string)
			prop["User Registration"] = fmt.Sprintf("User %d created but failed to create his card", user.Id)
			app.logger.PrintError(err, prop)
			app.mailer.Send(AdminEmail, "user_card_faliure.html", user.Id)
		}

		data := map[string]interface{}{
			"name":         user.Name,
			"userPassword": input.Password,
		}
		err = app.mailer.Send(user.Email, "user_welcome.html", data)
		if err != nil {
			prop := make(map[string]string)
			prop["User Registration"] = fmt.Sprintf("User %d colud not send welcome email", user.Id)
			app.logger.PrintError(err, prop)
		}
	})

	err = app.writeJSON(w, http.StatusAccepted, envelope{"user": user}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) activateUserHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		TokenPlainText string `json:"token"`
	}

	err := app.ReadJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
	}

	v := validator.New()

	if data.ValidateTokenPlainText(v, input.TokenPlainText); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	user, err := app.models.Users.GetForToken(data.ScopeActivation, input.TokenPlainText)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRcordNotFound):
			v.AddError("token", "invalid or expired activation token")
			app.failedValidationResponse(w, r, v.Errors)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	user.Activated = true

	err = app.models.Users.Update(user)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.models.Tokens.DeleteAllforUser(data.ScopeActivation, user.Id)

	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"user": user}, nil)

	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

func (app *application) updateUserPasswordHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Password       string `json:"password"`
		TokenPlaintext string `json:"token"`
	}

	err := app.ReadJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	v := validator.New()

	data.ValidatePasswordPlaintext(v, input.Password)
	data.ValidateTokenPlainText(v, input.TokenPlaintext)

	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	user, err := app.models.Users.GetForToken(data.ScopePasswordReset, input.TokenPlaintext)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRcordNotFound):
			v.AddError("token", "invalid or expired password reset token")
			app.failedValidationResponse(w, r, v.Errors)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = user.Password.Set(input.Password)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.models.Users.Update(user)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.models.Tokens.DeleteAllforUser(data.ScopePasswordReset, user.Id)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	env := envelope{"message": "your password was successfully reset"}

	err = app.writeJSON(w, http.StatusOK, env, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) listCoachusers(w http.ResponseWriter, r *http.Request) {
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

	users, metadata, err := app.models.Users.ListCoachUsers(*user, input.Filters)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	app.writeJSON(w, http.StatusOK, envelope{"users": users, "metadata": metadata}, nil)
}

func (app *application) updateUserWeight(w http.ResponseWriter, r *http.Request) {
	user := app.contextGetUser(r)

	var input struct {
		NewWeight int `json:"weight"`
	}

	err := app.ReadJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	v := validator.New()
	v.Check(input.NewWeight > 0, "weight", "must enter a valid weight")

	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Users.UpdateWeightUserCard(input.NewWeight, user.Id)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)

}

func (app *application) retrieveUserCard(w http.ResponseWriter, r *http.Request) {
	user := app.contextGetUser(r)

	user_card := &data.UserCard{
		Owner: user.Id,
	}

	err := app.models.Users.RetrieveUserCard(user_card)

	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	app.writeJSON(w, http.StatusOK, envelope{"card": user_card}, nil)

}

func (app *application) listuserHistory(w http.ResponseWriter, r *http.Request) {
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

	histories, metadata, err := app.models.Users.ListHistory(user.Id, input.Filters)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	app.writeJSON(w, http.StatusOK, envelope{"histories": histories, "metadata": metadata}, nil)
}

func (app *application) listuserHistoryCoach(w http.ResponseWriter, r *http.Request) {
	coach := app.contextGetUser(r)
	var input struct {
		data.Filters
	}
	user_id, err := app.ReadIDParam(r)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	ok, err := app.models.Users.CoachPermitted(user_id, coach.Id)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	if !ok {
		app.notPermittedResponse(w, r)
		return
	}

	v := validator.New()

	input.Filters.PageSize = app.readInt(r.URL.Query(), "page_size", 10, v)
	input.Filters.Page = app.readInt(r.URL.Query(), "page_number", 1, v)

	if data.ValidateFilters(v, input.Filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	histories, metadata, err := app.models.Users.ListHistory(user_id, input.Filters)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	app.writeJSON(w, http.StatusOK, envelope{"histories": histories, "metadata": metadata}, nil)
}
