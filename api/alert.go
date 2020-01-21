package api

import (
	"encoding/json"

	"github.com/button-tech/utils-rate-alerts/pkg/respond"
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

type trueCondition struct {
	Result string `json:"result"`
	Values struct {
		Currency     string `json:"currency"`
		Condition    string `json:"condition"`
		Fiat         string `json:"fiat"`
		Price        string `json:"price"`
		CurrentPrice string `json:"currentPrice"`
	} `json:"values"`
	URL string `json:"url"`
}

func (ac *apiController) alert(ctx *routing.Context) error {
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

	respond.WithJSON(ctx, fasthttp.StatusOK, map[string]interface{}{"result": "subscribe"})
	return nil
}

//func (ac *apiController) botAlert(ctx *routing.Context) error {
//	var r trueCondition
//	if err := json.Unmarshal(ctx.PostBody(), &r); err != nil {
//		return err
//	}
//	if err := ac.b.AlertUser(r); err != nil {
//		return err
//	}
//	respondWithJSON(ctx, fasthttp.StatusAccepted, map[string]interface{}{"result": "ok"})
//	return nil
//}

func (ac *apiController) healthCheck(ctx *routing.Context) error {
	respond.WithJSON(ctx, fasthttp.StatusOK, map[string]interface{}{"result": "alive"})
	return nil
}

func (s *Server) initAlertAPI() {
	//s.G.Post("/bot/alert", s.ac.botAlert)
	s.G.Post("/alert", s.ac.alert)
	s.G.Get("/health-check", s.ac.healthCheck)
}
