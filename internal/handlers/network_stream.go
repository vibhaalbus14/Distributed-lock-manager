package handlers

import (
	"distributed_lock_manager/internal/manager"
	"distributed_lock_manager/internal/network"
	"distributed_lock_manager/internal/node"
	"distributed_lock_manager/internal/startup"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// create websocket object
var wsObj = websocket.Upgrader{
	ReadBufferSize:  1024, //specifies read and write channel size
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type WebSocketPayload struct {
	Type    string
	Payload interface{}
}

type PayloadInfo struct {
	CurrentHolder string
	OpTime        string
	FencingToken  int64
}

func Network_stream(ctx *gin.Context) {
	//once client sneds a http req to servser
	//this request is upgraded from http to websocket
	// Upgrade the HTTP protocol connection to a permanent WebSocket tunnel

	conn, err := wsObj.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		return
	}
	fmt.Println("websocket connection established successfully from backed side")
	defer conn.Close()

	lm := startup.PointerToLM()

	for {
		select {
		case logMsg := <-network.NetworkLogChannel:
			err := conn.WriteJSON(WebSocketPayload{Type: "LOG", Payload: logMsg})
			fmt.Println("backend log", err)
			if err != nil {
				fmt.Println("Closing Connection: Client disconnected from network log pipe.")
				return
			}

		// 🚀 CATCH STATUS CHANGES FROM NODE CHANNELS
		case deltaMsg := <-node.NodeDeltaChannel:
			err := conn.WriteJSON(WebSocketPayload{Type: "NODE_DELTA", Payload: deltaMsg})
			fmt.Println("backend delta", err)
			if err != nil {
				fmt.Println("Closing Connection: Client disconnected from delta pipe.")
				return
			}

		// 🚀 CATCH LEASE TIMER ALERTS FROM TIMER HOOKS
		case opMsg := <-node.OpEventChannel:
			// Convert the time.Duration nanoseconds cleanly into an explicit numeric string of seconds
			secondsString := fmt.Sprintf("%d", int(opMsg.OpTime.Seconds()))
			err := conn.WriteJSON(WebSocketPayload{Type: "OP", Payload: PayloadInfo{
				CurrentHolder: lm.CurrentHolder,
				OpTime:        secondsString,
				FencingToken:  lm.FencingToken,
			}})
			fmt.Println("backend op", err)
			if err != nil {
				fmt.Println("Closing Connection: Client disconnected from opEvent pipe.")
				return
			}
		case freeChan := <-manager.ManagerFreeChan:
			err := conn.WriteJSON(WebSocketPayload{Type: "CHAN_FREE", Payload: freeChan})
			if err != nil {
				fmt.Println("Closing Connection: Client disconnected from manger free pipe.")
				return
			}
		}
	}
}
