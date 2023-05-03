package main

import (
	"fmt"
	"net/http"
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
