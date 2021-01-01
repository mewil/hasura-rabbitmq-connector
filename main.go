package main

import (
	"log"
	"os"
)

func main() {
	schema := os.Getenv("SCHEMA")
	hasuraClient := NewDefaultHasuraClient(os.Getenv("HASURA_ADDRESS"), schema)
	queue, err := NewRabbitMQQueue(os.Getenv("QUEUE_ADDRESS"))
	if err != nil {
		log.Fatal(err)
	}
	webhookCreator := NewWebhookCreator(os.Getenv("SERVER_ADDRESS"))

	tables, err := hasuraClient.ListTables()
	if err != nil {
		log.Fatal(err)
	}

	err = queue.CreateExchange(schema)
	if err != nil {
		log.Fatal(err)
	}

	for _, table := range tables {
		webhookUrl := webhookCreator.CreateWebhook(queue)
		err = hasuraClient.CreateEventTrigger(schema, table, webhookUrl)
		if err != nil {
			log.Fatal(err)
		}
	}

	log.Fatal(webhookCreator.ServeWebhooks())
}
