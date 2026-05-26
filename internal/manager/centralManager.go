package manager

import (
	"distributed_lock_manager/internal/protocol"
	"fmt"
	"sync"
	"time"
)

// LockManager acts as our central, thread-safe Distributed Lock Coordinator
type LockManager struct {
	CurrentHolder       string                // ID of the node currently owning the lock
	WaitQueue           []string              // FIFO slice tracking queued node IDs
	FencingToken        int64                 // Monotonically increasing generation ID to stop stale writes
	LeaseDuration       time.Duration         // How long a lock stays valid without a heartbeat (e.g., 5s)
	LeaseTimer          *time.Timer           // pointer to timer obj,The underlying OS countdown timer
	Mu                  sync.Mutex            // Safeguards global manager variables from competing threads
	OutgoingNetworkPipe chan protocol.Message // Channel leading back out to the nodes (via network siMulation)
}

//create new channel indicating manger free

var ManagerFreeChan chan bool

func init() {
	ManagerFreeChan = make(chan bool, 100)
}

// NewLockManager initializes our central controller with strict defaults
func NewLockManager(leaseDuration time.Duration, OutgoingNetworkPipe chan protocol.Message) *LockManager {
	return &LockManager{
		CurrentHolder:       "",
		WaitQueue:           []string{},
		FencingToken:        0,
		LeaseDuration:       leaseDuration,
		OutgoingNetworkPipe: OutgoingNetworkPipe,
	}
}

// ProcessMessage accepts an incoming packet and handles it according to its type
func (lm *LockManager) ProcessMessage(msg protocol.Message) {
	lm.Mu.Lock()         // lock is taken while reaing all values of struct so that no two process will have conflict at the same time
	defer lm.Mu.Unlock() //this is done only after executing the mapped function

	// Exactly what you guessed! Going over every message depending on type:
	switch msg.Type {
	case protocol.MsgRequest:
		lm.handleRequest(msg.NodeId)

	case protocol.MsgRelease:
		lm.handleRelease(msg.NodeId)

	case protocol.MsgHeartbeat:
		lm.handleHeartbeat(msg.NodeId, msg.Token)
	}
}

// --- CORE STATE ROUTINES (Executed under lm.Mu Lock) ---

func (lm *LockManager) handleRequest(NodeId string) {
	if lm.CurrentHolder == "" {
		// The resource is completely free! Grant it immediately.
		lm.grantLock(NodeId)
	} else if lm.CurrentHolder == NodeId {
		// Idempotency guard: Node already holds it, do nothing.
		return
	} else {
		//  Resource is busy. Add the node to our FIFO wait list array.
		// for _, queuedID := range lm.WaitQueue {
		// 	if queuedID == NodeId {
		// 		return // Avoid duplicate queueing
		// 	}
		// }
		lm.WaitQueue = append(lm.WaitQueue, NodeId)
		fmt.Printf("[MANAGER] Lock busy. Node %s appended to WaitQueue: %v\n", NodeId, lm.WaitQueue)
	}
}

func (lm *LockManager) handleRelease(NodeId string) {
	if lm.CurrentHolder != NodeId {
		// Fencing Protection: A late/stale node is trying to release a lock it doesn't own! Ignore it.
		return
	}

	fmt.Printf("[MANAGER] Node %s successfully released the lock.\n", NodeId)

	lm.stopLeaseTimer()
	lm.rotateLock()

}

func (lm *LockManager) handleHeartbeat(NodeId string, tokenRecieved int64) {
	if tokenRecieved != lm.FencingToken {
		// Ignore heartbeats from late/expired nodes
		return
	}

	// Postpone the eviction countdown clock! Reset it back to 5 seconds.
	lm.stopLeaseTimer()
	lm.startLeaseTimer(NodeId)

}

// --- INTERNAL CONCURRENCY UTILITIES ---

func (lm *LockManager) grantLock(nodeID string) {
	lm.CurrentHolder = nodeID
	lm.FencingToken++ // Increment generation sequence number

	fmt.Printf("[MANAGER] Lock GRANTED to %s. Fencing Token: %d\n", nodeID, lm.FencingToken)

	// Start the countdown timer on the manager's desk
	lm.startLeaseTimer(nodeID)

	// Dispatch an asynchronous GRANT message back into the network wire
	go func(target string, token int64) {
		lm.OutgoingNetworkPipe <- protocol.Message{
			NodeId: target,
			Type:   protocol.MsgGrant,
			Token:  token,
		}
	}(nodeID, lm.FencingToken)
}

func (lm *LockManager) rotateLock() {
	if len(lm.WaitQueue) > 0 {
		// Pop the first node from our FIFO slice queue
		nextNode := lm.WaitQueue[0]
		lm.WaitQueue = lm.WaitQueue[1:]

		// Pass ownership to the next candidate in line
		lm.grantLock(nextNode)
	} else {
		lm.CurrentHolder = ""
		fmt.Println("[MANAGER] Lock is completely free. No nodes are waiting.")
		ManagerFreeChan <- true
		//pass this message to opEvent channel
	}
}

func (lm *LockManager) startLeaseTimer(nodeID string) {
	// time.AfterFunc spawns a background thread that sleeps for 5 seconds.
	// If it isn't stopped before the timer pops, it triggers the eviction code!
	fmt.Println(("ticker start"))
	fmt.Println()
	lm.LeaseTimer = time.AfterFunc(lm.LeaseDuration, func() {
		lm.Mu.Lock()
		defer lm.Mu.Unlock()

		// Safety double-check: Verify the node hasn't changed or released in those 5s

		fmt.Printf("[MANAGER ] !!! LEASE EXPIRED !!! Node %s missed heartbeats. Forcibly evicting...\n", nodeID)
		go func(target string) {
			lm.OutgoingNetworkPipe <- protocol.Message{
				NodeId: target, // Destination node ID
				Type:   protocol.MsgEvict,
			}
		}(nodeID)
		lm.rotateLock() // Strip lock away and promote the next queued node!

	})
}

func (lm *LockManager) stopLeaseTimer() {
	if lm.LeaseTimer != nil {
		lm.LeaseTimer.Stop()
		fmt.Println(("ticker stopped"))
	}
}

// Start boots up the central manager's consumer loop.
// It reads sequentially from the incoming network pipeline and processes each message.
func (lm *LockManager) Start(incomingPipe chan protocol.Message) {
	// We run this in a background goroutine so it doesn't freeze main.go
	go func() {
		fmt.Println("[MANAGER] Lock Coordinator online")
		fmt.Println("NETWORK->MANAGER online , Listening for incoming node packets...")

		// This is the missing piece! It loops over the channel indefinitely.
		for msg := range incomingPipe {
			// Pass the message directly into your thread-safe switchboard!
			lm.ProcessMessage(msg)
		}
	}()
}
