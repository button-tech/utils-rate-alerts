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

func (s *Server) initAlertAPI() {
	s.G.Get("/alert", s.ac.alert)
}
