package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/streadway/amqp"
	"log"
	"os"
	"time"
)

type Message struct {
	Type   string `json:"type"`
	Action string `json:"action"`
	Params struct {
		Method string `json:"method"`
		URI    string `json:"uri"`
		Body   string `json:"body"`
	} `json:"params"`
	Context struct {
		MessageUUID string `json:"message_uuid"`
		ReplyTo     string `json:"reply_to"`
	} `json:"context"`
}

type Result struct {
	Type    string                 `json:"type"`
	Params  map[string]interface{} `json,mapstructure:"context"`
	Context map[string]interface{} `json,mapstructure:"context"`
}

type Log struct {
	Timestamp time.Time `json:"timestamp"`
	Lines     string    `json:"lines"`
}

type Task struct {
	Id             uint      `json:"id"`
	IdempotencyKey string    `json:"idempotency_key"`
	Name           string    `json:"name"`
	Timeout        uint      `json:"timeout"`
	ETA            uint      `json:"eta"`
	CollectLogs    bool      `json:"collect_logs"`
	Created        time.Time `json:"created"`
	Ended          time.Time `json:"ended"`
	Status         string    `json:"status"`
	Logs           []Log     `json:"logs"`
	Result         string    `json:"result"`
	Error          struct {
		Code    string `json:"result"`
		Reason  string `json:"reason"`
		Message string `json:"message"`
	} `json:"error"`
}

const ActionKubeApiCall = "kubernetes_api_call"

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func main() {
	var skipVerify bool

	if os.Getenv("KUBE_AGENT_SKIP_VERIFY") != "" {
		skipVerify = true
	}

	cfg := &tls.Config{InsecureSkipVerify: skipVerify}

	username := os.Getenv("KUBE_AGENT_NODE_UUID")
	password := os.Getenv("KUBE_AGENT_NODE_TOKEN")
	host := os.Getenv("KUBE_AGENT_SERVER_HOST")
	port := os.Getenv("KUBE_AGENT_SERVER_PORT")
	url := fmt.Sprintf("amqps://%s:%s@%s:%s/", username, password, host, port)

	conn, err := amqp.DialTLS(url, cfg)

	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	messages, err := ch.Consume(
		os.Getenv("KUBE_AGENT_QUEUE_NAME"),
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	failOnError(err, "Failed to register a consumer")

	forever := make(chan bool)

	go func() {
		for d := range messages {
			log.Printf(" [x] %s", d.Body)

			var msg Message

			if err := json.Unmarshal(
				d.Body, &msg); err != nil {
				panic(err)
			}

			if msg.Action == ActionKubeApiCall {
			}

			fmt.Printf("%+v\n", msg)
		}
	}()

	log.Printf(" Listening for messages")
	<-forever
}
