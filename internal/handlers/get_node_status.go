package handlers

import (
	cluster_manager "distributed_lock_manager/internal/nodeClusterManager"
	"distributed_lock_manager/internal/startup"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Get_node_status(ctx *gin.Context) {
	//ns := startup.PointerToNS()
	lm := startup.PointerToLM()
	hm := cluster_manager.GetAllNodeStatus()
	if hm == nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"res": "couldnt fetch node status"})
	}
	ctx.JSON(http.StatusOK, gin.H{"status": hm, "currHolder": lm.CurrentHolder, "fencingToken": lm.FencingToken})
}
