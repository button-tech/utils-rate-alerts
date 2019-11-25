package api

import (
	"encoding/json"
	routing "github.com/qiangxue/fasthttp-routing"
	"github.com/streadway/amqp"
	"github.com/valyala/fasthttp"
)

type alert struct {
	Currency  string `json:"currency"`
	Price     string `json:"price"`
	Fiat      string `json:"fiat"`
	Condition string `json:"condition"`
	URL       string `json:"url"`
}

func (ac *adiController) alert(ctx *routing.Context) error {
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

func (ac *adiController) healthCheck(ctx *routing.Context) error {
	respondWithJSON(ctx, fasthttp.StatusOK, map[string]interface{}{"result": "alive"})
	return nil
}

func (s *Server) initAlertAPI() {
	s.G.Post("/alert", s.ac.alert)
	s.G.Get("/health-check", s.ac.healthCheck)
}
