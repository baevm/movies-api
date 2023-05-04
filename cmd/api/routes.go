package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (app *app) routes() http.Handler {
	r := chi.NewRouter()

	r.Use(app.rateLimit)
	r.Use(app.recoverPanic)
	r.Use(app.authenticate)

	r.NotFound(app.err.notFoundResponse)
	r.MethodNotAllowed(app.err.notAllowedResponse)

	r.Route("/v1", func(r chi.Router) {
		r.Get("/healthcheck", app.healthcheckHandler)

		r.Mount("/movies", app.moviesRouter())
		r.Mount("/users", app.usersRouter())
		r.Mount("/tokens", app.tokensRouter())
	})

	return r
}

// /movies
func (app *app) moviesRouter() http.Handler {
	r := chi.NewRouter()

	r.Get("/", app.listMoviesHandler)
	r.Post("/", app.createMovieHandler)
	r.Get("/{id}", app.showMovieHandler)
	r.Patch("/{id}", app.updateMovieHandler)
	r.Delete("/{id}", app.deleteMovieHandler)

	return r
}

// /users
func (app *app) usersRouter() http.Handler {
	r := chi.NewRouter()

	r.Post("/", app.createUserHandler)
	r.Get("/", app.getUserHandler)
	r.Patch("/", app.updateUserHandler)
	r.Put("/activated", app.activateUserHandler)

	return r
}

// /tokens
func (app *app) tokensRouter() http.Handler {
	r := chi.NewRouter()

	r.Post("/authentication", app.createAuthTokenHandler)

	return r
}
