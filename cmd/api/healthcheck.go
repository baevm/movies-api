package main

import (
	"fmt"
	"net/http"
)

func (app *app) healthcheckHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "status: avaliable")
	fmt.Fprintf(w, "env: %s\n", app.config.env)
	fmt.Fprintf(w, "version: %s\n", version)
}
