package pubsub

import (
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

func DeclareAndBind(
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	simpleQueueType int,
) (*amqp.Channel, amqp.Queue, error) {
	amqpCh, err := conn.Channel()
	if err != nil {
		return nil, amqp.Queue{}, err
	}

	transientInt := int(amqp.Transient)
	amqpQueue, err := amqpCh.QueueDeclare(
		queueName,
		simpleQueueType != transientInt,
		simpleQueueType == transientInt,
		simpleQueueType == transientInt,
		false,
		nil,
	)
	if err != nil {
		return nil, amqp.Queue{}, err
	}

	err = amqpCh.QueueBind(queueName, key, exchange, false, nil)
	if err != nil {
		return nil, amqp.Queue{}, err
	}

	return amqpCh, amqpQueue, nil
}

func SubscribeJSON[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	simpleQueueType int,
	handler func(T),
) error {
	amqpCh, amqpQueue, err := DeclareAndBind(conn, exchange, queueName, key, simpleQueueType)
	if err != nil {
		return err
	}

	msgCh, err := amqpCh.Consume(
		amqpQueue.Name,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	go func() {
		defer amqpCh.Close()
		for msg := range msgCh {
			var target T
			err := json.Unmarshal(msg.Body, &target)
			if err != nil {
				fmt.Printf("could not unmarshal message: %v\n", err)
				continue
			}
			handler(target)
			msg.Ack(false)
		}
	}()

	return nil
}
