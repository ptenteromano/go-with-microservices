package main

import (
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

const webPort = "80"

type Config struct {
	Rabbit *amqp.Connection
}

// Accept and respond to HTTP requests.
func main() {
	// Try to connect to rabbitmq
	rabbitConn, err := connect()
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	defer rabbitConn.Close()

	app := Config{
		Rabbit: rabbitConn,
	}

	log.Printf("Starting broker on service on port %s\n", webPort)

	// define the http server
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", webPort),
		Handler: app.routes(),
	}

	// Start the server
	err = srv.ListenAndServe()

	if err != nil {
		log.Panic(err)
	}
}

// RabbitMQ connect function
func connect() (*amqp.Connection, error) {
	var counts int64
	var connection *amqp.Connection

	// Dont continue until rabbit is ready
	for {
		c, err := amqp.Dial("amqp://guest:guest@rabbitmq")
		if err != nil {
			log.Println("RabbitMQ not yet ready...")
			counts++
		} else {
			log.Println("Succesfully connected to RabbitMQ!")
			connection = c
			break
		}

		if counts > 5 {
			log.Println(err)
			return nil, err
		}

		// Implement exponential backoff
		backoff := time.Duration(math.Pow(float64(counts), 2)) * time.Second
		log.Println("Backing off from RabbitMQ initialization...")
		time.Sleep(backoff)
	}

	return connection, nil
}
