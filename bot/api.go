package bot

import (
	"encoding/json"
	"github.com/button-tech/utils-rate-alerts/pkg/respond"
	t "github.com/button-tech/utils-rate-alerts/types"
	routing "github.com/qiangxue/fasthttp-routing"
	"github.com/streadway/amqp"
	"github.com/valyala/fasthttp"
)

type apiController struct {
	channel *amqp.Channel
	queue   amqp.Queue
	b       *Bot
}

func (ac *apiController) botAlert(ctx *routing.Context) error {
	var r trueCondition
	if err := json.Unmarshal(ctx.PostBody(), &r); err != nil {
		return err
	}
	if err := ac.b.AlertUser(r); err != nil {
		respond.WithJSON(ctx, fasthttp.StatusBadRequest, t.Payload{"error": err})
		return nil
	}
	respond.WithJSON(ctx, fasthttp.StatusAccepted, t.Payload{"result": "ok"})
	return nil
}

func (ac *apiController) healthCheck(ctx *routing.Context) error {
	respond.WithJSON(ctx, fasthttp.StatusOK, t.Payload{"result": "alive"})
	return nil
}

func (s *Server) initBotAPI() {
	s.G.Post("/alert", s.ac.botAlert)
	s.G.Get("/health-check", s.ac.healthCheck)
}

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
