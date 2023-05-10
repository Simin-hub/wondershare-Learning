package algorithm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	. "github.com/simin/wondshare/internal/algorithm/cgo"
	. "github.com/simin/wondshare/internal/pkg"
	"github.com/streadway/amqp"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"
)

// GetConsume get consumer delivery queue
func GetConsume(conn *amqp.Connection, queName string) <-chan amqp.Delivery {
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
func Consuming(ctx context.Context, msgs <-chan amqp.Delivery) {
	MaxTasks := 50
	taskQueue := make(chan int, MaxTasks)
	var wg sync.WaitGroup
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
		taskQueue <- 1
		go func() {
			defer wg.Done()
			wg.Add(1)
			worker(ctx, task)
			<-taskQueue
		}()
	}
	wg.Wait()
}

// UpdateTask updates the status of the task
func UpdateTask(id string, status string, process string) error {
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
		fmt.Sprintf("http://localhost:8080/task/%s/update", task.ID),
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

func worker(ctx context.Context, task *Task) {
	// Task timeout settings handling
	child, cancelFunc := context.WithTimeout(ctx, time.Second*50)
	go func() {
		for i := 1; i <= 10; i++ {
			// Check for task timeout
			if child.Err() != nil {
				task.Status = TASK_TIMEDOUT
				break
			}

			// Single process
			StartProcess(task.ID)

			task.Process = strconv.Itoa(i * 10)
			err := UpdateTask(task.ID, TASK_PROCESSING, task.Process)
			if err != nil {
				// updating the task process fails
				log.Printf("The task %s has been cancelled", task.ID)
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

	err := UpdateTask(task.ID, task.Status, task.Process)
	if err != nil {
		log.Printf("The task %s has been cancelled.", task.ID)
		return
	}
	log.Printf("Task completion: %s,\t status: %s\n", task.ID, task.Status)
	return
}
