package api

import (
	"encoding/json"
	"github.com/button-tech/rate-alerts/rabbitmq"
	"github.com/pkg/errors"
	routing "github.com/qiangxue/fasthttp-routing"
	"github.com/streadway/amqp"
	"github.com/valyala/fasthttp"
	"log"
	"net/http"
	"time"
)

type Server struct {
	Core     *fasthttp.Server
	R        *routing.Router
	G        *routing.RouteGroup
	ac       *adiController
	rabbitMQ *rabbitmq.Instance
}

func NewServer() (*Server, error) {
	server := Server{
		R: routing.New(),
	}
	server.R.Use(cors)
	server.fs()

	r, err := rabbitmq.NewInstance()
	if err != nil {
		return nil, errors.Wrap(err, "rabbitMQ instance declaration")
	}
	server.rabbitMQ = r

	server.initBaseRoute()
	server.initAlertAPI()

	return &server, nil
}

func (s *Server) Finalize() {
	log.Println("rabbitMQ connection close...")
	if err := s.rabbitMQ.Conn.Close(); err != nil {
		return
	}

	log.Println("rabbitMQ channel close...")
	if err := s.ac.channel.Close(); err != nil {
		return
	}
}

func cors(ctx *routing.Context) error {
	ctx.Response.Header.Set("Access-Control-Allow-Origin", string(ctx.Request.Header.Peek("Origin")))
	ctx.Response.Header.Set("Access-Control-Allow-Credentials", "false")
	ctx.Response.Header.Set("Access-Control-Allow-Methods", "GET,HEAD,PUT,POST,DELETE")
	ctx.Response.Header.Set(
		"Access-Control-Allow-Headers",
		"Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization",
	)

	if string(ctx.Method()) == "OPTIONS" {
		ctx.Abort()
	}
	if err := ctx.Next(); err != nil {
		if httpError, ok := err.(routing.HTTPError); ok {
			ctx.Response.SetStatusCode(httpError.StatusCode())
		} else {
			ctx.Response.SetStatusCode(http.StatusInternalServerError)
		}

		b, err := json.Marshal(err)
		if err != nil {
			respondWithJSON(ctx, fasthttp.StatusInternalServerError, map[string]interface{}{
				"error": err},
			)
		}
		ctx.SetContentType("application/json")
		ctx.SetBody(b)
	}
	return nil
}

func (s *Server) fs() {
	s.Core = &fasthttp.Server{
		ReadTimeout:  time.Second * 30,
		WriteTimeout: time.Second * 30,
		Handler:      s.R.HandleRequest,
	}
}

func (s *Server) initBaseRoute() {
	s.G = s.R.Group("/api/v1")
	s.ac = &adiController{
		channel: s.rabbitMQ.Channel,
		queue:   s.rabbitMQ.Queue,
	}
}

func respondWithJSON(ctx *routing.Context, code int, payload map[string]interface{}) {
	ctx.SetContentType("application/json")
	ctx.SetStatusCode(code)
	if err := json.NewEncoder(ctx).Encode(payload); err != nil {
		log.Println("write answer", err)
	}
}

type adiController struct {
	channel *amqp.Channel
	queue   amqp.Queue
}
