package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/jeyldii/rate-alerts/receiver"
)

const port = ":5050"

func main() {
	r, err := receiver.New()
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
		if err := r.Server.ListenAndServe(port); err != nil {
			log.Fatal(err)
		}
	}()

	log.Println("Start processing")
	go r.Processing()
	go r.GetPrices()

	defer r.Finalize()
	defer func() {
		if err := r.Server.Shutdown(); err != nil {
			log.Fatal(err)
		}
	}()

	stop := <-signalEx
	log.Println("Received", stop)
	log.Println("Waiting for all jobs to stop")
}
