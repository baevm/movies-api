package main

import (
	"errors"
	"movies-api/internal/models"
	"movies-api/internal/models/acttokens"
	"movies-api/internal/models/users"
	"movies-api/internal/utils"
	"movies-api/internal/validator"
	"net/http"
	"time"
)

func (app *app) createAuthTokenHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := utils.ReadJSON(w, r, &input)
	if err != nil {
		app.err.badRequestResponse(w, r, err)
		return
	}

	v := validator.New()

	users.ValidateEmail(v, input.Email)
	users.ValidatePasswordPlaintext(v, input.Password)

	if !v.Valid() {
		app.err.failedValidationResponse(w, r, v.Errors)
		return
	}

	user, err := app.userService.GetByEmail(input.Email)

	if err != nil {
		switch {
		case errors.Is(err, models.ErrRecordNotFound):
			app.err.invalidCredentialsResponse(w, r)
		default:
			app.err.serverErrorResponse(w, r, err)
		}
		return
	}

	isMatch, err := user.Password.Matches(input.Password)

	if err != nil {
		app.err.serverErrorResponse(w, r, err)
		return
	}

	if !isMatch {
		app.err.invalidCredentialsResponse(w, r)
		return
	}

	token, err := app.actTokenService.New(user.Id, 24*time.Hour, acttokens.ScopeAuth)

	if err != nil {
		app.err.serverErrorResponse(w, r, err)
		return
	}

	err = utils.WriteJSON(w, http.StatusCreated, utils.Envelope{"auth_token": token}, nil)

	if err != nil {
		app.err.serverErrorResponse(w, r, err)
	}

}
