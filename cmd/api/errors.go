package main

import (
	"fmt"
	"net/http"
)

func (app *app) logError(r *http.Request, err error) {
	app.logger.Print(err)
}

func (app *app) errorResponse(w http.ResponseWriter, r *http.Request, status int, msg any) {
	env := envelope{"error": msg}

	err := writeJSON(w, status, env, nil)

	if err != nil {
		app.logError(r, err)
		w.WriteHeader(500)
	}
}

func (app *app) serverErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.logError(r, err)

	msg := "The server encountered a problem and could not process your request"
	app.errorResponse(w, r, http.StatusInternalServerError, msg)
}

func (app *app) notFoundResponse(w http.ResponseWriter, r *http.Request) {
	msg := "The requested resource could not be found"
	app.errorResponse(w, r, http.StatusNotFound, msg)
}

func (app *app) notAllowedResponse(w http.ResponseWriter, r *http.Request) {
	msg := fmt.Sprintf("The %s method is not supported for this resource", r.Method)
	app.errorResponse(w, r, http.StatusMethodNotAllowed, msg)
}

func (app *app) badRequestResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.errorResponse(w, r, http.StatusBadRequest, err.Error())
}

func (app *app) failedValidationResponse(w http.ResponseWriter, r *http.Request, errors map[string]string) {
	app.errorResponse(w, r, http.StatusUnprocessableEntity, errors)
}
