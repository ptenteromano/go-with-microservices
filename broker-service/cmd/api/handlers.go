package main

import (
	"net/http"
)

func (app *Config) Broker(w http.ResponseWriter, r *http.Request) {
	payload := jsonResponse{
		Error:   false,
		Message: "Hit the broker on port",
		Port:    webPort,
	}

	_ = app.writeJSON(w, http.StatusOK, payload)
}
