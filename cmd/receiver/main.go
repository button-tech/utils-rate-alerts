package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/button-tech/rate-alerts/receiver"
)

func main() {
	r, err := receiver.New()
	if err != nil {
		log.Fatal(err)
	}

	c, err := r.ProcessChannel()
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

	log.Println("Start processing")
	go func() {
		for msg := range c {
			log.Printf("Received a message: %s", msg.Body)

		}
	}()
	defer r.Finalize()

	stop := <-signalEx
	log.Println("Received", stop)
	log.Println("Waiting for all jobs to stop")
}
