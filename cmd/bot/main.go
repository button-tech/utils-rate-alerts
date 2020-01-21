package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/jeyldii/rate-alerts/bot/telegram"
)

const port = ":5055"

func main() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	s, err := telegram.NewServer(ctx)
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
	cancel()
	log.Println("API-BOT", stop)
	s.WG.Wait()
	log.Println("Waiting for all jobs to stop")
}
