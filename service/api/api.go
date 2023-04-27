package main

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	. "github.com/simin/wondshare/service"
	"log"
	"net/http"
)

var service Service

func main() {
	var err error
	service, err = NewService()
	if err != nil {
		log.Fatalf("Failed to create service: %v", err)
	}

	// 注册路由
	router := gin.Default()
	router.GET("/task", createTask)
	router.GET("/task/:id", queryTask)
	router.DELETE("/task/:id", cancelTask)
	router.POST("/task/:id/status", updateTaskStatus)
	router.POST("/task/:id/process", updateTaskProcess)

	// 启动http服务器
	if err := router.Run(":8080"); err != nil {
		log.Fatalf("failed to start server: %s", err)
	}
}

func createTask(c *gin.Context) {
	task, err := service.CreateTask(context.Background())
	if err != nil {
		c.String(http.StatusInternalServerError, "failed to create task : %s", err)
		return
	}
	// 将task id返回给客户端
	c.String(http.StatusOK, task)
}

func queryTask(c *gin.Context) {
	// 获取task id
	id := c.Param("id")
	task, err := service.QueryTask(context.Background(), id)
	if err == redis.Nil {
		// 如果task不存在，返回404
		c.JSON(http.StatusNotFound, nil)
	} else if err != nil {
		// 如果出现其他错误，返回500
		c.JSON(http.StatusInternalServerError, nil)
	} else {
		// 返回task状态
		c.JSON(http.StatusOK, *task)
	}
}

func updateTaskStatus(c *gin.Context) {
	id := c.Param("id")
	task := &Task{}
	err := c.BindJSON(task)
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
	}
	err = service.UpdateTaskStatus(context.Background(), id, task.Status)
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}
	// 返回成功状态给算法服务
	c.String(http.StatusOK, "OK")
}

func cancelTask(c *gin.Context) {
	id := c.Param("id")
	err := service.CancelTask(context.Background(), id)
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}
	c.String(http.StatusOK, "OK")
}

func updateTaskProcess(c *gin.Context) {
	id := c.Param("id")
	task := &Task{}
	err := c.BindJSON(task)
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
	}
	err = service.UpdateTaskProgress(context.Background(), id, task.Process)
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}
	// 返回成功状态给算法服务
	c.String(http.StatusOK, "OK")
}
