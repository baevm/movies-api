package main

import (
	"fmt"
	"movies-api/internal/jsonlog"
	"movies-api/internal/utils"
	"net/http"
)

type Error struct {
	logger *jsonlog.Logger
}

func (e *Error) logError(r *http.Request, err error) {
	e.logger.PrintError(err, map[string]string{
		"req_method": r.Method,
		"req_url":    r.URL.String(),
	})
}

func (e *Error) errorResponse(w http.ResponseWriter, r *http.Request, status int, msg any) {
	env := utils.Envelope{"error": msg}

	err := utils.WriteJSON(w, status, env, nil)

	if err != nil {
		e.logError(r, err)
		w.WriteHeader(500)
	}
}

func (e *Error) serverErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	e.logError(r, err)

	msg := "The server encountered a problem and could not process your request"
	e.errorResponse(w, r, http.StatusInternalServerError, msg)
}

func (e *Error) notFoundResponse(w http.ResponseWriter, r *http.Request) {
	msg := "The requested resource could not be found"
	e.errorResponse(w, r, http.StatusNotFound, msg)
}

func (e *Error) notAllowedResponse(w http.ResponseWriter, r *http.Request) {
	msg := fmt.Sprintf("The %s method is not supported for this resource", r.Method)
	e.errorResponse(w, r, http.StatusMethodNotAllowed, msg)
}

func (e *Error) badRequestResponse(w http.ResponseWriter, r *http.Request, err error) {
	e.errorResponse(w, r, http.StatusBadRequest, err.Error())
}

func (e *Error) failedValidationResponse(w http.ResponseWriter, r *http.Request, errors map[string]string) {
	e.errorResponse(w, r, http.StatusUnprocessableEntity, errors)
}

func (e *Error) editConflictResponse(w http.ResponseWriter, r *http.Request) {
	msg := "unable to update the record due to an edit conflit, please try again"
	e.errorResponse(w, r, http.StatusConflict, msg)
}
