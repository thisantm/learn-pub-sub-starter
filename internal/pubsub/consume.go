package pubsub

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

type AckType int

const (
	Ack AckType = iota
	NackRequeue
	NackDiscard
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
		amqp.Table{
			"x-dead-letter-exchange": routing.ExchangePerilDlx,
		},
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
	handler func(T) AckType,
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
			switch handler(target) {
			case Ack:
				msg.Ack(false)
				fmt.Println("Ack")
			case NackDiscard:
				msg.Nack(false, false)
				fmt.Println("NackDiscard")
			case NackRequeue:
				msg.Nack(false, true)
				fmt.Println("NackRequeue")
			}
		}
	}()

	return nil
}

func SubscribeGob[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	simpleQueueType int,
	handler func(T) AckType,
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
			if msg.ContentType != "application/gob" {
				fmt.Printf("unexpected content type: %s\n", msg.ContentType)
				msg.Nack(false, false)
				continue
			}
			buffer := bytes.NewBuffer(msg.Body)
			gobDec := gob.NewDecoder(buffer)

			var target T
			err := gobDec.Decode(&target)
			if err != nil {
				fmt.Printf("could not decode message: %v\n", err)
				continue
			}
			switch handler(target) {
			case Ack:
				msg.Ack(false)
				fmt.Println("Ack")
			case NackDiscard:
				msg.Nack(false, false)
				fmt.Println("NackDiscard")
			case NackRequeue:
				msg.Nack(false, true)
				fmt.Println("NackRequeue")
			}
		}
	}()

	return nil
}
