package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (app *app) routes() http.Handler {
	r := chi.NewRouter()

	r.NotFound(app.err.notFoundResponse)
	r.MethodNotAllowed(app.err.notAllowedResponse)
	r.Use(app.recoverPanic)

	r.Route("/v1", func(r chi.Router) {
		r.Get("/healthcheck", app.healthcheckHandler)

		r.Mount("/movies", app.moviesRouter())
	})

	return r
}

func (app *app) moviesRouter() http.Handler {
	r := chi.NewRouter()

	r.Get("/", app.listMoviesHandler)
	r.Post("/", app.createMovieHandler)
	r.Get("/{id}", app.showMovieHandler)
	r.Patch("/{id}", app.updateMovieHandler)
	r.Delete("/{id}", app.deleteMovieHandler)

	return r
}
