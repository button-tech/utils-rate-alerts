package main

import (
	"log"

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

	log.Println("start processing...")
	go func() {
		for msg := range c {
			log.Printf("Received a message: %s", msg.Body)

		}
	}()

	select {}
}
