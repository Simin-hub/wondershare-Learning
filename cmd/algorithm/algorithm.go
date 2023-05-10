package main

import (
	"context"
	. "github.com/simin/wondshare/internal/algorithm"
	. "github.com/simin/wondshare/internal/pkg"
	"github.com/streadway/amqp"
	"log"
)

func main() {
	// Creating a RabbitMQ connection
	conn, err := amqp.Dial("amqp://root:root@localhost:5672/")
	if err != nil {
		log.Fatal("Failed to connect to RabbitMQ: ", err)
	}
	defer conn.Close()

	// Get the consumption queue
	ConsuDelvi := GetConsume(conn, TASK_QUEUQ)

	ctx := context.Background()
	// Open goroutine consuming messages
	go Consuming(ctx, ConsuDelvi)

	forever := make(chan bool)
	println("Start processing tasks")
	<-forever
}
