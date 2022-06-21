package main

import (
	"log"
	"log-service/data"
	"net/http"

	"github.com/ptenteromano/jsontools"
)

type JSONPayload struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

var jtools jsontools.Tools

func (app *Config) WriteLog(w http.ResponseWriter, r *http.Request) {
	var requestPayload JSONPayload

	_ = jtools.ReadJSON(w, r, &requestPayload)

	// insert data
	event := data.LogEntry{
		Name: requestPayload.Name,
		Data: requestPayload.Data,
	}
	log.Println("Inserting entry into logger...")
	err := app.Models.LogEntry.Insert(event)
	if err != nil {
		jtools.ErrorJSON(w, err)
	}

	resp := jsontools.JsonResponse{
		Error:   false,
		Message: "logged sucessfully",
	}

	jtools.WriteJSON(w, http.StatusAccepted, resp)
}
