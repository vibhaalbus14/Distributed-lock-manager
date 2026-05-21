package handlers

import (
	cluster_manager "distributed_lock_manager/internal/nodeClusterManager"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Node_request_status(ctx *gin.Context) {
	nodeId := ctx.Param("id")

	// Look up the live memory pointer in your cluster map
	n, exists := cluster_manager.HmNodes[nodeId]

	// Gracefully handle a missing node instead of crashing the server
	if !exists || n == nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"error": "Node execution context not found or already deleted",
		})
		return
	}

	n.RequestLock()

	ctx.JSON(http.StatusOK, gin.H{
		"message": " Request sent to nw",
	})

}
