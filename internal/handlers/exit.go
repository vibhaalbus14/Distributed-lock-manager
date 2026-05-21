package handlers

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func Exit_server(ctx *gin.Context) {
	// 1. Tell React that the server is shutting down successfully
	ctx.JSON(http.StatusOK, gin.H{"message": "Server shutting down ..."})

	// 2. Push a termination signal into the channel in a background thread
	// so it doesn't block this current HTTP response from finishing
	os.Exit(0)
}
