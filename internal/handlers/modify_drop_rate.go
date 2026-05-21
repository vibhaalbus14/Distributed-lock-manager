package handlers

import (
	"distributed_lock_manager/internal/startup"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func Change_droprate(ctx *gin.Context) {
	ns := startup.PointerToNS()
	//lm := startup.PointerToLM()
	valStr := ctx.Query("rate")
	if valStr == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Missing 'rate' query parameter"})
		return
	}

	// 2. Parse the string into a float64 (64 indicates the bit size)
	valFloat, err := strconv.ParseFloat(valStr, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid float value provided"})
		return
	}

	// 3. Pass the successfully parsed float to your simulator
	ns.UpdateDropRate(valFloat)

	ctx.JSON(http.StatusAccepted, gin.H{"res": "updated drop rate successfully"})
}
