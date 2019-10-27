package receiver

import (
	"log"
	"os"

	"github.com/imroc/req"
	"github.com/pkg/errors"
	"github.com/streadway/amqp"
)

type process struct {
	Currency string `json:"currency"`
	Fiat     string `json:"fiat"`
}

func (r *Receiver) Process() (<-chan amqp.Delivery, error) {
	msgs, err := r.channel.Consume(
		r.queue.Name,
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

func (r *Receiver) schedule(d amqp.Delivery) {

}

func doRequest(p process) error {
	rq := req.New()
	resp, err := rq.Post(os.Getenv("PRICES"), req.BodyJSON(&p))
	if err != nil {
		return err
	}


}

func hunting202(url string) error {
	var err error
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
		return errors.Wrap(errors.New("no 202 statusCodes"), "hunting202")
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
