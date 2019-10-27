package api

import (
	"encoding/json"
	"github.com/streadway/amqp"
	"github.com/valyala/fasthttp"
	"time"

	"github.com/pkg/errors"
	routing "github.com/qiangxue/fasthttp-routing"
)

type alert struct {
	Currency  string `json:"currency"`
	Price     string `json:"price"`
	Fiat      string `json:"fiat"`
	Condition string `json:"condition"`
	URL       string `json:"url"`
}

var t = time.NewTicker(time.Second * 3)

func (ac *apiContoller) alert(ctx *routing.Context) error {
	var (
		body alert
		err  error
	)
	if err = json.Unmarshal(ctx.PostBody(), &body); err != nil {
		return err
	}

	if err = ac.channel.Publish(
		"",
		ac.queue.Name,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        ctx.PostBody(),
		},
	); err != nil {
		return err
	}

	respondWithJSON(ctx, fasthttp.StatusOK, map[string]interface{}{"result": "subscribe"})
	return nil
}

func (s *Server) queueSettings() error {
	q, err := s.ac.channel.QueueDeclare(
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
	s.ac.queue = q

	return nil
}

func (s *Server) initAlertAPI() error {
	if err := s.queueSettings(); err != nil {
		return err
	}
	s.G.Get("/alert", s.ac.alert)

	return nil
}
