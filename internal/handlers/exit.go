package handlers

import (
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

func Exit_server(ctx *gin.Context) {
	// Send the shutdown confirmation response back to React
	ctx.JSON(http.StatusOK, gin.H{"message": "Server shutting down ..."})

	// Launch a goroutine to wait slightly, allowing the network buffers to flush
	go func() {
		time.Sleep(2000 * time.Millisecond) // Half a second is plenty of time
		os.Exit(0)
	}()
}
