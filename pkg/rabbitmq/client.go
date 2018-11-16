package rabbitmq

import (
	"context"
	"crypto/tls"
	"errors"
	"github.com/streadway/amqp"
	"log"
	"os"
	"time"
)

// Client represents a connection to a specific queue.
type Client struct {
	exchange        string
	queue           string
	url             string
	logger          *log.Logger
	connection      *amqp.Connection
	channelIn       *amqp.Channel
	channelOut      *amqp.Channel
	done            chan bool
	notifyClose     chan *amqp.Error
	notifyConfirm   chan amqp.Confirmation
	inChannelReady  chan bool
	outChannelReady chan bool
	isConnected     bool
	skipVerify      bool
	context         context.Context
}

const (
	// When reconnecting to the server after connection failure
	reconnectDelay = 5 * time.Second

	// When resending messages the server didn't confirm
	resendDelay = 5 * time.Second
)

var (
	errNotConnected = errors.New("not connected to the queue")
	//errNotConfirmed  = errors.New("message not confirmed")
	errAlreadyClosed = errors.New("already closed: not connected to the queue")
)

func NewClient(ctx context.Context, exchangeName string, queueName string, addr string, skip bool) *Client {
	client := Client{
		logger:          log.New(os.Stdout, "", log.LstdFlags),
		exchange:        exchangeName,
		queue:           queueName,
		done:            make(chan bool),
		inChannelReady:  make(chan bool, 1),
		outChannelReady: make(chan bool, 1),
		skipVerify:      skip,
		url:             addr,
		context:         ctx,
	}

	return &client
}

// handleReconnect will wait for a connection error on
// notifyClose, and then continuously attempt to reconnect.
func (client *Client) Connect() {
	for {
		client.isConnected = false
		log.Println("Attempting to connect")
		for !client.connect() {
			log.Println("Failed to connect. Retrying...")
			time.Sleep(reconnectDelay)
		}
		select {
		case <-client.done:
			return
		case <-client.notifyClose:
		case <-client.context.Done():

			return
		}
	}
}

// connect will make a single attempt to connect to
// RabbitMQ. It returns the success of the attempt.
func (client *Client) connect() bool {
	cfg := &tls.Config{InsecureSkipVerify: client.skipVerify}

	conn, err := amqp.DialTLS(client.url, cfg)
	if err != nil {
		return false
	}
	chOut, err := conn.Channel()
	if err != nil {
		return false
	}
	chOut.Confirm(false)
	chIn, err := conn.Channel()
	if err != nil {
		return false
	}

	client.changeConnection(conn, chOut, chIn)
	client.isConnected = true
	log.Println("Connected!")
	client.inChannelReady <- true
	client.outChannelReady <- true
	return true
}

// changeConnection takes a new connection to the queue,
// and updates the channel listeners to reflect this.
func (client *Client) changeConnection(connection *amqp.Connection, channelIn *amqp.Channel, channelOut *amqp.Channel) {
	client.connection = connection
	client.channelIn = channelIn
	client.channelOut = channelOut

	client.notifyClose = make(chan *amqp.Error)
	client.channelIn.NotifyClose(client.notifyClose)
	client.channelOut.NotifyClose(client.notifyClose)

	client.notifyConfirm = make(chan amqp.Confirmation)
	client.channelIn.NotifyPublish(client.notifyConfirm)
}

// Push will push data onto the queue, and wait for a confirm.
// If no confirms are received until within the resendTimeout,
// it continuously re-sends messages until a confirm is received.
// This will block until the server sends a confirm. Errors are
// only returned if the push action itself fails, see UnsafePush.
func (client *Client) Push(routingKey string, publishing amqp.Publishing) error {
	<-client.outChannelReady

	if !client.isConnected {
		return errors.New("failed to push push: not connected")
	}
	for {
		err := client.unsafePush(routingKey, publishing)
		if err != nil {
			client.logger.Println("Push failed. Retrying...")
			continue
		}
		select {
		case <-client.context.Done():
			client.Close()
			return nil
		case confirm := <-client.notifyConfirm:
			if confirm.Ack {
				client.logger.Println("Push confirmed!")
				return nil
			}
		case <-time.After(resendDelay):
		}
		client.logger.Println("Push didn't confirm. Retrying...")
	}
}

// UnsafePush will push to the queue without checking for
// confirmation. It returns an error if it fails to connect.
// No guarantees are provided for whether the server will
// receive the message.
func (client *Client) unsafePush(routingKey string, publishing amqp.Publishing) error {
	if !client.isConnected {
		return errNotConnected
	}
	return client.channelOut.Publish(
		client.exchange,
		routingKey,
		true,
		false,
		publishing,
	)
}

// Listen will continuously put queue items on the channel.
// It is required to call delivery.Ack when it has been
// successfully processed, or delivery.Nack when it fails.
// Ignoring this will cause data to build up on the server.
func (client *Client) Listen() (<-chan amqp.Delivery, error) {
	<-client.inChannelReady

	deliveries, err := client.channelIn.Consume(
		client.queue,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, err
	}

	return deliveries, nil
}

// Close will cleanly shutdown the channel and connection.
func (client *Client) Close() error {
	if !client.isConnected {
		return errAlreadyClosed
	}
	err := client.channelIn.Close()
	if err != nil {
		return err
	}
	err = client.channelOut.Close()
	if err != nil {
		return err
	}
	err = client.connection.Close()
	if err != nil {
		return err
	}
	close(client.done)
	close(client.inChannelReady)
	close(client.outChannelReady)
	client.isConnected = false
	return nil
}
