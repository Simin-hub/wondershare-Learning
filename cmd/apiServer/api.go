package main

import (
	"github.com/gin-gonic/gin"
	. "github.com/simin/wondshare/internal/api"
	. "github.com/simin/wondshare/internal/pkg"
	"log"
)

func main() {
	var err error
	if err != nil {
		log.Fatalf("Failed to create service: %v", err)
	}

	// Registered Routing
	router := gin.Default()
	router.GET("/task", CreateTask)
	router.GET("/task/:id", QueryTask)
	router.DELETE("/task/:id", CancelTask)
	router.POST("/task/:id/update", UpdateTask)

	// Start the http server
	if err := router.Run(":8080"); err != nil {
		log.Fatalf("failed to start server: %s", err)
	}
}
