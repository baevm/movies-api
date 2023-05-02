package main

import (
	"movies-api/internal/utils"
	"net/http"
)

func (app *app) healthcheckHandler(w http.ResponseWriter, r *http.Request) {
	data := utils.Envelope{
		"status": "avaliable",
		"system_info": map[string]string{
			"env":     app.config.env,
			"version": version,
		},
	}

	err := utils.WriteJSON(w, http.StatusOK, data, nil)

	if err != nil {
		app.err.serverErrorResponse(w, r, err)
	}
}
