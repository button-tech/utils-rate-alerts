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
	var r t.TrueCondition
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
