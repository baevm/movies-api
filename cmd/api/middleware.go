package main

import (
	"errors"
	"fmt"
	"movies-api/internal/context"
	"movies-api/internal/models"
	"movies-api/internal/models/acttokens"
	"movies-api/internal/models/users"
	"movies-api/internal/validator"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

func (app *app) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("Connection", "close")
				app.err.serverErrorResponse(w, r, fmt.Errorf("%s", err))
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func (app *app) rateLimit(next http.Handler) http.Handler {
	type client struct {
		limiter  *rate.Limiter
		lastSeen time.Time
	}

	var (
		mu      sync.Mutex
		clients = make(map[string]*client)
	)

	// goroutine to clear client if they
	// havent been seen in last 3 minutes
	go func() {
		for {
			time.Sleep(time.Minute)

			mu.Lock()

			for ip, client := range clients {
				if time.Since(client.lastSeen) > 3*time.Minute {
					delete(clients, ip)
				}
			}

			mu.Unlock()
		}
	}()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip, _, err := net.SplitHostPort(r.RemoteAddr)

		if err != nil {
			app.err.serverErrorResponse(w, r, err)
			return
		}

		mu.Lock()

		if _, found := clients[ip]; !found {
			clients[ip] = &client{
				limiter: rate.NewLimiter(rate.Limit(app.config.limiter.rps), app.config.limiter.burst),
			}
		}

		clients[ip].lastSeen = time.Now()

		if !clients[ip].limiter.Allow() {
			mu.Unlock()
			app.err.rateLimitExceededResponse(w, r)
			return
		}

		mu.Unlock()

		next.ServeHTTP(w, r)
	})
}

func (app *app) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Authorization")

		// get auth header from req
		authHeader := r.Header.Get("Authorization")

		// if auth header is empty set
		// anon user in context and return
		if authHeader == "" {
			r = context.ContextSetUser(r, users.AnonUser)

			next.ServeHTTP(w, r)
			return
		}

		// split auth header in 2 parts
		headerParts := strings.Split(authHeader, " ")

		if len(headerParts) != 2 || headerParts[0] != "Bearer" {
			app.err.invalidAuthenticationTokenResponse(w, r)
			return
		}

		token := headerParts[1]

		v := validator.New()

		if acttokens.ValidateTokenPlaintext(v, token); !v.Valid() {
			app.err.invalidAuthenticationTokenResponse(w, r)
			return
		}

		// find user by his token
		user, err := app.userService.GetByToken(acttokens.ScopeAuth, token)

		if err != nil {
			switch {
			case errors.Is(err, models.ErrRecordNotFound):
				app.err.invalidAuthenticationTokenResponse(w, r)
			default:
				app.err.serverErrorResponse(w, r, err)
			}
			return
		}

		// set user in context
		r = context.ContextSetUser(r, user)
		next.ServeHTTP(w, r)
	})
}

func (app *app) requireAuthenticatedUser(next http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := context.ContextGetUser(r)

		if user.IsAnon() {
			app.err.authenticationRequiredResponse(w, r)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (app *app) requireActivatedUser(next http.Handler) http.HandlerFunc {
	fn := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := context.ContextGetUser(r)

		if !user.Activated {
			app.err.inactiveAccountResponse(w, r)
			return
		}

		next.ServeHTTP(w, r)
	})

	return app.requireAuthenticatedUser(fn)
}

func (app *app) requirePermission(code string, next http.HandlerFunc) http.HandlerFunc {
	fn := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := context.ContextGetUser(r)

		// get user permissions
		permissions, err := app.permissionsService.GetAllForUser(user.Id)
		if err != nil {
			app.err.serverErrorResponse(w, r, err)
			return
		}

		// check if req permission includes in user permissions
		if !permissions.IsInclude(code) {
			app.err.notPermittedResponse(w, r)
			return
		}

		next.ServeHTTP(w, r)
	})

	return app.requireActivatedUser(fn)
}

func (app *app) enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Origin")
		w.Header().Add("Vary", "Access-Control-Request-Method")

		origin := r.Header.Get("Origin")

		if origin != "" {
			for i := range app.config.cors.trustedOrigins {

				if origin == app.config.cors.trustedOrigins[i] {
					w.Header().Set("Access-Control-Allow-Origin", "*")

					// check if request is preflight request
					if r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != "" {
						w.Header().Set("Access-Control-Allow-Methods", "OPTIONS, PUT, PATCH, DELETE")
						w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
						w.WriteHeader(http.StatusOK)
						return
					}
					break
				}

			}
		}

		next.ServeHTTP(w, r)
	})
}
