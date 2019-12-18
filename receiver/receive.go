package receiver

import (
	"log"
	"os"

	"github.com/button-tech/rate-alerts/pkg/rabbitmq"
	"github.com/button-tech/rate-alerts/pkg/storage/cache"
	"github.com/pkg/errors"
)

type Receiver struct {
	botAlertURL string
	store       *cache.Cache
	rabbitMQ    *rabbitmq.Instance
}

func New() (*Receiver, error) {
	rabbitMQ, err := rabbitmq.NewInstance()
	if err != nil {
		return nil, errors.Wrap(err, "rabbitMQ instance declaration")
	}

	return &Receiver{
		store:       cache.NewCache(),
		rabbitMQ:    rabbitMQ,
		botAlertURL: os.Getenv("ALERT_BOT_URL"),
	}, nil
}

func (r *Receiver) Finalize() {
	log.Println("rabbitMQ connection close...")
	if err := r.rabbitMQ.Conn.Close(); err != nil {
		log.Println(err)
	}

	log.Println("rabbitMQ channel close...")
	if err := r.rabbitMQ.Channel.Close(); err != nil {
		log.Println(err)
	}
}
