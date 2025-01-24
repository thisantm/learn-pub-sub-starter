package main

import (
	"fmt"
	"log"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	fmt.Println("Starting Peril client...")
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

	username, err := gamelogic.ClientWelcome()
	if err != nil {
		log.Fatal(err)
	}
	gameState := gamelogic.NewGameState(username)

	err = pubsub.SubscribeJSON(
		amqpConnection,
		routing.ExchangePerilDirect,
		routing.PauseKey+"."+gameState.GetUsername(),
		routing.PauseKey,
		int(amqp.Transient),
		handlerPause(gameState),
	)
	if err != nil {
		log.Fatal(err)
	}

	err = pubsub.SubscribeJSON(
		amqpConnection,
		routing.ExchangePerilTopic,
		routing.ArmyMovesPrefix+"."+gameState.GetUsername(),
		routing.ArmyMovesPrefix+".*",
		int(amqp.Transient),
		handlerMove(gameState),
	)
	if err != nil {
		log.Fatal(err)
	}

	for {
		input := gamelogic.GetInput()
		if len(input) == 0 {
			continue
		}

		switch input[0] {
		case "spawn":
			err := gameState.CommandSpawn(input)
			if err != nil {
				fmt.Println(err)
				continue
			}

		case "move":
			armyMove, err := gameState.CommandMove(input)
			if err != nil {
				fmt.Println(err)
				continue
			}
			err = pubsub.PublishJSON(amqpChannel, routing.ExchangePerilTopic, routing.ArmyMovesPrefix+"."+armyMove.Player.Username, armyMove)
			if err != nil {
				fmt.Println(err)
				continue
			}
			fmt.Printf("Moved %v units to %s\n", len(armyMove.Units), armyMove.ToLocation)

		case "status":
			gameState.CommandStatus()

		case "help":
			gamelogic.PrintClientHelp()

		case "spam":
			fmt.Println("Spamming not allowed yet!")

		case "quit":
			gamelogic.PrintQuit()
			return

		default:
			fmt.Println("Unknown command")
		}
	}
}
