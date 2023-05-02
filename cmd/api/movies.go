package main

import (
	"errors"
	"fmt"
	"movies-api/internal/models"
	"movies-api/internal/models/dto"
	"movies-api/internal/utils"
	"movies-api/internal/validator"
	"net/http"
)

func (app *app) showMovieHandler(w http.ResponseWriter, r *http.Request) {
	id, err := utils.ReadIdParam(r)

	if err != nil {
		app.err.notFoundResponse(w, r)
		return
	}

	movie, err := app.movieService.Get(id)

	if err != nil {
		switch {
		case errors.Is(err, models.ErrRecordNotFound):
			app.err.notFoundResponse(w, r)
		default:
			app.err.serverErrorResponse(w, r, err)
		}
		return
	}

	err = utils.WriteJSON(w, http.StatusOK, utils.Envelope{"movie": movie}, nil)

	if err != nil {
		app.err.serverErrorResponse(w, r, err)
	}
}

func (app *app) createMovieHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title   string   `json:"title"`
		Year    int32    `json:"year"`
		Runtime int32    `json:"runtime"`
		Genres  []string `json:"genres"`
	}

	err := utils.ReadJSON(w, r, &input)

	if err != nil {
		app.err.badRequestResponse(w, r, err)
		return
	}

	movie := &models.Movie{
		Title:   input.Title,
		Year:    input.Year,
		Runtime: input.Runtime,
		Genres:  input.Genres,
	}

	v := validator.New()

	if dto.ValidateMovie(v, movie); !v.Valid() {
		app.err.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.movieService.Create(movie)

	if err != nil {
		app.err.serverErrorResponse(w, r, err)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/movies/%d", movie.Id))
	err = utils.WriteJSON(w, http.StatusCreated, utils.Envelope{"movie": movie}, headers)

	if err != nil {
		app.err.serverErrorResponse(w, r, err)
		return
	}
}

func (app *app) updateMovieHandler(w http.ResponseWriter, r *http.Request) {
	// get id
	id, err := utils.ReadIdParam(r)

	if err != nil {
		app.err.notFoundResponse(w, r)
		return
	}

	// get movie
	movie, err := app.movieService.Get(id)

	if err != nil {
		switch {
		case errors.Is(err, models.ErrRecordNotFound):
			app.err.notFoundResponse(w, r)
		default:
			app.err.serverErrorResponse(w, r, err)
		}
		return
	}

	var input struct {
		Title   *string  `json:"title"`
		Year    *int32   `json:"year"`
		Runtime *int32   `json:"runtime"`
		Genres  []string `json:"genres"`
	}

	// read input
	err = utils.ReadJSON(w, r, &input)

	if err != nil {
		app.err.badRequestResponse(w, r, err)
		return
	}

	// if value is nil, then that means that no value was provided
	// and we do not need to change it
	if input.Title != nil {
		movie.Title = *input.Title
	}

	if input.Year != nil {
		movie.Year = *input.Year
	}
	if input.Runtime != nil {
		movie.Runtime = *input.Runtime
	}

	if input.Genres != nil {
		movie.Genres = input.Genres
	}

	v := validator.New()

	// validate
	if dto.ValidateMovie(v, movie); !v.Valid() {
		app.err.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.movieService.Update(movie)

	if err != nil {
		switch {
		case errors.Is(err, models.ErrEditConflict):
			app.err.editConflictResponse(w, r)
		default:
			app.err.serverErrorResponse(w, r, err)
		}
		return
	}

	err = utils.WriteJSON(w, http.StatusCreated, utils.Envelope{"movie": movie}, nil)

	if err != nil {
		app.err.serverErrorResponse(w, r, err)
	}
}

func (app *app) deleteMovieHandler(w http.ResponseWriter, r *http.Request) {
	id, err := utils.ReadIdParam(r)

	if err != nil {
		app.err.notFoundResponse(w, r)
		return
	}

	err = app.movieService.Delete(id)

	if err != nil {
		switch {
		case errors.Is(err, models.ErrRecordNotFound):
			app.err.notFoundResponse(w, r)
		default:
			app.err.serverErrorResponse(w, r, err)
		}
		return
	}

	err = utils.WriteJSON(w, http.StatusOK, utils.Envelope{"movie": "movie successfuly deleted"}, nil)

	if err != nil {
		app.err.serverErrorResponse(w, r, err)
		return
	}

}
