package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/jeyldii/rate-alerts/api"
)

const port = ":5001"

func main() {
	s, err := api.NewServer()
	if err != nil {
		log.Fatal(err)
	}

	signalEx := make(chan os.Signal, 1)
	defer close(signalEx)

	signal.Notify(signalEx,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	go func() {
		log.Printf("start http server on port:%s", port)
		if err := s.Core.ListenAndServe(port); err != nil {
			log.Fatal(err)
		}
	}()
	defer s.Finalize()
	defer func() {
		if err := s.Core.Shutdown(); err != nil {
			log.Fatal(err)
		}
	}()

	stop := <-signalEx
	log.Println("Received", stop)
	log.Println("Waiting for all jobs to stop")
}
