package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/simin/wondshare/service"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/streadway/amqp"

	"log"
	"net/http"
)

var mqChannel *amqp.Channel
var redisClient *redis.Client

func main() {
	// 初始化mq连接和redis连接
	// 连接RabbitMQ服务
	mqConn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatal("Failed to connect to RabbitMQ: ", err)
	}
	defer mqConn.Close()

	// 连接RabbitMQ
	mqChannel, err = mqConn.Channel()
	if err != nil {
		log.Fatal("Failed to open a channel: ", err)
	}
	defer mqChannel.Close()

	redisClient = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	_, err = redisClient.Ping(context.Background()).Result()
	if err != nil {
		log.Fatal("Failed to connect to redis: ", err)
	}

	// 注册路由
	http.HandleFunc("/create_task", createTaskHandler)
	http.HandleFunc("/query_task", queryTaskHandler)
	http.HandleFunc("/callback", callbackHandler)

	// 启动http服务器
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func createTaskHandler(w http.ResponseWriter, r *http.Request) {
	// 生成task id
	id := uuid.NewString()
	task := &service.Task{ID: id}

	// 将task信息转为json格式
	taskJson, err := json.Marshal(task)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Failed to create task")
		return
	}

	// 将task信息发送到mq
	err = mqChannel.Publish(
		"",
		"task_queue",
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        taskJson,
		})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Failed to create task")
		return
	}

	// 将task id返回给客户端
	fmt.Fprint(w, id)
	log.Println("Task created:", id)
}

func queryTaskHandler(w http.ResponseWriter, r *http.Request) {
	// 获取task id
	id := r.URL.Query().Get("id")
	log.Println("querying task:", id)

	// 从redis中获取task状态
	status, err := redisClient.Get(r.Context(), id).Result()
	if err == redis.Nil {
		// 如果task不存在，返回404
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, "Task not found")
	} else if err != nil {
		// 如果出现其他错误，返回500
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Failed to query task")
	} else {
		// 返回task状态
		fmt.Fprint(w, status)
	}
}

func callbackHandler(w http.ResponseWriter, r *http.Request) {
	// 获取回调信息
	var status service.TaskStatus
	err := json.NewDecoder(r.Body).Decode(&status)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Invalid callback payload")
		return
	}

	// 将task状态写入redis
	err = redisClient.Set(r.Context(), status.ID, status.Status, 0).Err()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Failed to update task status")
		return
	}
	// 返回成功状态给算法服务
	fmt.Fprint(w, "OK")
	log.Printf("callback task:%s status: %s\n", status.ID, status.Status)
}
