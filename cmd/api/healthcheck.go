package main

import (
	"net/http"
)

func (app *app) healthcheckHandler(w http.ResponseWriter, r *http.Request) {
	data := envelope{
		"status": "avaliable",
		"system_info": map[string]string{
			"env":     app.config.env,
			"version": version,
		},
	}

	err := writeJSON(w, http.StatusOK, data, nil)

	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
