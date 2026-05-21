package handlers

import (
	"distributed_lock_manager/internal/node"
	cluster_manager "distributed_lock_manager/internal/nodeClusterManager"
	"distributed_lock_manager/internal/protocol"
	"distributed_lock_manager/internal/startup"

	"github.com/gin-gonic/gin"
)

func Create_node(ctx *gin.Context) {
	ns := startup.PointerToNS()
	//lm := startup.PointerToLM()
	bufferChan := make(chan protocol.Message, 50)
	n := node.CreateNode(ns.FromNodePipe, bufferChan)
	go ns.AddToRegistry(n.Id, bufferChan) //initialise node registry
	go n.NodeBufferIteration()
	go n.RequestLock()
	go cluster_manager.RegisterNode(n)

	ctx.JSON(201, gin.H{"nodeId": n.Id, "status": n.Status}) //201 = created
}
