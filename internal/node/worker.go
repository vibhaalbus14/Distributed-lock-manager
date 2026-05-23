package node

import (
	"distributed_lock_manager/internal/protocol"
	"fmt"
	"math/rand/v2"
	"sync"
	"time"

	"github.com/google/uuid"
)

type NodeStatus int //alias

// diff const variables of type nodestatus and iota is an auto increement , starts from 0
const (
	Idle    NodeStatus = iota //0
	Request                   //1
	Holding                   //2
	Crashed                   //3
)

type Node struct {
	Id             string
	Status         NodeStatus
	NetworkChannel chan protocol.Message //node -> network channel
	StopHeartBeat  chan bool             //indication to stop heartbeat
	CurrentToken   int64                 //token given by centram manager to filter stale tokens
	bufferChannel  chan protocol.Message //every node will have its own buffer channel through which it receives messages from manager
	Mu             sync.Mutex            //locks to access its current state variables
	AttemptCount   int                   //no of attempts used for exponential backoff
}

type NodeDelta struct {
	NodeId string
	Status NodeStatus
}

type OpEvent struct {
	NodeId string
	OpTime time.Duration
}

var (
	NodeDeltaChannel chan NodeDelta
	OpEventChannel   chan OpEvent
)

func init() {
	NodeDeltaChannel = make(chan NodeDelta, 100)
	OpEventChannel = make(chan OpEvent, 100)
} //automatically called by compiler when the package is seen

func PushDeltaChannel(n *Node) {
	NodeDeltaChannel <- NodeDelta{NodeId: n.Id, Status: n.Status}
}

func PushOpChannel(n *Node, tm time.Duration) {
	OpEventChannel <- OpEvent{NodeId: n.Id, OpTime: tm}
}

func CreateNode(networkMsg chan protocol.Message, bufferMsg chan protocol.Message) *Node {
	generatedId := fmt.Sprint(uuid.New())
	return &Node{Id: generatedId, Status: 0, NetworkChannel: networkMsg, StopHeartBeat: make(chan bool, 1), bufferChannel: bufferMsg}

}

func (n *Node) RequestLock() {
	n.Mu.Lock()
	if n.Status != Idle {
		n.Mu.Unlock()
		return
	} //not idle

	n.Status = Request
	PushDeltaChannel(n)
	n.Mu.Unlock()
	fmt.Printf("[NODE %s] State changed to REQUESTING. Transmitting MsgRequest...\n", n.Id)

	//since its a chnnel variable, a message struct variable is created and is fed into nodes's networkchannel using channel initialization <-
	n.NetworkChannel <- protocol.Message{Id: fmt.Sprintf("req-%s-%d", n.Id, time.Now().UnixNano()),
		NodeId: n.Id, Type: protocol.MsgRequest,
	}

}

// status holding->idle
func (n *Node) ReleaseLock() {
	n.Mu.Lock()

	if n.Status != Holding {
		n.Mu.Unlock()
		return
	}

	n.Status = Idle
	PushDeltaChannel(n)
	n.Mu.Unlock()
	n.StopHeartBeat <- true
	fmt.Printf("[NODE %s] Yielding Resource. Transmitting MsgRelease...\n", n.Id)

	n.NetworkChannel <- protocol.Message{Id: fmt.Sprintf("req-%s-%d", n.Id, time.Now().UnixNano()),
		NodeId: n.Id, Type: protocol.MsgRelease,
	}
}

// to restart killed/crashed  nodes
// set them to idle state
func (n *Node) Restart() {
	n.Mu.Lock()
	if n.Status != Crashed {
		n.Mu.Unlock()
		return
	}

	//set all the members of that node to def
	n.Status = Idle
	PushDeltaChannel(n)
	//no cvhange in id
	n.Mu.Unlock()
	//the stophearbeat channel could be corrupted or may contain stale data=> make a new channel
	n.StopHeartBeat = make(chan bool)
	fmt.Printf("[NODE %s] Clean recovery successful. State transitioned from CRASHED -> IDLE. Ready for action.\n", n.Id)
}

// Kill siMulates a sudden hard cluster failure, immediately freezing operations
func (n *Node) Kill() {
	n.Mu.Lock()
	defer n.Mu.Unlock()

	if n.Status == Holding {
		// Suppress channel write blocks in case routine already terminated
		select {
		case n.StopHeartBeat <- true:
		default:
		}
	}
	n.Status = Crashed
	PushDeltaChannel(n)
	fmt.Printf("[NODE %s] !!! HARD SYSTEM CRASH SIMuLATED !!!\n", n.Id)
}

// StartHeartbeatLoop runs an out-of-band ticker routine only when state == HOLDING
func (n *Node) StartHeartbeatLoop() {
	n.Mu.Lock()
	if n.Status != Holding { //heartbeat starts only on hold
		n.Mu.Unlock()
		return
	}
	n.Mu.Unlock()
	fmt.Printf("heartbeat started for node %s", n.Id)

	go func() { //another backround routine that constantly sends heartbeat to manager inc or restart the lock lease timer
		//time.NewTicker creates a timer object that internally contains and manages a channel. That channel is accessible via ticker.C.
		ticker := time.NewTicker(1500 * time.Millisecond) // Heartbeat cadence , evry time the timer obj
		// reaches 1.5 sec, it passes a tick to the obj's channel
		defer ticker.Stop() //automatically done when the function is excited

		fmt.Printf("[NODE %s] Asynchronous Heartbeat Pipeline Active.\n", n.Id)

		for {
			select { //select waits for both cases, whichever case comes first in current iteration, corresponding code is executed
			case <-ticker.C:
				// Pulse heartbeat packet to network siMulator channel
				fmt.Println("heartbeat sent")
				fmt.Println()
				n.NetworkChannel <- protocol.Message{
					Id:     fmt.Sprintf("hb-%s-%d", n.Id, time.Now().UnixNano()),
					NodeId: n.Id,
					Type:   protocol.MsgHeartbeat,
					Token:  n.CurrentToken,
				}
			case <-n.StopHeartBeat:
				fmt.Printf("[NODE %s] Asynchronous Heartbeat Pipeline Safely Exited.\n", n.Id)
				return
			}
		}
	}()
}

// StartInboundConsumer spikes an infinite background thread loop for this specific node.
// It drains its private buffer channel and transitions the node's status based on manager messages.
func (n *Node) NodeBufferIteration() {
	go func() {
		fmt.Printf("[NODE %s] Inbound network consumer active. Monitoring mailbox buffer...\n", n.Id)

		// This loop blocks and sleeps until a new message packet lands in the channel buffer.
		// It automatically processes packets in strict FIFO order.
		for msg := range n.bufferChannel {

			// Grab the local node Mutex lock to update properties safely without race conditions
			n.Mu.Lock()

			// Edge-Case Guard: If this node is completely crashed/offline, ignore incoming packets
			if n.Status == Crashed {
				n.Mu.Unlock()
				continue
			}

			// Evaluate the packet depending on its protocol type
			switch msg.Type {
			case protocol.MsgGrant:
				// SUCCESS: The manager officially gave this node exclusive control of the lock!
				n.Status = Holding

				n.CurrentToken = msg.Token // Cache the incremented fencing token in local RAM
				fmt.Printf("[NODE %s]  LOCK GRANTED! State -> HOLDING. Fencing Token: %d\n", n.Id, msg.Token)
				PushDeltaChannel(n)
				n.Mu.Unlock()

				n.StartHeartbeatLoop()
				//a node that has lock must let go of it after some intended op
				//we do that by using timer ,after which the lock is released
				val := rand.IntN(15) + 5 //setting 5 as base
				duration := time.Duration(val) * time.Second
				fmt.Printf("[NODE %s] started performing operation, it will last for %d sec", n.Id[:5], val)
				PushOpChannel(n, duration)

				// It will wait for 'duration' to pass, then call n.ReleaseLock() automatically.
				time.AfterFunc(duration, func() {
					fmt.Printf("[NODE %s] Operation completed,  releasing lock...\n", n.Id[:5])
					n.ReleaseLock() // Call your node release method here
				})

			case protocol.MsgEvict:
				fmt.Println("inside message evict")
				if n.Status != Holding {
					n.Mu.Unlock()
					return
				}
				n.StopHeartBeat <- true
				n.Status = Idle
				PushDeltaChannel(n)
				n.AttemptCount++

				backoffSecs := 1 << uint(n.AttemptCount)
				if backoffSecs > 5 {
					backoffSecs = 5
				}
				n.Mu.Unlock()

				// select {
				// case :
				// 	// Signal successfully delivered if the heartbeat loop was listening
				// default:
				// 	// If the heartbeat loop was busy/blocked, we don't stall!
				// 	// Since StopHeartBeat has a buffer of 1, it will sit in the box
				// 	// and the heartbeat loop will catch it on its very next iteration loop.
				// }

				fmt.Printf("[NODE %s] Evicted! Attempt #%d. Exponential backoff cooling down for %ds...\n",
					n.Id, n.AttemptCount, backoffSecs)

				time.Sleep(time.Duration(backoffSecs) * time.Second)

				fmt.Printf("[NODE %s] Backoff finished. Retrying lock request...\n", n.Id)
				n.RequestLock()
			default:
				n.Mu.Unlock()
			}

			// Always release the lock at the end of the message processing cycle

		}
	}()
}
