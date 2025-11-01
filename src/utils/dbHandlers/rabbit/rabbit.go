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
