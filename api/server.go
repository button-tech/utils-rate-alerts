package api

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/button-tech/utils-rate-alerts/pkg/rabbitmq"
	t "github.com/button-tech/utils-rate-alerts/types"
	"github.com/pkg/errors"
	routing "github.com/qiangxue/fasthttp-routing"
	"github.com/streadway/amqp"
	"github.com/valyala/fasthttp"
)

type Server struct {
	Core     *fasthttp.Server
	WG       sync.WaitGroup
	R        *routing.Router
	G        *routing.RouteGroup
	ac       *apiController
	rabbitMQ *rabbitmq.Instance
}

func NewServer() (*Server, error) {
	server := Server{
		R:  routing.New(),
		WG: sync.WaitGroup{},
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
	log.Println("rabbitMQ channel close...")
	if err := s.ac.channel.Close(); err != nil {
		return
	}

	log.Println("rabbitMQ connection close...")
	if err := s.rabbitMQ.Conn.Close(); err != nil {
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
			respondWithJSON(ctx, fasthttp.StatusInternalServerError, t.Payload{"error": err})
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
	s.ac = &apiController{
		channel: s.rabbitMQ.Channel,
		queue:   s.rabbitMQ.Queue,
	}
}

func respondWithJSON(ctx *routing.Context, code int, payload t.Payload) {
	ctx.SetContentType("application/json")
	ctx.SetStatusCode(code)
	if err := json.NewEncoder(ctx).Encode(payload); err != nil {
		log.Println("write answer", err)
	}
}

type apiController struct {
	channel *amqp.Channel
	queue   amqp.Queue
}
