package api

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	. "github.com/simin/wondshare/internal/pkg"
	"net/http"
)

var service Service

func CreateTask(c *gin.Context) {
	task, err := service.CreateTask(context.Background())
	if err != nil {
		c.String(http.StatusInternalServerError, "failed to create task : %s", err)
		return
	}
	// Return the task id to the client
	c.String(http.StatusOK, task)
}

func QueryTask(c *gin.Context) {
	// Get task id
	id := c.Param("id")
	task, err := service.QueryTask(context.Background(), id)
	if err == redis.Nil {
		c.JSON(http.StatusNotFound, nil)
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, nil)
	} else {
		c.JSON(http.StatusOK, *task)
	}
}

func UpdateTask(c *gin.Context) {
	id := c.Param("id")
	task := &Task{}
	err := c.BindJSON(task)
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
	}

	err = service.UpdateTask(context.Background(), id, task.Status, task.Process)
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	// Return success status to the algorithm service
	c.String(http.StatusOK, "OK")
}

func CancelTask(c *gin.Context) {
	id := c.Param("id")
	err := service.CancelTask(context.Background(), id)
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}
	c.String(http.StatusOK, "OK")
}
