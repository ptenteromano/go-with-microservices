package main

import (
	"broker/event"
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/ptenteromano/jsontools"
)

// Primary json entrypoint. Need an action and a nested json structure
type requestPayload struct {
	Action string      `json:"action"`
	Auth   AuthPayload `json:"auth,omitempty"`
	Log    LogPayload  `json:"log,omitempty"`
	Mail   MailPayload `json:"mail,omitempty"`
}

type AuthPayload struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LogPayload struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

type MailPayload struct {
	From    string `json:"from"`
	To      string `json:"to"`
	Subject string `json:"subject"`
	Message string `json:"message"`
}

var jtools jsontools.Tools

func (app *Config) Broker(w http.ResponseWriter, r *http.Request) {
	payload := jsontools.JsonResponse{
		Error:   false,
		Message: "Hit the broker!",
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
	case "log":
		app.logEventViaRabbit(w, requestPayload.Log)
	case "mail":
		app.SendMail(w, requestPayload.Mail)
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
		log.Println("error in brokerApp: ", err)
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
		jtools.ErrorJSON(w, errors.New("error calling auth service"))
		return
	}

	var jsonFromService jsontools.JsonResponse
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
	payload := jsontools.JsonResponse{
		Error:   false,
		Message: "authenticated successfully",
		Data:    jsonFromService.Data,
	}

	jtools.WriteJSON(w, http.StatusAccepted, payload)
}

// Point to the Logger-Service to post a log entry
// Deprecated (this uses REST)
func (app *Config) logItem(w http.ResponseWriter, entry LogPayload) {
	jsonData, _ := json.MarshalIndent(entry, "", "\t")

	request, err := http.NewRequest("POST", "http://logger-service/log", bytes.NewBuffer(jsonData))
	if err != nil {
		jtools.ErrorJSON(w, err)
		return
	}

	request.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		jtools.ErrorJSON(w, err)
		return
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusAccepted {
		jtools.ErrorJSON(w, err)
		return
	}

	payload := jsontools.JsonResponse{
		Error:   false,
		Message: "logged successfully",
	}

	jtools.WriteJSON(w, http.StatusAccepted, payload)
}

func (app *Config) SendMail(w http.ResponseWriter, msg MailPayload) {
	jsonData, _ := json.MarshalIndent(msg, "", "\t")

	// Call the mail service
	request, err := http.NewRequest("POST", "http://mailer-service/send", bytes.NewBuffer(jsonData))
	if err != nil {
		jtools.ErrorJSON(w, err)
		return
	}

	request.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		jtools.ErrorJSON(w, err)
		return
	}
	defer response.Body.Close()

	log.Println("response from mailer: ", response.StatusCode)

	if response.StatusCode != http.StatusAccepted {
		jtools.ErrorJSON(w, errors.New("error calling mail service from broker"))
		return
	}

	payload := jsontools.JsonResponse{
		Error:   false,
		Message: "Message sent to " + msg.To,
	}

	jtools.WriteJSON(w, http.StatusAccepted, payload)
}

// Push event into RabbitMQ
func (app *Config) logEventViaRabbit(w http.ResponseWriter, l LogPayload) {
	err := app.pushToQueue(l.Name, l.Data)
	if err != nil {
		jtools.ErrorJSON(w, err)
		return
	}

	payload := jsontools.JsonResponse{
		Error:   false,
		Message: "Logged via RabbitMQ",
	}

	jtools.WriteJSON(w, http.StatusAccepted, payload)
}

func (app *Config) pushToQueue(name, msg string) error {
	emitter, err := event.NewEventEmitter(app.Rabbit)
	if err != nil {
		return err
	}

	payload := LogPayload{
		Name: name,
		Data: msg,
	}

	j, _ := json.MarshalIndent(&payload, "", "\t")
	err = emitter.Push(string(j), "log.INFO")
	if err != nil {
		return err
	}

	return nil
}
