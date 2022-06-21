package main

import (
	"listener/event"
	"log"
	"math"
	"os"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	// Try to connect to rabbitmq
	rabbitConn, err := connect()
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	defer rabbitConn.Close()

	// Create Consumer
	consumer, err := event.NewConsumer(rabbitConn)
	if err != nil {
		panic(err)
	}

	log.Println("Listening and consuming RabbitMQ messages...")

	// watch the queue and consume events
	err = consumer.Listen([]string{"log.INFO", "log.WARNING", "log.ERROR"})
	if err != nil {
		log.Println(err)
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
