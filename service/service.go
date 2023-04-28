package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/streadway/amqp"
	"log"
)

const (
	TASK_PENDING    = "pending"
	TASK_PROCESSING = "processing"
	TASK_COMPLETED  = "completed"
	TASK_CANCELLED  = "cancelled"
	TASK_TIMEDOUT   = "timed out"
	TASK_ERROR      = "error"
	TASK_QUEUQ      = "task_queue"
	CANCEL_QUEUE    = "cancel_queue"
)

type Task struct {
	ID      string `json:"id"`
	Status  string `json:"status"`
	Process string `json:"process"`
}

type Service interface {
	CreateTask(context.Context) (string, error)
	QueryTask(context.Context, string) (*Task, error)
	CancelTask(context.Context, string) error
	UpdateTaskStatus(context.Context, string, string) error
	UpdateTaskProgress(context.Context, string, string) error
}

type AlgService struct {
	MqChan      map[string]*amqp.Channel
	RedisClient *redis.Client
}

// CreateTask function Create a task and send it to the message queue,
//returning the task id
func (s *AlgService) CreateTask(ctx context.Context) (string, error) {
	// Create Task
	task := &Task{
		uuid.NewString(),
		TASK_PENDING,
		"0",
	}
	// Serialization of tasks
	taskJson, err := json.Marshal(task)

	// Send task information to mq First declare a queue, then send a message to the queue
	q, err := s.MqChan[TASK_QUEUQ].QueueDeclare(
		TASK_QUEUQ,
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return "", err
	}

	err = s.MqChan[TASK_QUEUQ].Publish(
		"",
		q.Name,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        taskJson,
		})
	if err != nil {
		return "", err
	}

	// Save the status and progress of a task
	err = s.RedisClient.HSet(ctx, task.ID, "status", task.Status).Err()
	err = s.RedisClient.HSet(ctx, task.ID, "process", task.Process).Err()
	if err != nil {
		return "", err
	}

	return task.ID, nil
}

// QueryTask function query tasks by task id
func (s *AlgService) QueryTask(ctx context.Context, id string) (*Task, error) {
	// Query whether a task exists
	status, err := s.RedisClient.HGet(ctx, id, "status").Result()
	if err != nil {
		return nil, err
	}
	process, err := s.RedisClient.HGet(ctx, id, "process").Result()
	if err != nil {
		return nil, err
	}
	// return to task
	return &Task{
		id,
		status,
		process,
	}, nil
}

// UpdateTaskStatus function update tasks by task id
func (s *AlgService) UpdateTaskStatus(ctx context.Context, id string, status string) error {
	// First check if the task exists
	task, err := s.QueryTask(ctx, id)
	if err != nil {
		return err
	}
	if task.Status == TASK_CANCELLED {
		return errors.New(fmt.Sprint("task canceled : ", id))
	}
	// Writing task status to redis
	err = s.RedisClient.HSet(ctx, id, "status", status).Err()
	if err != nil {
		return errors.New(fmt.Sprint("Failed to update task status : ", id))
	}
	return nil
}

// UpdateTaskProgress function update tasks by task id
func (s *AlgService) UpdateTaskProgress(ctx context.Context, id string, process string) error {
	// First check if the task exists
	task, err := s.QueryTask(ctx, id)
	if err != nil {
		return err
	}
	if task.Status == TASK_CANCELLED {
		return errors.New(fmt.Sprint("task canceled : ", id))
	}
	// Writing task process to redis
	err = s.RedisClient.HSet(ctx, id, "process", process).Err()
	if err != nil {
		return errors.New(fmt.Sprint("Failed to update task process : ", id))
	}
	return nil
}

// CancelTask function Cancel tasks by task id
func (s *AlgService) CancelTask(ctx context.Context, id string) error {
	// First check if the task exists
	task, err := s.QueryTask(ctx, id)
	if err != nil {
		return err
	}

	if task.Status == TASK_CANCELLED {
		return nil
	}

	// Update task status to cancelled
	err = s.UpdateTaskStatus(ctx, id, TASK_CANCELLED)
	if err != nil {
		return err
	}

	return nil
}

// NewService function creating a server instance
func NewService() (*AlgService, error) {
	// Initialize mq connections and redis connections
	// Connect to the RabbitMQ service
	// Configure the connection sockets, which mainly define the protocol
	// and authentication of the connection, etc.
	mqConn, err := amqp.Dial("amqp://root:root@localhost:5672/")
	if err != nil {
		mqConn.Close()
		log.Fatal("Failed to connect to RabbitMQ: ", err)
	}

	// Create task creation queue channel
	mqChan := make(map[string]*amqp.Channel)
	mqChan[TASK_QUEUQ], err = mqConn.Channel()
	if err != nil {
		mqChan[TASK_QUEUQ].Close()
		log.Fatal("Failed to open a channel: ", err)
	}

	//mqChan[CANCEL_QUEUE], err = mqConn.Channel()
	//if err != nil {
	//	mqChan[TASK_QUEUQ].Close()
	//	log.Fatal("Failed to open a channel: ", err)
	//}

	redisClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	_, err = redisClient.Ping(context.Background()).Result()
	if err != nil {
		log.Fatal("Failed to connect to redis: ", err)
	}
	return &AlgService{
		mqChan,
		redisClient,
	}, nil
}
