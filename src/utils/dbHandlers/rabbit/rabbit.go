package rabbit

import (
	"fmt"

	"github.com/streadway/amqp"
)

type RabbitHandler struct {
	Conn    *amqp.Connection
	Channel *amqp.Channel
}

// NewRabbitHandler creates and initializes a new RabbitHandler
func NewRabbitHandler(uri string) (*RabbitHandler, error) {
	conn, err := amqp.Dial(uri)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	return &RabbitHandler{
		Conn:    conn,
		Channel: ch,
	}, nil
}

// Publish declares the queue and publishes a message to it
func (h *RabbitHandler) Publish(queueName string, body []byte) error {
	q, err := h.Channel.QueueDeclare(queueName, true, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("failed to declare queue %s: %w", queueName, err)
	}

	err = h.Channel.Publish("", q.Name, false, false, amqp.Publishing{
		DeliveryMode: amqp.Persistent,
		ContentType:  "application/json",
		Body:         body,
	})
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	return nil
}

// Consume declares the queue and returns a delivery channel
func (h *RabbitHandler) Consume(queueName string) (<-chan amqp.Delivery, error) {
	q, err := h.Channel.QueueDeclare(queueName, true, false, false, false, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to declare queue %s: %w", queueName, err)
	}

	msgs, err := h.Channel.Consume(q.Name, "", false, false, false, false, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to register consumer: %w", err)
	}

	return msgs, nil
}

// Close closes the RabbitMQ connection and channel
func (h *RabbitHandler) Close() error {
	if h.Channel != nil {
		if err := h.Channel.Close(); err != nil {
			return err
		}
	}
	if h.Conn != nil {
		if err := h.Conn.Close(); err != nil {
			return err
		}
	}
	return nil
}
