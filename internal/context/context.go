package context

import (
	"context"
	"movies-api/internal/models/users"
	"net/http"
)

type contextKey string

const userContextKey = contextKey("user")

func ContextSetUser(r *http.Request, user *users.User) *http.Request {
	ctx := context.WithValue(r.Context(), userContextKey, user)
	return r.WithContext(ctx)
}

func ContextGetUser(r *http.Request) *users.User {
	user, ok := r.Context().Value(userContextKey).(*users.User)

	if !ok {
		panic("missing user value in context")
	}

	return user
}
