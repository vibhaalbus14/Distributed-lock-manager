package main

import (
	"distributed_lock_manager/internal/handlers"
	"distributed_lock_manager/internal/startup"
	"fmt"

	"github.com/gin-gonic/gin"
)

func main() {
	server := gin.Default()

	// IMPORTANT FOR FULL-STACK: Allow React (localhost:5173) to talk to Go (localhost:8080)
	server.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	//initailse all channels
	startup.Initialize()
	nodes := server.Group("/nodes")
	{
		nodes.GET("", handlers.Create_node)
		nodes.GET("/status", handlers.Get_node_status)
		nodes.POST("/kill", handlers.Node_kill_status)
		nodes.POST("/request", handlers.Node_request_status)
	}
	//// 1. Path Param: Explicitly type the colon ":" followed by the name
	//server.POST("/api/nodes/:node_id/kill", handlers.Node_kill_status)

	// 2. Query Param: Type the plain static path. NO question marks, NO colons!
	//server.PUT("/api/network/droprate", handlers.Change_droprate)

	server.PUT("/droprate", handlers.Change_droprate)
	server.PUT("/latency", handlers.Change_latency)
	server.GET("/exit", handlers.Exit_server)
	server.Run(":8080")
	<-startup.ShutdownChan

	// Clean up background pipelines here if necessary before the process closes
	fmt.Println("Server stopped cleanly.")
}
