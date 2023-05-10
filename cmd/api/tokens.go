package main

import (
	"errors"
	"fmt"
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

func (app *app) createPasswordResetTokenHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email string `json:"email,omitempty"`
	}

	err := utils.ReadJSON(w, r, &input)

	if err != nil {
		app.err.badRequestResponse(w, r, err)
		return
	}

	v := validator.New()

	if users.ValidateEmail(v, input.Email); !v.Valid() {
		app.err.failedValidationResponse(w, r, v.Errors)
		return
	}

	user, err := app.userService.GetByEmail(input.Email)

	if err != nil {
		switch {
		case errors.Is(err, models.ErrRecordNotFound):
			v.AddError("email", "no matching email address found")
			app.err.failedValidationResponse(w, r, v.Errors)
		default:
			app.err.serverErrorResponse(w, r, err)
		}
		return
	}

	if !user.Activated {
		v.AddError("email", "user account must activated")
		app.err.notAllowedResponse(w, r)
		return
	}

	token, err := app.actTokenService.New(user.Id, 45*time.Minute, acttokens.ScopePasswordReset)

	if err != nil {
		app.err.serverErrorResponse(w, r, err)
		return
	}

	// send email in background
	app.wg.Add(1)
	go func() {
		defer app.wg.Done()

		// recover to catch any panics
		defer func() {
			if err := recover(); err != nil {
				app.logger.PrintError(fmt.Errorf("%s", err), nil)
			}
		}()

		data := map[string]any{
			"passwordResetToken": token.Plaintext,
		}

		err = app.mailer.Send(user.Email, "token_password_reset.tmpl.html", data)
		if err != nil {
			app.logger.PrintError(err, nil)
		}
	}()

	env := utils.Envelope{"message": "an email will be sent to you containing password reset instructions"}

	err = utils.WriteJSON(w, http.StatusAccepted, env, nil)

	if err != nil {
		app.err.serverErrorResponse(w, r, err)
	}
}
