package main

import (
	"fmt"
	"github.com/streadway/amqp"
	"log"
)

type Queue interface {
	CreateExchange(schema string) error
	PublishMessage(schema, table, action string, message []byte) error
}

type RabbitMQQueue struct {
	ch *amqp.Channel
}

func NewRabbitMQQueue(address string) (Queue, error) {
	conn, err := amqp.Dial(address)
	if err != nil {
		return nil, err
	}
	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}
	log.Println("connected to queue")
	return &RabbitMQQueue{ch: ch}, nil
}

func getTopicName(schema string) string {
	return fmt.Sprintf("%s_topic", schema)
}

func (q *RabbitMQQueue) CreateExchange(schema string) error {
	topicName := getTopicName(schema)
	log.Println("creating queue exchange", topicName)
	return q.ch.ExchangeDeclare(
		topicName, // name
		"topic",   // type
		true,      // durable
		false,     // auto-deleted
		false,     // internal
		false,     // no-wait
		nil,       // arguments
	)
}

func (q *RabbitMQQueue) PublishMessage(schema, table, action string, message []byte) error {
	return q.ch.Publish(
		getTopicName(schema),                           // name
		fmt.Sprintf("%s.%s.%s", schema, table, action), // routing key
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        message,
		})
}
