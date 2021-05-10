package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	log "github.com/OloloevReal/go-simple-log"
)

const version = "0.0.1"

func init() {
	sigChan := make(chan os.Signal)
	go func() {
		for range sigChan {
			log.Printf("[INFO] SIGQUIT detected")
		}
	}()
	signal.Notify(sigChan, syscall.SIGQUIT)

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
		<-stop
		log.Printf("[INFO] interrupt signal")
		cancel()
	}()

	_ = ctx
}

func main() {
	log.Printf("Started go-baby-bot version %s\r\n", version)
	defer log.Println("Finished!")
}
