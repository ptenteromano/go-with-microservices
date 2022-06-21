package main

import (
	"log"
	"net/http"

	"github.com/ptenteromano/jsontools"
)

var jtools jsontools.Tools

func (app *Config) SendMail(w http.ResponseWriter, r *http.Request) {
	type mailMessage struct {
		From    string `json:"from"`
		To      string `json:"to"`
		Subject string `json:"subject"`
		Message string `json:"message"`
	}

	var requestPayload mailMessage
	log.Println("Receiving data!")
	err := jtools.ReadJSON(w, r, &requestPayload)
	if err != nil {
		jtools.ErrorJSON(w, err)
		return
	}

	msg := Message{
		From:    requestPayload.From,
		To:      requestPayload.To,
		Subject: requestPayload.Subject,
		Data:    requestPayload.Message,
	}

	err = app.Mailer.SendSMTPMessage(msg)
	if err != nil {
		log.Println("Bad happened during smtp: ", err)
		jtools.ErrorJSON(w, err)
		return
	}

	payload := jsontools.JsonResponse{
		Error:   false,
		Message: "sent to " + requestPayload.To,
	}

	jtools.WriteJSON(w, http.StatusAccepted, payload)
}
