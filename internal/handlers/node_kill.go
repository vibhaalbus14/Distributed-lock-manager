package handlers

import (
	cluster_manager "distributed_lock_manager/internal/nodeClusterManager"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Val struct {
	Id string `binding:"required"`
}

func Node_kill_status(ctx *gin.Context) {

	var v Val
	err := ctx.ShouldBindJSON(&v)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"res": "please provide id in string format"})
	}
	nodeId := v.Id
	// Look up the live memory pointer in your cluster map
	n, exists := cluster_manager.HmNodes[nodeId]

	// Gracefully handle a missing node instead of crashing the server
	if !exists || n == nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"error": "Node execution context not found or already deleted",
		})
		return
	}

	n.Kill()

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Termination instruction sent to node background buffer",
	})

}
