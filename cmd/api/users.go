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

func (app *app) createUserHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := utils.ReadJSON(w, r, &input)
	if err != nil {
		app.err.badRequestResponse(w, r, err)
		return
	}

	user := &users.User{
		Name:      input.Name,
		Email:     input.Email,
		Activated: false,
	}

	err = user.Password.Set(input.Password)
	if err != nil {
		app.err.serverErrorResponse(w, r, err)
		return
	}

	v := validator.New()

	if users.ValidateUser(v, user); !v.Valid() {
		app.err.failedValidationResponse(w, r, v.Errors)
		return
	}

	// create user
	err = app.userService.Create(user)
	if err != nil {
		switch {
		case errors.Is(err, users.ErrDuplicateEmail):
			v.AddError("email", "user with this email already exists")
			app.err.failedValidationResponse(w, r, v.Errors)
		default:
			app.err.serverErrorResponse(w, r, err)
		}
		return
	}

	// grand movies:read permission
	err = app.permissionsService.AddForUser(user.Id, "movies:read")
	if err != nil {
		app.err.serverErrorResponse(w, r, err)
		return
	}

	// create activation token
	token, err := app.actTokenService.New(user.Id, 3*24*time.Hour, acttokens.ScopeActivation)
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
			"activationToken": token.Plaintext,
			"userID":          user.Id,
		}

		err = app.mailer.Send(user.Email, "user_welcome.tmpl.html", data)
		if err != nil {
			app.logger.PrintError(err, nil)
		}
	}()

	err = utils.WriteJSON(w, http.StatusAccepted, utils.Envelope{"user": user}, nil)
	if err != nil {
		app.err.serverErrorResponse(w, r, err)
	}
}

func (app *app) getUserHandler(w http.ResponseWriter, r *http.Request) {

}

func (app *app) updateUserHandler(w http.ResponseWriter, r *http.Request) {

}

func (app *app) activateUserHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		TokenPlaintext string `json:"token,omitempty"`
	}

	err := utils.ReadJSON(w, r, &input)
	if err != nil {
		app.err.badRequestResponse(w, r, err)
		return
	}

	v := validator.New()

	if acttokens.ValidateTokenPlaintext(v, input.TokenPlaintext); !v.Valid() {
		app.err.failedValidationResponse(w, r, v.Errors)
		return
	}

	// find user by activation token
	user, err := app.userService.GetByToken(acttokens.ScopeActivation, input.TokenPlaintext)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrRecordNotFound):
			v.AddError("token", "invalid or expired activation token")
			app.err.failedValidationResponse(w, r, v.Errors)
		default:
			app.err.serverErrorResponse(w, r, err)
		}
		return
	}

	// activate account
	user.Activated = true

	// update user
	err = app.userService.Update(user)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrEditConflict):
			app.err.editConflictResponse(w, r)
		default:
			app.err.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.actTokenService.DeleteAllForUser(acttokens.ScopeActivation, user.Id)
	if err != nil {
		app.err.serverErrorResponse(w, r, err)
		return
	}

	err = utils.WriteJSON(w, http.StatusOK, utils.Envelope{"user": user}, nil)
	if err != nil {
		app.err.serverErrorResponse(w, r, err)
	}
}

func (app *app) updateUserPasswordHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Password string `json:"password"`
		Token    string `json:"token"`
	}

	err := utils.ReadJSON(w, r, &input)

	if err != nil {
		app.err.badRequestResponse(w, r, err)
		return
	}

	v := validator.New()

	users.ValidatePasswordPlaintext(v, input.Password)
	acttokens.ValidateTokenPlaintext(v, input.Token)

	if !v.Valid() {
		app.err.failedValidationResponse(w, r, v.Errors)
		return
	}

	// find user by his token
	user, err := app.userService.GetByToken(acttokens.ScopePasswordReset, input.Token)

	if err != nil {
		switch {
		case errors.Is(err, models.ErrRecordNotFound):
			v.AddError("token", "invalid or expired password reset token")
			app.err.failedValidationResponse(w, r, v.Errors)
		default:
			app.err.serverErrorResponse(w, r, err)
		}
		return
	}

	// create new hashed password
	err = user.Password.Set(input.Password)
	if err != nil {
		app.err.serverErrorResponse(w, r, err)
		return
	}

	// update user with new password
	err = app.userService.Update(user)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrEditConflict):
			app.err.editConflictResponse(w, r)
		default:
			app.err.serverErrorResponse(w, r, err)
		}
		return
	}

	// delete all password reset tokens
	err = app.actTokenService.DeleteAllForUser(acttokens.ScopePasswordReset, user.Id)
	if err != nil {
		app.err.serverErrorResponse(w, r, err)
		return
	}

	env := utils.Envelope{"message": "your password was successfully reset"}
	err = utils.WriteJSON(w, http.StatusOK, env, nil)
	if err != nil {
		app.err.serverErrorResponse(w, r, err)
		return
	}
}
