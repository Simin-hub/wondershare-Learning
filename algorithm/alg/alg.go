package main

/*
#include "hello.h"
*/
import "C"

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	. "github.com/simin/wondshare/service"
	"github.com/streadway/amqp"
	"log"
	"net/http"
	"strconv"
	"sync"
)

var cancelTask sync.Map

func main() {
	// 连接mq
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatal("Failed to connect to RabbitMQ: ", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatal("Failed to open a channel: ", err)
	}
	defer ch.Close()

	conuserQue := getConsume(ch, TASK_QUEUQ)
	cancelQue := getConsume(ch, CANCEL_QUEUE)

	ctx := context.Background()

	go consuming(ctx, conuserQue)
	go cancelMQ(ctx, cancelQue)

	forever := make(chan bool)
	// 阻塞主进程
	<-forever
}

func getConsume(ch *amqp.Channel, queName string) <-chan amqp.Delivery {
	// 声明队列
	q, err := ch.QueueDeclare(
		queName, // 队列名称
		false,   // 队列持久化
		false,   // 不自动删除队列
		false,   // 不独占队列
		false,   // 不等待队列消费者
		nil,
	)
	if err != nil {
		log.Fatal("Failed to declare a queue: ", err)
	}

	// 消费队列中的消息
	msgs, err := ch.Consume(
		q.Name, // 队列名称
		"",     // 消费者名称
		false,  // 关闭自动确认
		false,  // 不独占队列
		false,  // 不等待队列消费者
		false,  // 不阻塞
		nil,
	)
	if err != nil {
		log.Fatal("Failed to register a consumer: ", err)
	}
	return msgs
}

func consuming(ctx context.Context, msgs <-chan amqp.Delivery) {
	for d := range msgs {
		task := &Task{}
		err := json.Unmarshal(d.Body, task)
		if err != nil {
			log.Println("Failed to decode task message: ", err)
			continue
		} else {
			log.Println("Ongoing consumption task: ", task.ID)
		}

		// 模拟算法处理, 设置函数执行超时设置
		task.Process = "0"
		err = updateTaskStatus(task.ID, TASK_PROCESSING, task.Process)
		if err != nil {
			_ = updateTaskStatus(task.ID, TASK_ERROR, task.Process)
		}
		cancel, status := false, TASK_COMPLETED
		for i := 1; i <= 10; i++ {
			C.Hello()
			task.Process = strconv.Itoa(i * 10)
			err = updateTaskProcess(task.ID, TASK_PROCESSING, task.Process)
			if _, cancel = cancelTask.Load(task.ID); cancel {
				status = TASK_CANCELLED
				break
			}
		}
		err = updateTaskStatus(task.ID, status, task.Process)
		if err != nil {
			log.Println(err)
		}
		// 确认消息已发送
		d.Ack(false)
		log.Println("Task completion: ", task.ID)
	}
}

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

	_, err = http.Post(
		fmt.Sprintf("http://localhost:8080/task/%s/process", task.ID),
		"application/json", bytes.NewBuffer(taskJson))
	return err
}

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

	_, err = http.Post(
		fmt.Sprintf("http://localhost:8080/task/%s/status", task.ID),
		"application/json", bytes.NewBuffer(taskJson))
	return err
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
