package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (app *app) routes() http.Handler {
	r := chi.NewRouter()

	r.NotFound(app.notFoundResponse)
	r.MethodNotAllowed(app.notAllowedResponse)

	r.Route("/v1", func(r chi.Router) {
		r.Get("/healthcheck", app.healthcheckHandler)
		r.Mount("/movies", app.moviesRouter())
	})

	return r
}

func (app *app) moviesRouter() http.Handler {
	r := chi.NewRouter()

	r.Get("/{id}", app.showMovieHandler)
	r.Post("/", app.createMovieHandler)

	return r
}
