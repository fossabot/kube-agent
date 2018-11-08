package main

import (
	"fmt"
	"github.com/wodby/kube-agent/pkg/rabbitmq"
	"log"
	"os"
)

func main() {
	var skipVerify bool

	if os.Getenv("KUBE_AGENT_SKIP_VERIFY") != "" {
		skipVerify = true
	}

	username := os.Getenv("KUBE_AGENT_NAME")
	password := os.Getenv("KUBE_AGENT_NODE_TOKEN")
	host := os.Getenv("KUBE_AGENT_SERVER_HOST")
	port := os.Getenv("KUBE_AGENT_SERVER_PORT")

	if port == "" {
		port = "443"
	}

	if username == "" || password == "" || host == "" {
		log.Fatalf("Missing required parameter")
	}

	client := rabbitmq.Client{
		Username:   username,
		Password:   password,
		Host:       host,
		Port:       port,
		SkipVerify: skipVerify,
	}

	fmt.Printf("%+v\n", client)

	queue := os.Getenv("KUBE_AGENT_INBOUND")
	if queue == "" {
		log.Fatalf("Queue must be specified")
	}

	err := client.Consume(queue)
	if err != nil {
		log.Fatalf("%s", err)
	}
}
