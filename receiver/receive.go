package receiver

import (
	"log"
	"os"

	"github.com/pkg/errors"
	"github.com/streadway/amqp"
)

type Receiver struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	queue   amqp.Queue
}

func New() (*Receiver, error) {
	conn, err := amqp.Dial(os.Getenv("RABBIT_MQ_CONN_URL"))
	if err != nil {
		return nil, errors.Wrap(err, "rabbitMQ connection")
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, errors.Wrap(err, "rabbitMQ channel")
	}

	r := Receiver{
		conn:    conn,
		channel: ch,
	}
	if err := r.queueSettings(); err != nil {
		return nil, err
	}

	return &r, nil
}

func (r *Receiver) queueSettings() error {
	q, err := r.channel.QueueDeclare(
		"alert",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return errors.Wrap(err, "queue settings init")
	}
	r.Queue = q

	return nil
}

func (r *Receiver) Finalize() error {
	log.Println("rabbitMQ connection close...")
	if err := r.conn.Close(); err != nil {
		return err
	}

	log.Println("rabbitMQ channel close...")
	if err := r.channel.Close(); err != nil {
		return err
	}

	return nil
}
