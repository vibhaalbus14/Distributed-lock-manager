package handlers

import (
	"distributed_lock_manager/internal/startup"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func Change_latency(ctx *gin.Context) {
	// 1. Extract the specific query parameter strings by their key names
	ns := startup.PointerToNS()
	//lm := startup.PointerToLM()
	minL := ctx.Query("minLatency")
	maxL := ctx.Query("maxLatency")

	// Validate that BOTH parameters were provided by the frontend
	if minL == "" || maxL == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Missing 'minLatency' or 'maxLatency' query parameters"})
		return
	}

	// 2. Parse minLatency into an int
	minInt, err := strconv.Atoi(minL)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid minLatency value provided (must be an integer)"})
		return
	}

	// 3. Parse maxLatency into an int
	maxInt, err := strconv.Atoi(maxL)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid maxLatency value provided (must be an integer)"})
		return
	}

	ns.UpdateLatencyBounds(minInt, maxInt)

	ctx.JSON(http.StatusAccepted, gin.H{
		"res": "updated latency bounds successfully",
	})
}
