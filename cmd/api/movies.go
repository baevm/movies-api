package main

import (
	"fmt"
	"movies-api/internal/models"
	"movies-api/internal/validator"
	"net/http"
)

func (app *app) showMovieHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIdParam(r)

	if err != nil {
		app.notFoundResponse(w, r)
	}

	movie := models.Movie{
		Id:      id,
		Title:   "Tenet",
		Year:    2020,
		Runtime: 140,
		Genres:  []string{"drama", "thriller"},
		Version: 1,
	}

	err = writeJSON(w, http.StatusOK, envelope{"movie": movie}, nil)

	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *app) createMovieHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title   string   `json:"title"`
		Year    int32    `json:"year"`
		Runtime int32    `json:"runtime"`
		Genres  []string `json:"genres"`
	}

	err := readJSON(w, r, &input)

	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	movie := &models.Movie{
		Title:   input.Title,
		Year:    input.Year,
		Runtime: input.Runtime,
		Genres:  input.Genres,
	}

	v := validator.New()

	if models.ValidateMovie(v, movie); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	fmt.Fprintf(w, "%+v\n", input)
}
