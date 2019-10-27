package api

import (
	"encoding/json"
	"github.com/pkg/errors"
	"log"
	"net/http"
	"os"

	routing "github.com/qiangxue/fasthttp-routing"
	"github.com/streadway/amqp"
	"github.com/valyala/fasthttp"
)

type Server struct {
	R        *routing.Router
	G        *routing.RouteGroup
	ac       *apiContoller
	rabbitMQ *amqp.Connection
}

func NewServer() (*Server, error) {
	server := Server{
		R: routing.New(),
	}
	server.R.Use(func(ctx *routing.Context) error {
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
	})

	conn, err := amqp.Dial(os.Getenv("RABBIT_MQ_CONN_URL"))
	if err != nil {
		return nil, errors.Wrap(err, "open rabbitmq connection")
	}
	server.rabbitMQ = conn

	if err := server.initBaseRoute(); err != nil {
		return nil, err
	}

	if err := server.initAlertAPI(); err != nil {
		return nil, err
	}

	return &server, nil
}

func (s *Server) Finalize() error {
	log.Println("rabbitMQ connection close...")
	if err := s.rabbitMQ.Close(); err != nil {
		return err
	}

	log.Println("rabbitMQ channel close...")
	if err := s.ac.channel.Close(); err != nil {
		return err
	}
	return nil
}

func (s *Server) initBaseRoute() error {
	s.G = s.R.Group("/api/v1")

	ch, err := s.rabbitMQ.Channel()
	if err != nil {
		return errors.Wrap(err, "open rabbitMQ channel")
	}
	s.ac = &apiContoller{channel: ch}

	return nil
}

func respondWithJSON(ctx *routing.Context, code int, payload map[string]interface{}) {
	ctx.SetContentType("application/json")
	ctx.SetStatusCode(code)
	if err := json.NewEncoder(ctx).Encode(payload); err != nil {
		log.Println("write answer", err)
	}
}

type apiContoller struct {
	channel *amqp.Channel
	queue   amqp.Queue
}
