package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
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

	gamelogic.PrintServerHelp()
	for {
		input := gamelogic.GetInput()
		if len(input) == 0 {
			continue
		}

		switch input[0] {
		case "pause":
			log.Println("Sending pause message...")
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
		case "resume":
			log.Println("Sending resume message...")
			playingState, err := json.Marshal(routing.PlayingState{
				IsPaused: false,
			})
			if err != nil {
				log.Fatal(err)
			}
			err = pubsub.PublishJSON(amqpChannel, routing.ExchangePerilDirect, routing.PauseKey, playingState)
			if err != nil {
				log.Fatal(err)
			}
		case "quit":
			log.Println("Ending connection...")
			return

		default:
			log.Println("Unknown command")
		}
	}
}
