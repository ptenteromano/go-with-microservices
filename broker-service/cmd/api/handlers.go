package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/ptenteromano/jsontools"
)

type requestPayload struct {
	Action string      `json:"action"`
	Auth   AuthPayload `json:"auth,omitempty"`
}

type AuthPayload struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type jsonResponse struct {
	Error   bool   `json:"error"`
	Message string `json:"message"`
	Port    string `json:"port"`
	Data    any    `json:"data,omitempty"`
}

var jtools jsontools.Tools

func (app *Config) Broker(w http.ResponseWriter, r *http.Request) {
	payload := jsonResponse{
		Error:   false,
		Message: "Hit the broker on port",
		Port:    webPort,
	}

	_ = jtools.WriteJSON(w, http.StatusOK, payload)
}

func (app *Config) HandleSubmission(w http.ResponseWriter, r *http.Request) {
	var requestPayload requestPayload

	err := jtools.ReadJSON(w, r, &requestPayload)
	if err != nil {
		jtools.ErrorJSON(w, err)
		return
	}

	switch requestPayload.Action {
	case "auth":
		app.authenticate(w, requestPayload.Auth)
	default:
		jtools.ErrorJSON(w, errors.New("unknown action"))
	}

}

// Point to the Authentication-Service to authenticate
func (app *Config) authenticate(w http.ResponseWriter, a AuthPayload) {
	// Create json and send to auth microservice
	jsonData, _ := json.MarshalIndent(a, "", "\t")

	// call the Auth Service
	request, err := http.NewRequest("POST", "http://authentication-service/authenticate", bytes.NewBuffer(jsonData))
	if err != nil {
		jtools.ErrorJSON(w, err)
		return
	}

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		jtools.ErrorJSON(w, err)
		return
	}

	defer response.Body.Close()

	// get back correct status code
	if response.StatusCode == http.StatusUnauthorized {
		jtools.ErrorJSON(w, errors.New("invalid credentials"))
		return
	}
	if response.StatusCode != http.StatusAccepted {
		log.Println("error: ", response.Status)
		jtools.ErrorJSON(w, errors.New("error calling auth service"))
		return
	}

	var jsonFromService jsonResponse
	err = json.NewDecoder(response.Body).Decode(&jsonFromService)

	if err != nil {
		jtools.ErrorJSON(w, err)
		return
	}
	if jsonFromService.Error {
		jtools.ErrorJSON(w, errors.New(jsonFromService.Message), http.StatusUnauthorized)
		return
	}

	// login is valid
	payload := jsonResponse{
		Error:   false,
		Message: "authenticated successfully",
		Data:    jsonFromService.Data,
	}

	jtools.WriteJSON(w, http.StatusAccepted, payload)
}
