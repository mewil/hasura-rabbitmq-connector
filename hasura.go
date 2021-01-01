package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

const (
	jsonContentType                   = "application/json"
	listTablesRequest                 = `{ "type": "run_sql", "args": { "sql": "SELECT * FROM pg_catalog.pg_tables;" } }`
	createEventTriggerRequestTemplate = `{ "type" : "create_event_trigger", "args" : { "name": "%s", "table": { "schema": "%s", "name": "%s" }, "webhook": "%s", "insert": { "columns": "*", "payload": "*" }, "update": { "columns": "*", "payload": "*" }, "delete": { "columns": "*", "payload": "*" } }}`
	deleteEventTriggerRequestTemplate = `{ "type" : "delete_event_trigger", "args" : { "name": "%s" }}`
	listTablesResultLen               = 8
)

type EventPayload struct {
	Event Event `json:"event"`
	Table Table `json:"table"`
}

type Table struct {
	Schema string `json:"schema"`
	Name   string `json:"name"`
}

type Event struct {
	Data json.RawMessage `json:"data"`
	Op   string          `json:"op"`
}

type HasuraClient interface {
	ListTables() ([]string, error)
	CreateEventTrigger(schema, table, webhookUrl string) error
}

type DefaultHasuraClient struct {
	baseUrl, schema string
}

func NewDefaultHasuraClient(baseUrl string, schema string) HasuraClient {
	return &DefaultHasuraClient{baseUrl: baseUrl, schema: schema}
}

func (c *DefaultHasuraClient) post(body io.Reader) ([]byte, error) {
	res, err := http.Post(fmt.Sprintf("%s/v1/query", c.baseUrl), jsonContentType, body)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		data, _ := ioutil.ReadAll(res.Body)
		return nil, fmt.Errorf("hasura returned status %d: %s", res.StatusCode, string(data))
	}
	return ioutil.ReadAll(res.Body)
}

func (c *DefaultHasuraClient) ListTables() ([]string, error) {
	data, err := c.post(strings.NewReader(listTablesRequest))
	if err != nil {
		return nil, err
	}
	results := &struct{ Result [][]string }{}
	err = json.Unmarshal(data, results)
	if err != nil {
		return nil, err
	}
	tables := make([]string, 0)
	for _, result := range results.Result {
		if len(result) != listTablesResultLen {
			return nil, fmt.Errorf("")
		}
		if result[0] != c.schema {
			continue
		}
		tables = append(tables, result[1])
	}
	log.Println("found", len(tables), "tables:", tables)
	return tables, nil
}

func (c *DefaultHasuraClient) deleteEventTrigger(triggerName string) error {
	body := fmt.Sprintf(deleteEventTriggerRequestTemplate, triggerName)
	_, err := c.post(strings.NewReader(body))
	return err
}

func (c *DefaultHasuraClient) CreateEventTrigger(schema, table, webhookUrl string) error {
	triggerName := fmt.Sprintf("%s_%s_trigger", schema, table)
	if err := c.deleteEventTrigger(triggerName); err != nil {
		return err
	}
	body := fmt.Sprintf(createEventTriggerRequestTemplate, triggerName, schema, table, webhookUrl)
	_, err := c.post(strings.NewReader(body))
	if err != nil {
		return err
	}
	log.Println("created event trigger", triggerName)
	return nil
}
