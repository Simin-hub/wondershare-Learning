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
	UpdateTaskStatus(context.Context, string, string) error   // 更新任务
	UpdateTaskProgress(context.Context, string, string) error // 更新任务进度
}

type AlgService struct {
	MqChan      map[string]*amqp.Channel
	RedisClient *redis.Client
}

func (s *AlgService) CreateTask(ctx context.Context) (string, error) {
	// 生成task id
	task := &Task{
		uuid.NewString(),
		TASK_PENDING,
		"0",
	}
	// 序列化
	taskJson, err := json.Marshal(task)

	// 将task信息发送到mq 首先声明一个队列，然后向队列中发送消息
	q, err := s.MqChan[TASK_QUEUQ].QueueDeclare(
		TASK_QUEUQ, // name 队列名称
		false,      // durable 持久化
		false,      // delete when unused 自动删除
		false,      // exclusive 独占队列
		false,      // no-wait 是否阻塞
		nil,        // arguments
	)
	if err != nil {
		return "", err
	}
	err = s.MqChan[TASK_QUEUQ].Publish(
		"",
		q.Name, // 队列名称
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        taskJson,
		})
	if err != nil {
		return "", err
	}
	err = s.RedisClient.HSet(ctx, task.ID, "status", task.Status).Err()
	err = s.RedisClient.HSet(ctx, task.ID, "process", task.Process).Err()
	if err != nil {
		return "", err
	}
	return task.ID, nil
}

func (s *AlgService) QueryTask(ctx context.Context, id string) (*Task, error) {
	status, err := s.RedisClient.HGet(ctx, id, "status").Result()
	if err != nil {
		return nil, err
	}
	process, err := s.RedisClient.HGet(ctx, id, "process").Result()
	if err != nil {
		return nil, err
	}
	return &Task{
		id,
		status,
		process,
	}, nil
}

func (s *AlgService) UpdateTaskStatus(ctx context.Context, id string, status string) error {
	// 将task状态写入redis
	err := s.RedisClient.HSet(ctx, id, "status", status).Err()
	if err != nil {
		return errors.New(fmt.Sprint("Failed to update task status : ", id))
	}
	return nil
}

func (s *AlgService) UpdateTaskProgress(ctx context.Context, id string, process string) error {
	// 将task进度写入redis
	err := s.RedisClient.HSet(ctx, id, "process", process).Err()
	if err != nil {
		return errors.New(fmt.Sprint("Failed to update task process : ", id))
	}
	return nil
}

func (s *AlgService) CancelTask(ctx context.Context, id string) error {

	task, err := s.QueryTask(ctx, id)
	if err != nil {
		return err
	}
	err = s.UpdateTaskStatus(ctx, id, TASK_CANCELLED)
	if task.Status == TASK_CANCELLED {
		return nil
	} else if task.Status == TASK_COMPLETED {
		return err
	}

	taskJson, err := json.Marshal(task)
	if err != nil {
		return err
	}
	q, err := s.MqChan[CANCEL_QUEUE].QueueDeclare(
		CANCEL_QUEUE, // name 队列名称
		false,        // durable 持久化
		false,        // delete when unused 自动删除
		false,        // exclusive 独占队列
		false,        // no-wait 是否阻塞
		nil,          // arguments
	)
	if err != nil {
		return err
	}
	err = s.MqChan[CANCEL_QUEUE].Publish(
		"",
		q.Name, // 队列名称
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        taskJson,
		})
	if err != nil {
		return err
	}
	return nil
}

func NewService() (*AlgService, error) {
	// 初始化mq连接和redis连接
	// 连接RabbitMQ服务
	// 配置连接套接字，它主要定义连接的协议和身份验证等。
	mqConn, err := amqp.Dial("amqp://root:root@localhost:5672/")
	if err != nil {
		mqConn.Close()
		log.Fatal("Failed to connect to RabbitMQ: ", err)
	}

	mqChan := make(map[string]*amqp.Channel)
	mqChan[TASK_QUEUQ], err = mqConn.Channel()
	if err != nil {
		mqChan[TASK_QUEUQ].Close()
		log.Fatal("Failed to open a channel: ", err)
	}

	mqChan[CANCEL_QUEUE], err = mqConn.Channel()
	if err != nil {
		mqChan[TASK_QUEUQ].Close()
		log.Fatal("Failed to open a channel: ", err)
	}

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
