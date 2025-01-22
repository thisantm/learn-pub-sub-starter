package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	fmt.Println("Starting Peril server...")

	const connectionString string = "amqp://guest:guest@localhost:5672/"

	amqpConnection, err := amqp.Dial(connectionString)
	if err != nil {
		log.Fatal(err)
	}
	defer amqpConnection.Close()
	log.Println("Connection succeeded to RabbitMQ")

	amqpChannel, err := amqpConnection.Channel()
	if err != nil {
		log.Fatal(err)
	}
	defer amqpChannel.Close()

	playingState, err := json.Marshal(routing.PlayingState{
		IsPaused: true,
	})
	if err != nil {
		log.Fatal(err)
	}

	err = pubsub.PublishJSON(amqpChannel, routing.ExchangePerilDirect, routing.PauseKey, playingState)
	if err != nil {
		log.Fatal(err)
	}

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan

	log.Println("Ending connection")
}
