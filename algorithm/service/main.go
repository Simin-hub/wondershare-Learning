package main

/*
#include "hello.h"
*/
import "C"

import (
	"bytes"
	"context"
	"encoding/json"
	. "github.com/simin/wondshare/service"
	"github.com/streadway/amqp"
	"log"
	"net/http"
	"time"
)

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

	// 声明队列
	q, err := ch.QueueDeclare(
		"task_queue", // 队列名称
		true,         // 队列持久化
		false,        // 不自动删除队列
		false,        // 不独占队列
		false,        // 不等待队列消费者
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

	forever := make(chan bool)
	// 处理消息
	ctx := context.Background()
	go consuming(ctx, msgs)
	// 阻塞主进程
	<-forever
}

func consuming(ctx context.Context, msgs <-chan amqp.Delivery) {
	for d := range msgs {
		task := &Task{}
		err := json.Unmarshal(d.Body, task)
		if err != nil {
			log.Println("Failed to decode task message: ", err)
		} else {
			log.Println("Ongoing consumption task: ", task.ID)
		}

		status := &TaskStatus{ID: task.ID}

		// 模拟算法处理, 设置函数执行超时设置
		ch := make(chan bool, 1)
		go func(ch chan bool) {
			C.Hello()
			ch <- true
		}(ch)
		select {
		case <-time.After(time.Second * 5):
			status.Status = "timeout"
		case <-ch:
			status.Status = "completed"
		}
		close(ch)

		// 回调api服务通知任务完成状态
		statusJson, err := json.Marshal(status)
		if err != nil {
			log.Println("Failed to encode task status: ", err)
		}

		resp, err := http.Post("http://localhost:8080/callback", "application/json", bytes.NewBuffer(statusJson))
		if err != nil {
			log.Println("Failed to notify task completion: ", err)
		} else if resp.StatusCode != http.StatusOK {
			log.Println("Failed to notify task completion: ", resp.Status)
		}

		// 确认消息已发送
		d.Ack(false)
	}
}
