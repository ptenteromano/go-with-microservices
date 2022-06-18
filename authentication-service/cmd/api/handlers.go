package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/ptenteromano/jsontools"
)

type jsonResponse struct {
	Error   bool   `json:"error"`
	Message string `json:"message"`
	Port    string `json:"port"`
	Data    any    `json:"data,omitempty"`
}

var jtools jsontools.Tools

// Endpoint to auth the user through Postgres
func (app *Config) Authenticate(w http.ResponseWriter, r *http.Request) {
	var requestPayload struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	log.Println("Hitting /authenticate post route")
	err := jtools.ReadJSON(w, r, &requestPayload)
	if err != nil {
		jtools.ErrorJSON(w, err, http.StatusBadRequest)
		return
	}

	// validate the user against the database
	user, err := app.Models.User.GetByEmail(requestPayload.Email)
	if err != nil {
		jtools.ErrorJSON(w, errors.New("invalid credentials"), http.StatusBadRequest)
		return
	}

	valid, err := user.PasswordMatches(requestPayload.Password)
	if err != nil || !valid {
		jtools.ErrorJSON(w, errors.New("invalid credentials"), http.StatusBadRequest)
		return
	}

	payload := jsonResponse{
		Error:   false,
		Message: fmt.Sprintf("Logged in user %s", user.Email),
		Data:    user,
	}

	jtools.WriteJSON(w, http.StatusAccepted, payload)
}
