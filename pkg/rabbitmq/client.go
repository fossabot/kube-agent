package rabbitmq

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/streadway/amqp"
	"github.com/wodby/kube-agent/pkg/kubernetes"
	"log"
)

type consumer func(amqp.Delivery)

type Client struct {
	Username   string
	Password   string
	Host       string
	Port       string
	SkipVerify bool
	conn       *amqp.Connection
}

func (c *Client) connect() (*amqp.Connection, error) {
	if c.conn == nil {
		var err error

		cfg := &tls.Config{InsecureSkipVerify: c.SkipVerify}
		url := fmt.Sprintf("amqps://%s:%s@%s:%s/", c.Username, c.Password, c.Host, c.Port)
		fmt.Println(url)
		c.conn, err = amqp.DialTLS(url, cfg)

		if err != nil {
			return nil, err
		}
	}

	return c.conn, nil
}

func (c *Client) Consume(queue string) error {
	conn, err := c.connect()
	if err != nil {
		return err
	}

	fmt.Println("Opening channel")

	ch, err := conn.Channel()
	if err != nil {
		return err
	}

	defer ch.Close()

	fmt.Println("Start listening")

	messages, err := ch.Consume(
		queue,
		"",
		true,
		false,
		false,
		false,
		nil,
	)

	forever := make(chan bool)

	// set events non-persistent
	// result persistent

	go func() {
		fmt.Println("Listening")

		for d := range messages {
			switch d.Type {
			//case MsgTypePing:
			//	r := Response{Succeed: true}
			//	break

			case MsgTypeKubeApiRequest:
				var msg KubeApiRequest
				err = json.Unmarshal(d.Body, &msg)
				kubernetes.ApiRequest(msg.URI, msg.Method, msg.Body)
				break

			case MsgTypeStreamResourceLogs:
				var msg StreamResourceLogs
				err = json.Unmarshal(d.Body, &msg)
				break

			case MsgTypeTaskKubeDeploy:
				var msg TaskKubeDeploy
				err = json.Unmarshal(d.Body, &msg)
				break

			case MsgTypeTaskKubeRunJob:
				var msg TaskKubeRunJob
				err = json.Unmarshal(d.Body, &msg)
				break

			case MsgTypeTaskGet:
				var msg TaskGet
				err = json.Unmarshal(d.Body, &msg)
				break

			case MsgTypeTaskStreamLogs:
				var msg TaskStreamLogs
				err = json.Unmarshal(d.Body, &msg)
				break
			}
		}
	}()

	log.Printf(" Listening for messages")
	<-forever

	return nil
}

func (c *Client) publish(exchange string, routingKey string, json string) error {
	conn, err := c.connect()
	if err != nil {
		return err
	}

	ch, err := conn.Channel()
	if err != nil {
		return err
	}

	defer ch.Close()

	if err = ch.Publish(
		exchange,
		routingKey,
		false,
		false,
		amqp.Publishing{
			Headers:         amqp.Table{},
			ContentType:     "text/json",
			ContentEncoding: "",
			Body:            []byte(json),
			DeliveryMode:    amqp.Transient,
		},
	); err != nil {
		return err
	}

	return nil
}
