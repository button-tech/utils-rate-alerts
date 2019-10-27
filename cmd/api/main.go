package main

import (
	"github.com/button-tech/rate-alerts/api"
	"github.com/valyala/fasthttp"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	s, err := api.NewServer()
	if err != nil {
		log.Fatal(err)
	}

	server := fasthttp.Server{Handler: s.R.HandleRequest}

	signalEx := make(chan os.Signal, 1)
	defer close(signalEx)

	signal.Notify(signalEx,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	go func() {
		log.Println("start http server")
		if err := fasthttp.ListenAndServe(":5000", s.R.HandleRequest); err != nil {
			log.Fatal(err)
		}
	}()
	defer s.Finalize()
	defer func() {
		if err := server.Shutdown(); err != nil {
			log.Fatal(err)
		}
	}()

	stop := <-signalEx
	log.Println("Received", stop)
	log.Println("Waiting for all jobs to stop")
}
