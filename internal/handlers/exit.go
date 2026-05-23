package handlers

import (
	"distributed_lock_manager/internal/node"
	cluster_manager "distributed_lock_manager/internal/nodeClusterManager"
	"distributed_lock_manager/internal/protocol"
	"distributed_lock_manager/internal/startup"
	"net/http"

	"github.com/gin-gonic/gin"
)

var ExitChan chan bool

func init() {
	ExitChan = make(chan bool, 1)
}
func Exit_server(ctx *gin.Context) {

	lm := startup.PointerToLM()
	ns := startup.PointerToNS()
	ns.Mu.Lock()
	ns.NodeRegistry = make(map[string]chan protocol.Message)
	ns.Mu.Unlock()
	lm.Mu.Lock()
	lm.CurrentHolder = ""
	lm.WaitQueue = []string{}
	lm.FencingToken = 0

	if lm.LeaseTimer != nil {
		lm.LeaseTimer.Stop()
		lm.LeaseTimer = nil // Free the reference pointer
	}

	lm.Mu.Unlock()
	cluster_manager.RegMu.Lock()

	cluster_manager.AllNodes = []*node.Node{}

	cluster_manager.HmNodes = make(map[string]*node.Node)

	cluster_manager.RegMu.Unlock()

	ExitChan <- true
	ctx.JSON(http.StatusOK, gin.H{"message": "application restarted successfully"})
}
