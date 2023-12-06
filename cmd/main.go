package main

import (
	"WB0/internal/nats_streaming/publisher"
	"WB0/internal/nats_streaming/subscriber"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	go subscriber.Subscriber()
	go publisher.Publisher()
	go publisher.PublishNewData()

	exitSignal := make(chan os.Signal, 1)
	signal.Notify(exitSignal, syscall.SIGINT, syscall.SIGTERM)
	<-exitSignal

}
