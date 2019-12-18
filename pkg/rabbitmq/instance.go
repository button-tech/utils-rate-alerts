package rabbitmq

import (
	"github.com/pkg/errors"
	"github.com/streadway/amqp"
	"os"
)

type Instance struct {
	Conn    *amqp.Connection
	Channel *amqp.Channel
	Queue   amqp.Queue
}

func NewInstance() (*Instance, error) {
	conn, err := amqp.Dial(os.Getenv("RABBIT_MQ_CONN_URL"))
	if err != nil {
		return nil, errors.Wrap(err, "rabbitMQ connection")
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, errors.Wrap(err, "rabbitMQ channel")
	}

	i := Instance{
		Conn:    conn,
		Channel: ch,
	}
	if err := i.queueSettings(); err != nil {
		return nil, err
	}

	return &i, nil
}

func (i *Instance) queueSettings() error {
	q, err := i.Channel.QueueDeclare(
		"alert",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return errors.Wrap(err, "queue settings init")
	}
	i.Queue = q

	return nil
}
