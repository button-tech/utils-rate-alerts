package receiver

import (
	"github.com/button-tech/rate-alerts/rabbitmq"
	"github.com/button-tech/rate-alerts/storage"
	"github.com/button-tech/rate-alerts/storage/prices"
	"github.com/pkg/errors"
	"log"
)

type Receiver struct {
	store    *storage.Cache
	pricesStore *prices.Cache
	rabbitMQ *rabbitmq.Instance
}

func New() (*Receiver, error) {
	rabbitMQ, err := rabbitmq.NewInstance()
	if err != nil {
		return nil, errors.Wrap(err, "rabbitMQ instance declaration")
	}

	r := Receiver{
		store:    storage.NewCache(),
		pricesStore: prices.NewCache(),
		rabbitMQ: rabbitMQ,
	}

	return &r, nil
}

func (r *Receiver) Finalize() {
	log.Println("rabbitMQ connection close...")
	if err := r.rabbitMQ.Conn.Close(); err != nil {
		log.Println(err)
		return
	}

	log.Println("rabbitMQ channel close...")
	if err := r.rabbitMQ.Channel.Close(); err != nil {
		log.Println(err)
		return
	}
}
