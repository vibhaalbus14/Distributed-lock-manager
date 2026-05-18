package node

import (
	"distributed_lock_manager/internal/protocol"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

type NodeStatus int //alias

//diff const variables of type nodestatus and iota is an auto increement , starts from 0
const (
	Idle NodeStatus = iota //0
	Request //1
	Holding //2
	Crashed //3
)

type Node struct {
	Id             string
	Status         NodeStatus
	NetworkChannel chan protocol.Message
	StopHeartBeat chan bool
	CurrentToken int64
	bufferChannel chan protocol.Message
	mu sync.Mutex
}

func CreateNode(networkMsg chan protocol.Message,bufferMsg chan protocol.Message)(*Node){
	generatedId :=fmt.Sprint(uuid.New())

	return &Node{Id:generatedId,Status:0,NetworkChannel:networkMsg,StopHeartBeat: make(chan bool),bufferChannel:bufferMsg}
	
}

func (n *Node)RequestLock(){
	n.mu.Lock()
	if n.Status!=Idle{
n.mu.Unlock()
return
	} //not idle

	n.Status=Request
	n.mu.Unlock()
	fmt.Printf("[NODE %s] State changed to REQUESTING. Transmitting MsgRequest...\n", n.Id)

	//since its a chnnel variable, a message struct variable is created and is fed into nodes's networkchannel using channel initialization <-
	n.NetworkChannel <- protocol.Message{Id:fmt.Sprintf("req-%s-%d", n.Id, time.Now().UnixNano()),
NodeId: n.Id,Type: protocol.MsgRequest,
	}

}

//status holding->idle
func (n *Node)ReleaseLock(){
	n.mu.Lock()

	if n.Status!=Holding{
		n.mu.Unlock()
		return 
	}

	n.Status=Idle
	n.mu.Unlock()
	n.StopHeartBeat<-true
	fmt.Printf("[NODE %s] Yielding Resource. Transmitting MsgRelease...\n", n.Id)

	n.NetworkChannel <- protocol.Message{Id:fmt.Sprintf("req-%s-%d", n.Id, time.Now().UnixNano()),
NodeId: n.Id,Type: protocol.MsgRelease,
	}
}

//to move from status request-> holding, from the quesue maintained by manager
func (n *Node)SetToHolding(){
	n.mu.Lock()

	if n.Status!=Request{
		n.mu.Unlock()
		return 
	}

	n.Status=Holding
	n.mu.Unlock()
	fmt.Printf("[NODE %s] Yielding Resource. Transmitting MsgRelease...\n", n.Id)

}

//to restart killed/crashed  nodes
//set them to idle state
func (n *Node)Restart(){
	n.mu.Lock()
	if n.Status!=Crashed{
		n.mu.Unlock()
		return
	}

	//set all the members of that node to def
	n.Status=Idle
	//no cvhange in id
	n.mu.Unlock()
	//the stophearbeat channel could be corrupted or may contain stale data=> make a new channel
	n.StopHeartBeat=make(chan bool)
	fmt.Printf("[NODE %s] Clean recovery successful. State transitioned from CRASHED -> IDLE. Ready for action.\n", n.Id)
}

// Kill simulates a sudden hard cluster failure, immediately freezing operations
func (n *Node) Kill() {
	n.mu.Lock()
	defer n.mu.Unlock()
	
	if n.Status == Holding {
		// Suppress channel write blocks in case routine already terminated
		select {
		case n.StopHeartBeat <- true:
		default:
		}
	}
	n.Status = Crashed
	fmt.Printf("[NODE %s] !!! HARD SYSTEM CRASH SIMULATED !!!\n", n.Id)
}

// StartHeartbeatLoop runs an out-of-band ticker routine only when state == HOLDING
func (n *Node) StartHeartbeatLoop() {
	n.mu.Lock()
	if n.Status != Holding {//heartbeat starts only on hold
		n.mu.Unlock()
		return
	}
	n.mu.Unlock()

	go func() {//another backround routine that constantly sends heartbeat to manager inc or restart the lock lease timer
		//time.NewTicker creates a timer object that internally contains and manages a channel. That channel is accessible via ticker.C.
		ticker := time.NewTicker(1500 * time.Millisecond) // Heartbeat cadence , evry time the timer obj 
		// reaches 1.5 sec, it passes a tick to the obj's channel
		defer ticker.Stop()//automatically done when the function is excited

		fmt.Printf("[NODE %s] Asynchronous Heartbeat Pipeline Active.\n", n.Id)

		for {
			select {//select waits for both cases, whichever case comes first in current iteration, corresponding code is executed
			case <-ticker.C:
				// Pulse heartbeat packet to network simulator channel
				n.NetworkChannel <- protocol.Message{
					Id:       fmt.Sprintf("hb-%s-%d", n.Id, time.Now().UnixNano()),
					NodeId: n.Id,
					Type:     protocol.MsgHeartbeat,
					Token:n.CurrentToken,
				}
			case <-n.StopHeartBeat:
				fmt.Printf("[NODE %s] Asynchronous Heartbeat Pipeline Safely Exited.\n", n.Id)
				return
			}
		}
	}()}
	
	// StartInboundConsumer spikes an infinite background thread loop for this specific node.
// It drains its private buffer channel and transitions the node's status based on manager messages.
func (n *Node) NodeBufferIteration() {
	go func() {
		fmt.Printf("[NODE %s] Inbound network consumer active. Monitoring mailbox buffer...\n", n.Id)

		// This loop blocks and sleeps until a new message packet lands in the channel buffer.
		// It automatically processes packets in strict FIFO order.
		for msg := range n.bufferChannel {
			
			// Grab the local node mutex lock to update properties safely without race conditions
			n.mu.Lock()

			// Edge-Case Guard: If this node is completely crashed/offline, ignore incoming packets
			if n.Status == Crashed {
				n.mu.Unlock()
				continue
			}

			// Evaluate the packet depending on its protocol type
			switch msg.Type {
			case protocol.MsgGrant:
				// SUCCESS: The manager officially gave this node exclusive control of the lock!
				n.Status = Holding
				n.CurrentToken = msg.Token // Cache the incremented fencing token in local RAM
				fmt.Printf("[NODE %s]  LOCK GRANTED! State -> HOLDING. Fencing Token: %d\n", n.Id, msg.Token)
				
				// Automagic Ignition: Automatically kick off our parallel heartbeat ticking thread!
				n.StartHeartbeatLoop()

			case protocol.MsgEvict:
				// EVICTION NOTICE: The manager forcefully stripped our lock due to timeout/glitch.
				n.Status = Idle
				n.CurrentToken = 0
				
				// Safely notify the background heartbeat thread to drop its tools and shut down
				if n.StopHeartBeat != nil {
					select {
					case n.StopHeartBeat <- true:
					default:
					}
				}
				fmt.Printf("[NODE %s] FORCIBLY EVICTED BY MANAGER! State reverted back to -> IDLE.\n", n.Id)
			}

			// Always release the lock at the end of the message processing cycle
			n.mu.Unlock()
		}
	}()
}


