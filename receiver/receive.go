package receiver

import (
	"log"
	"os"
	"time"

	"github.com/jeyldii/rate-alerts/pkg/rabbitmq"
	"github.com/jeyldii/rate-alerts/pkg/storage/cache"
	"github.com/pkg/errors"
	routing "github.com/qiangxue/fasthttp-routing"
	"github.com/valyala/fasthttp"
)

type Receiver struct {
	Server *fasthttp.Server
	r      *routing.Router
	g      *routing.RouteGroup
	c      *controller

	botAlertURL string
	store       *cache.Cache
	rabbitMQ    *rabbitmq.Instance
}

func New() (*Receiver, error) {
	rabbitMQ, err := rabbitmq.NewInstance()
	if err != nil {
		return nil, errors.Wrap(err, "rabbitMQ instance declaration")
	}

	r := &Receiver{
		store:       cache.NewCache(),
		rabbitMQ:    rabbitMQ,
		botAlertURL: os.Getenv("ALERT_BOT_URL"),
		r:           routing.New(),
	}
	r.r.Use(cors)
	r.fs()
	r.initRoute()
	r.mount()

	return r, nil
}

func (r *Receiver) fs() {
	r.Server = &fasthttp.Server{
		ReadTimeout:  time.Second * 30,
		WriteTimeout: time.Second * 30,
		Handler:      r.r.HandleRequest,
	}
}

func (r *Receiver) initRoute() {
	r.g = r.r.Group("/api/processing")
	r.c = &controller{store: r.store}
}

func (r *Receiver) Finalize() {
	log.Println("rabbitMQ channel close...")
	if err := r.rabbitMQ.Channel.Close(); err != nil {
		log.Println(err)
	}

	log.Println("rabbitMQ connection close...")
	if err := r.rabbitMQ.Conn.Close(); err != nil {
		log.Println(err)
	}
}
