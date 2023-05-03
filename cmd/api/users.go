package main

import (
	"errors"
	"movies-api/internal/models/users"
	"movies-api/internal/utils"
	"movies-api/internal/validator"
	"net/http"
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

	err = utils.WriteJSON(w, http.StatusCreated, utils.Envelope{"user": user}, nil)

	if err != nil {
		app.err.serverErrorResponse(w, r, err)
	}
}

func (app *app) getUserHandler(w http.ResponseWriter, r *http.Request) {

}

func (app *app) updateUserHandler(w http.ResponseWriter, r *http.Request) {

}
