package main

import (
	"fmt"
	"movies-api/internal/jsonlog"
	"movies-api/internal/utils"
	"net/http"
)

type CustomError struct {
	logger *jsonlog.Logger
}

func (e *CustomError) logError(r *http.Request, err error) {
	e.logger.PrintError(err, map[string]string{
		"req_method": r.Method,
		"req_url":    r.URL.String(),
	})
}

func (e *CustomError) errorResponse(w http.ResponseWriter, r *http.Request, status int, msg any) {
	env := utils.Envelope{"error": msg}

	err := utils.WriteJSON(w, status, env, nil)

	if err != nil {
		e.logError(r, err)
		w.WriteHeader(500)
	}
}

func (e *CustomError) serverErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	e.logError(r, err)

	msg := "The server encountered a problem and could not process your request"
	e.errorResponse(w, r, http.StatusInternalServerError, msg)
}

func (e *CustomError) notFoundResponse(w http.ResponseWriter, r *http.Request) {
	msg := "The requested resource could not be found"
	e.errorResponse(w, r, http.StatusNotFound, msg)
}

func (e *CustomError) notAllowedResponse(w http.ResponseWriter, r *http.Request) {
	msg := fmt.Sprintf("The %s method is not supported for this resource", r.Method)
	e.errorResponse(w, r, http.StatusMethodNotAllowed, msg)
}

func (e *CustomError) badRequestResponse(w http.ResponseWriter, r *http.Request, err error) {
	e.errorResponse(w, r, http.StatusBadRequest, err.Error())
}

func (e *CustomError) failedValidationResponse(w http.ResponseWriter, r *http.Request, errors map[string]string) {
	e.errorResponse(w, r, http.StatusUnprocessableEntity, errors)
}

func (e *CustomError) editConflictResponse(w http.ResponseWriter, r *http.Request) {
	msg := "unable to update the record due to an edit conflit, please try again"
	e.errorResponse(w, r, http.StatusConflict, msg)
}

func (e *CustomError) rateLimitExceededResponse(w http.ResponseWriter, r *http.Request) {
	msg := "Too many requests. Please try again in a moment"
	e.errorResponse(w, r, http.StatusTooManyRequests, msg)
}

func (e *CustomError) invalidCredentialsResponse(w http.ResponseWriter, r *http.Request) {
	msg := "invalid credentials"
	e.errorResponse(w, r, http.StatusForbidden, msg)
}

func (e *CustomError) invalidAuthenticationTokenResponse(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("WWW-Authenticate", "Bearer")
	
	msg := "invalid authentication token"
	e.errorResponse(w, r, http.StatusForbidden, msg)
}
