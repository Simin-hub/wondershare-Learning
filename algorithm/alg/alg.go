package main

/*
#include "hello.h"
*/
import "C"

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	. "github.com/simin/wondshare/service"
	"github.com/streadway/amqp"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"
)

var cancelTask sync.Map

func main() {
	// Creating a RabbitMQ connection
	conn, err := amqp.Dial("amqp://root:root@localhost:5672/")
	if err != nil {
		log.Fatal("Failed to connect to RabbitMQ: ", err)
	}
	defer conn.Close()

	// Get the consumption queue
	ConsuDelvi := getConsume(conn, TASK_QUEUQ)
	//cancelQue := getConsume(conn, CANCEL_QUEUE)

	ctx := context.Background()
	// Open goroutine consuming messages
	go consuming(ctx, ConsuDelvi)
	//go cancelMQ(ctx, cancelQue)

	forever := make(chan bool)
	println("Start processing tasks")
	<-forever
}

// getConsume get consumer delivery queue
func getConsume(conn *amqp.Connection, queName string) <-chan amqp.Delivery {
	ch, err := conn.Channel()
	if err != nil {
		log.Fatal("Failed to open a channel: ", err)
	}

	q, err := ch.QueueDeclare(
		queName,
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatal("Failed to declare a queue: ", err)
	}

	msgs, err := ch.Consume(
		q.Name,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatal("Failed to register a consumer: ", err)
	}
	return msgs
}

// Consuming task
func consuming(ctx context.Context, msgs <-chan amqp.Delivery) {
	for d := range msgs {
		// decode message
		task := &Task{}
		err := json.Unmarshal(d.Body, task)
		if err != nil {
			log.Println("Failed to decode task message: ", err)
			continue
		} else {
			log.Println("Ongoing consumption task: ", task.ID)
		}

		// Simulation algorithm processing
		task.Process, task.Status = "0", TASK_COMPLETED
		err = updateTaskStatus(task.ID, TASK_PROCESSING, task.Process)
		if err != nil {
			// updating the task status fails
			d.Ack(false)
			log.Printf("The task %s has been cancelled or does not exist. ", task.ID)
			break
		}

		// Task timeout settings handling
		child, cancelFunc := context.WithTimeout(ctx, time.Second*50)
		go func() {
			for i := 1; i <= 10; i++ {
				// Check for task timeout
				if child.Err() != nil {
					task.Status = TASK_TIMEDOUT
					break
				}

				//if _, ok := cancelTask.LoadAndDelete(task.ID); ok {
				//	task.Status = TASK_CANCELLED
				//	break
				//}

				C.Hello()
				task.Process = strconv.Itoa(i * 10)
				err = updateTaskProcess(task.ID, TASK_PROCESSING, task.Process)
				if err != nil {
					// updating the task process fails
					d.Ack(false)
					log.Printf("The task %s has been cancelled or does not exist. ", task.ID)
					break
				}
				log.Printf("Task id:%s\t process:%s\t", task.ID, task.Process)
			}
			cancelFunc()
		}()

		// Wait for task completion or timeout
		select {
		case <-child.Done():
		}

		err = updateTaskStatus(task.ID, task.Status, task.Process)
		if err != nil {
			d.Ack(false)
			log.Printf("The task %s has been cancelled or does not exist. ", task.ID)
			break
		}
		d.Ack(false)
	}
}

// updateTaskStatus updates the process of the task
func updateTaskProcess(id string, status string, process string) error {
	task := &Task{
		id,
		status,
		process,
	}
	taskJson, err := json.Marshal(task)
	if err != nil {
		log.Println("Failed to encode task status: ", err)
	}

	// Request to API server to modify task progress
	resp, err := http.Post(
		fmt.Sprintf("http://localhost:8080/task/%s/process", task.ID),
		"application/json", bytes.NewBuffer(taskJson))
	if err != nil {
		// Request error
		return err
	}

	// Whether the requested task status is cancelled or does not exist
	if resp.StatusCode != 200 {
		return errors.New("the task has been cancelled or does not exist")
	}
	return nil
}

// updateTaskStatus updates the status of the task
func updateTaskStatus(id string, status string, process string) error {
	task := &Task{
		id,
		status,
		process,
	}
	taskJson, err := json.Marshal(task)
	if err != nil {
		log.Println("Failed to encode task status: ", err)
	}

	// Request to API server to modify task progress, resp indicates a response to a request
	resp, err := http.Post(
		fmt.Sprintf("http://localhost:8080/task/%s/status", task.ID),
		"application/json", bytes.NewBuffer(taskJson))
	if err != nil {
		// Request error
		return err
	}

	// Whether the requested task status is cancelled or does not exist
	if resp.StatusCode != 200 {
		return errors.New("the task has been cancelled or does not exist")
	}
	return nil
}

func cancelMQ(ctx context.Context, msgs <-chan amqp.Delivery) {
	for d := range msgs {
		task := &Task{}
		err := json.Unmarshal(d.Body, task)
		if err != nil {
			log.Println("Failed to decode task message: ", err)
		} else {
			log.Println("Ongoing cancel task: ", task.ID)
		}
		if _, ok := cancelTask.Load(task.ID); !ok {
			cancelTask.Store(task.ID, false)
		}
		d.Ack(false)
	}
}
