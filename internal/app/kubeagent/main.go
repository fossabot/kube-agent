package kubeagent

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/streadway/amqp"
	"github.com/wodby/kube-agent/pkg/kubernetes"
	"github.com/wodby/kube-agent/pkg/rabbitmq"
)

type ResponseMsg struct {
	publishing amqp.Publishing
	routingKey string
}

type Agent struct {
	client    *rabbitmq.Client
	context   context.Context
	responses chan ResponseMsg
}

func NewAgent(ctx context.Context, c *rabbitmq.Client) *Agent {
	agent := Agent{
		context:   ctx,
		client:    c,
		responses: make(chan ResponseMsg),
	}

	return &agent
}

func (agent *Agent) Consume() error {
	deliveries, err := agent.client.Listen()
	if err != nil {
		return err
	}

	for {
		select {
		case <-agent.context.Done():
			agent.client.Close()
		case d, ok := <-deliveries:
			if ok {
				fmt.Printf("Received 1 : %s \n", d.Type)
				//if d.CorrelationId != "" && d.
				d.Reject()

				agent.processDelivery(d)
				d.Ack(false)
			}
		}
	}
}

func (agent *Agent) Process() {

}

func (agent *Agent) response(msgType string, body []byte, d amqp.Delivery) {
	p := amqp.Publishing{
		ContentType:   "text/json",
		Body:          body,
		DeliveryMode:  amqp.Persistent,
		Type:          msgType,
		CorrelationId: d.CorrelationId,
	}

	agent.responses <- ResponseMsg{publishing: p, routingKey: d.ReplyTo}
}

func (agent *Agent) Respond(msgType string, body []byte, d amqp.Delivery) {
	for {
		select {
		case <-agent.context.Done():
			agent.client.Close()
		case r, ok := <-agent.responses:
			if ok {
				agent.client.Push(r.routingKey, r.publishing)
			}
		}
	}
}

func saveTask(task Task) {
	fmt.Println(task.Id)
}

func (agent *Agent) respondError() {

}

func (agent *Agent) processDelivery(d amqp.Delivery) error {
	switch d.Type {
	case MsgTypePing:
		//r := Response{Succeed: true}
		break

	case MsgTypeKubeApiRequest:
		var msg KubeApiRequest
		go func() {
			err := json.Unmarshal(d.Body, &msg)
			code, data, err := kubernetes.ApiRequest(msg.URI, msg.Method, msg.Body)

			if err != nil {
				return nil, err
			}
			response := KubeApiResponse{HttpCode: code, Body: data}
			json, err := json.Marshal(respomse)
			if err != nil {
				return nil, err
			}
			return json, err
		}()
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

	if err != nil {
		return nil, err
	}

	d.Ack(false)

	return data, nil
}
