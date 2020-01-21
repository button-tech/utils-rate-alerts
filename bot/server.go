package bot

import (
	"context"
	"encoding/json"
	"github.com/button-tech/utils-rate-alerts/pkg/rabbitmq"
	"github.com/button-tech/utils-rate-alerts/pkg/respond"
	"github.com/pkg/errors"
	routing "github.com/qiangxue/fasthttp-routing"
	"github.com/valyala/fasthttp"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

type Server struct {
	Core     *fasthttp.Server
	WG       sync.WaitGroup
	Bot      *Bot
	R        *routing.Router
	G        *routing.RouteGroup
	ac       *apiController
	rabbitMQ *rabbitmq.Instance
}

func NewServer(ctx context.Context) (*Server, error) {
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

	bp := SetupBot(r.Channel, r.Queue, os.Getenv("BOT_TOKEN"))
	b, err := CreateBot(bp)
	if err != nil {
		return nil, err
	}
	server.WG.Add(1)
	go b.ProcessingUpdates(ctx, &server.WG)
	server.Bot = b

	server.initBaseRoute()
	server.initBotAPI()

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
			respond.WithJSON(ctx, fasthttp.StatusInternalServerError, map[string]interface{}{"error": err})
			return err
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
	s.G = s.R.Group("/api/tel-bot")
	s.ac = &apiController{
		channel: s.rabbitMQ.Channel,
		queue:   s.rabbitMQ.Queue,
		b:       s.Bot,
	}
}
