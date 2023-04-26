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
	QueryTask(string, context.Context) (*Task, error)
	CancelTask(string, context.Context) error
	UpdateTaskStatus(string, string, context.Context) error   // 更新任务
	UpdateTaskProgress(string, string, context.Context) error // 更新任务进度
}

type AlgService struct {
	MqChannel   *amqp.Channel
	RedisClient *redis.Client
}

func (s *AlgService) UpdateTaskProgress(id string, process string, ctx context.Context) error {
	// 将task进度写入redis
	err := s.RedisClient.HSet(ctx, id, "process", process).Err()
	if err != nil {
		return errors.New(fmt.Sprint("Failed to update task process : ", id))
	}
	return nil
}

func (s *AlgService) CancelTask(id string, ctx context.Context) error {
	status, err := s.RedisClient.HGet(ctx, id, "status").Result()
	if err != nil {
		return err
	}
	if status == TASK_COMPLETED {
		return nil
	}

	taskJson, err := json.Marshal(&Task{ID: id, Status: status})
	if err != nil {
		return err
	}
	q, err := s.MqChannel.QueueDeclare(
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
	err = s.MqChannel.Publish(
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
	q, err := s.MqChannel.QueueDeclare(
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
	err = s.MqChannel.Publish(
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

func (s *AlgService) QueryTask(id string, ctx context.Context) (*Task, error) {
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

func (s *AlgService) UpdateTaskStatus(id string, status string, ctx context.Context) error {
	// 将task状态写入redis
	err := s.RedisClient.HSet(ctx, id, "status", status).Err()
	if err != nil {
		return errors.New(fmt.Sprint("Failed to update task status : ", id))
	}
	return nil
}

func NewService() (*AlgService, error) {
	// 初始化mq连接和redis连接
	// 连接RabbitMQ服务
	// 配置连接套接字，它主要定义连接的协议和身份验证等。
	mqConn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		mqConn.Close()
		log.Fatal("Failed to connect to RabbitMQ: ", err)
	}

	// 创建一个channel来传递消息
	mqChannel, err := mqConn.Channel()
	if err != nil {
		mqChannel.Close()
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
		mqChannel,
		redisClient,
	}, nil
}
