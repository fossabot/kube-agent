package main

import (
	"context"
	"fmt"
	"github.com/wodby/kube-agent/internal/app/kubeagent"
	"github.com/wodby/kube-agent/pkg/rabbitmq"
	"log"
	"os"
	"os/signal"
	"syscall"
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

	if username == "" || password == "" || host == "" {
		log.Fatalln("One of the following parameters is missing: username, password, host")
	}

	if port == "" {
		port = "443"
	}

	addr := fmt.Sprintf("amqps://%s:%s@%s:%s/", username, password, host, port)

	queue := os.Getenv("KUBE_AGENT_QUEUE")
	if queue == "" {
		log.Fatalln("Queue must be specified")
	}

	exchange := os.Getenv("KUBE_AGENT_EXCHANGE")
	if exchange == "" {
		log.Fatalln("Exchange must be specified")
	}

	ctx, cancel := context.WithCancel(context.Background())
	var gracefulStop = make(chan os.Signal)
	signal.Notify(gracefulStop, syscall.SIGTERM)
	signal.Notify(gracefulStop, syscall.SIGINT)

	go func() {
		select {
		case sig := <-gracefulStop:
			fmt.Printf("caught sig: %+v", sig)
			cancel()
		case <-ctx.Done():
		}
	}()

	client := rabbitmq.NewClient(ctx, exchange, queue, addr, skipVerify)

	go client.Connect()

	agent := kubeagent.NewAgent(ctx, client)

	go agent.Consume()
	go agent.Process()
	go agent.Respond()

	<-ctx.Done()
}
