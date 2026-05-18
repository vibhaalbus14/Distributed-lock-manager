package network

import (
	"distributed_lock_manager/internal/protocol"
	"fmt"
	"math/rand"
	"sync" // Required for thread safety
	"time"
)

// Default Configuration Constants
const (
	DefaultDropRate = 0.10                  // 10% packet loss
	DefaultMinDelay = 100 * time.Millisecond // 100ms floor
	DefaultMaxDelay = 300 * time.Millisecond // 300ms ceiling
)

// NetworkSimulator acts as the central hub and chaos engine.
type NetworkSimulator struct {
	dropRate     float64
	minDelay     time.Duration
	maxDelay     time.Duration
	mu           sync.RWMutex         // Protects configuration fields from concurrent frontend updates => readers-writer protocol
	//r-r =>allowed
	//r-w not allowed
	//w-r not allowd
	//w-w not allowed
	//for one write, all the readers count must be 0
	FromNodePipe chan protocol.Message //  Central funnel nodes-> manager channel
	ToManagerPipe  chan protocol.Message // Output pipe leading to Lock Manager
	NodeRegistry map[string]chan protocol.Message

}

// NewNetworkSimulator initializes our chaos engine with safe fallback defaults
func NewNetworkSimulator(ToManagerPipe chan protocol.Message) *NetworkSimulator {
	return &NetworkSimulator{
		dropRate:     DefaultDropRate,
		minDelay:     DefaultMinDelay,
		maxDelay:     DefaultMaxDelay,
		FromNodePipe: make(chan protocol.Message, 100),//current network pipe channel is of length 100=> max 100 packets allowed to stay
		ToManagerPipe:  ToManagerPipe,
		NodeRegistry:  make(map[string]chan protocol.Message),
	}
}

// --- FRONTEND MUTATOR FUNCTIONS (SETTERS) ---
//why are locks needed on delays and drop rates?
//thses variables can be modified and the values are taken from frontend, thers a chance that the momnet im reading thses values, a new value from front end comes in and thus entire calulation is inconsistent, thus 
//mutex r/w locks are taken to forgo the plausible conflicts

// UpdateDropRate allows the React frontend to dynamically change packet loss intensity
func (ns *NetworkSimulator) UpdateDropRate(newRate float64) {
	ns.mu.Lock() // Exclusive Write Lock
	defer ns.mu.Unlock()

	// Safe bounding check
	if newRate < 0.0 {
		newRate = 0.0
	} else if newRate > 1.0 {
		newRate = 1.0
	}

	ns.dropRate = newRate
	fmt.Printf("[CONFIG] Network Drop Rate updated dynamically to: %.2f%%\n", newRate*100)
}

// UpdateLatencyBounds allows the React frontend to adjust network lag ranges in milliseconds
func (ns *NetworkSimulator) UpdateLatencyBounds(minMs, maxMs int) {
	ns.mu.Lock() // Exclusive Write Lock
	defer ns.mu.Unlock()

	minDuration := time.Duration(minMs) * time.Millisecond
	maxDuration := time.Duration(maxMs) * time.Millisecond

	// Sanity fail-safes
	if minDuration < 0 {
		minDuration = 0
	}
	if maxDuration < minDuration {
		maxDuration = minDuration // Max cannot be lower than min
	}

	ns.minDelay = minDuration
	ns.maxDelay = maxDuration
	fmt.Printf("[CONFIG] Network Latency updated dynamically to: %v - %v\n", minDuration, maxDuration)
}

// StartPipeline runs an infinite out-of-band loop to intercept traffic
func (ns *NetworkSimulator) StartPipeline() {
	go func() {
		fmt.Println("NETWORK manager pipeline online. Monitoring channel traffic...")

		for msg := range ns.FromNodePipe {
			
			// 1. Read configuration safely using Read Lock (RLock)
			ns.mu.RLock()
			currentDropRate := ns.dropRate
			currentMinDelay := ns.minDelay
			currentMaxDelay := ns.maxDelay
			ns.mu.RUnlock()

			// 2. Process Drops
			//since we dont know thw number of messages that are present in the network channel , will ocme or has come,
			//droprate => prob of individual packet being dropped
			//a random number is generated and if we get lucky and >=drop rate, it doesnt get dropped
			if rand.Float64() < currentDropRate {
				fmt.Printf("[NETWORK Layer] !!! PACKET DROPPED !!! Discarded %d from %s\n", msg.Type, msg.NodeId)
				continue //move onto next message packet if avail
			}

			// 3. Process Latency Calculations
			randomDelay := currentMinDelay
			if currentMaxDelay > currentMinDelay {
				delta := currentMaxDelay - currentMinDelay
				randomDelay = currentMinDelay + time.Duration(rand.Int63n(int64(delta)))//rnadom number from [0,delta-1]
			}

			// 4. Asynchronous Delivery
			//why this routine?
			//every pcket will have a diff delay , thsu the packet is sent to manager only after its delay
			//but the parent routine cant wait till the delay is completed
			//so we spin off a new rotine/thread for every packet that runs in the background
			go func(m protocol.Message, delay time.Duration) {
				time.Sleep(delay) 
				ns.ToManagerPipe <- m//feed the entire message into manager's channel
			}(msg, randomDelay)
		}

		

		
	}()//since its an anonymous function,we make the call here as well
}

// StartReturnPipeline runs an independent background loop to intercept manager-to-node traffic.
// It accepts the manager's outbound channel as an argument and routes packets to the correct node buffer.
func (ns *NetworkSimulator) ReturnPipeline(FromManager chan protocol.Message,nodeBuffer chan protocol.Message) {
	go func() {
		fmt.Println("[NETWORK] Outbound return pipeline online. Routing manager responses...")

		// 1. Pull the packet OUT of the manager's outbound channel
		for msg := range FromManager {
			
			// 5. Asynchronous Delivery to the specific Node Buffer
			go func(m protocol.Message) {
				ns.mu.RLock()
				// Look up the target node's buffer channel pointer in our clipboard map
				targetNodeBufferChannel := ns.NodeRegistry[m.NodeId]
				ns.mu.RUnlock()

					// THE HANDOFF: Push the message INTO the node's private buffer channel!
				targetNodeBufferChannel <- m 
				
			}(msg)
		}
	}()
}

func (ns *NetworkSimulator) AddToRegistry(nodeId string, nodeBuffer chan protocol.Message) {
    ns.mu.Lock() // Grab write lock to modify the map safely
    defer ns.mu.Unlock()

    ns.NodeRegistry[nodeId] = nodeBuffer
    fmt.Printf("[NETWORK REGISTRY] Successfully mapped route for %s\n", nodeId)
}