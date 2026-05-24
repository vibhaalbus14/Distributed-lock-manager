package handlers

import (
	"distributed_lock_manager/internal/manager"
	"distributed_lock_manager/internal/network"
	"distributed_lock_manager/internal/node"
	"distributed_lock_manager/internal/startup"
	"fmt"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// ws hub creation
type Hub struct {
	clients map[*websocket.Conn]bool
	mu      sync.Mutex
}

var MainHub = &Hub{
	clients: make(map[*websocket.Conn]bool),
}

func (h *Hub) Register(conn *websocket.Conn) {
	h.mu.Lock()
	h.clients[conn] = true
	h.mu.Unlock()
}

func (h *Hub) Unregister(conn *websocket.Conn) {
	h.mu.Lock()
	if _, ok := h.clients[conn]; ok {
		delete(h.clients, conn)
		conn.Close()
	}
	h.mu.Unlock()
}

func (h *Hub) BroadcastJSON(v interface{}) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for client := range h.clients {
		// Pushes to every active connection concurrently
		err := client.WriteJSON(v)
		if err != nil {
			client.Close()
			delete(h.clients, client)
		}
	}
}

// StartChannelListener runs ONCE in the background to drain your simulation channels
func StartChannelListener() {
	lm := startup.PointerToLM()

	for {
		select {
		case logMsg := <-network.NetworkLogChannel:
			MainHub.BroadcastJSON(WebSocketPayload{Type: "LOG", Payload: logMsg})

		case deltaMsg := <-node.NodeDeltaChannel:
			MainHub.BroadcastJSON(WebSocketPayload{Type: "NODE_DELTA", Payload: deltaMsg})

		case opMsg := <-node.OpEventChannel:
			secondsString := fmt.Sprintf("%d", int(opMsg.OpTime.Seconds()))
			MainHub.BroadcastJSON(WebSocketPayload{Type: "OP", Payload: PayloadInfo{
				CurrentHolder: lm.CurrentHolder,
				OpTime:        secondsString,
				FencingToken:  lm.FencingToken,
			}})

		case freeChan := <-manager.ManagerFreeChan:
			MainHub.BroadcastJSON(WebSocketPayload{Type: "CHAN_FREE", Payload: freeChan})

		case exitChan := <-ExitChan:
			MainHub.BroadcastJSON(WebSocketPayload{Type: "EXIT_APP", Payload: exitChan})
		}
	}
}

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
	// Upgrade the HTTP protocol connection to a permanent WebSocket tunnel
	conn, err := wsObj.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		return
	}
	fmt.Println("Websocket connection established successfully from backend side")

	// Register this new browser connection to our central hub broadcast list
	MainHub.Register(conn)

	// Clean up if user closes the browser window or netflix disconnects
	defer func() {
		fmt.Println("Closing Connection: Client disconnected from network stream hub.")
		MainHub.Unregister(conn)
	}()

	// Keep the socket connection open by blocking on reads (handles heartbeats/disconnects)
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}

}
