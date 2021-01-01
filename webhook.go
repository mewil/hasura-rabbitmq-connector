package main

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

type WebhookCreator interface {
	CreateWebhook(queue Queue) string
	ServeWebhooks() error
}

type DefaultWebhookCreator struct {
	mux     *http.ServeMux
	address string
}

func NewWebhookCreator(address string) WebhookCreator {
	return &DefaultWebhookCreator{mux: http.NewServeMux(), address: address}
}

func (w DefaultWebhookCreator) CreateWebhook(queue Queue) string {
	path := fmt.Sprintf("/%v", uuid.New())
	w.mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		payload := &EventPayload{}
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		err = json.Unmarshal(body, payload)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		err = queue.PublishMessage(payload.Table.Schema, payload.Table.Name, strings.ToLower(payload.Event.Op), body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})
	log.Println("created webhook at path", path)
	return fmt.Sprint("http://", w.address, path)
}

func (w DefaultWebhookCreator) ServeWebhooks() error {
	log.Println("serving webhooks...")
	return http.ListenAndServe(w.address, w.mux)
}
