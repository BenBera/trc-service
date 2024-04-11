package library

import (
	"context"
	"encoding/json"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
)

func Publish(conn *amqp.Connection, name string, payload interface{}, priority uint8) error {

	js, _ := json.Marshal(payload)
	log.Printf("publish to %s | %s", name, string(js))

	queue := name
	exchange := name
	key := name
	exchangeType := "direct"
	ch, err := conn.Channel()
	if err != nil {

		log.Printf(" got error opening rabbitMQ channel %s ", err.Error())
		return err
	}

	defer ch.Close()

	err = ch.ExchangeDeclare(
		queue,        // name
		exchangeType, // type
		true,         // durable
		false,        // auto-deleted
		false,        // internal
		false,        // no-wait
		nil,          // arguments
	)

	if err != nil {

		log.Printf(" got error Failed to declare a queue %s error %s ", name, err.Error())
		return err
	}
	message, err := json.Marshal(payload)

	if err != nil {

		log.Printf(" got error decoding payload to string %s ", err.Error())
		return err
	}

	err = ch.PublishWithContext(
		context.TODO(),
		exchange, // exchange
		key,      // routing key
		false,    // mandatory
		false,    // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        message,
			Priority:    priority,
		})

	if err != nil {

		log.Printf(" got error publishing message %s error %s ", message, err.Error())
		return err
	}

	return nil
}
