package receiver

import (
	"encoding/json"
	"github.com/button-tech/rate-alerts/storage"
	"log"
	"os"
	"time"

	"github.com/imroc/req"
	"github.com/pkg/errors"
	"github.com/streadway/amqp"
)

type process struct {
	Currency string `json:"currency"`
	Fiat     string `json:"fiat"`
}

func (r *Receiver) deliveryChannel() (<-chan amqp.Delivery, error) {
	msgs, err := r.rabbitMQ.Channel.Consume(
		r.rabbitMQ.Queue.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, err
	}

	return msgs, nil
}

func (r *Receiver) Processing() error {
	c, err := r.deliveryChannel()
	if err != nil {
		return err
	}

	for msg := range c {
		var block storage.ConditionBlock
		if err := json.Unmarshal(msg.Body, &block); err != nil {
			log.Println(err)
		}

		r.store.Set(block.Currency, block.Fiat, block)
	}
	select {

	}
}

func (r *Receiver) schedule(d amqp.Delivery) {

}

func doRequest(p process) error {
	rq := req.New()
	_, err := rq.Post(os.Getenv("PRICES"), req.BodyJSON(&p))
	if err != nil {
		return err
	}
	return nil
}

func hunting202(url string) error {
	var err error
	t := time.NewTicker(time.Second * 3)

	counter := 0
	for ; counter != 3; <-t.C {
		err = checkURL(url)
		if err == nil {
			break
		}
		log.Println(err)
		counter++
	}

	if err != nil {
		return errors.Wrap(errors.New("no 202 status codes"), "hunting202")
	}

	return nil
}

func checkURL(url string) error {
	rq := req.New()
	resp, err := rq.Get(url)
	if err != nil {
		return errors.Wrap(err, "checkURL")
	}

	if resp.Response().StatusCode != 202 {
		return errors.Wrap(errors.New("response statusCode not 202"), "checkURL")
	}

	return nil
}
