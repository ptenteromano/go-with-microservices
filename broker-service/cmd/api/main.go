package main

import (
	"fmt"
	"log"
	"net/http"
)

const webPort = "80"

type Config struct{}

// Accept and respond to HTTP requests.
func main() {
	app := Config{}

	log.Printf("Starting broker on service on port %s\n", webPort)

	// define the http server
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", webPort),
		Handler: app.routes(),
	}

	// Start the server
	err := srv.ListenAndServe()

	if err != nil {
		log.Panic(err)
	}
}