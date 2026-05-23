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
	DefaultDropRate = 0.0                     // 10% packet loss
	DefaultMinDelay = 0 * time.Millisecond    // 100ms floor
	DefaultMaxDelay = 1000 * time.Millisecond // 300ms ceiling
)

type NetworkEvent struct {
	Event   string
	Message string
}

var NetworkLogChannel = make(chan NetworkEvent, 100) //network channel , this is the conveyor belt from which
//messages are taken and given to client by websockets

// NetworkSimulator acts as the central hub and chaos engine.
type NetworkSimulator struct {
	dropRate float64
	minDelay time.Duration
	maxDelay time.Duration
	mu       sync.RWMutex // Protects configuration fields from concurrent frontend updates => readers-writer protocol
	//r-r =>allowed
	//r-w not allowed
	//w-r not allowd
	//w-w not allowed
	//for one write, all the readers count must be 0
	FromNodePipe  chan protocol.Message //  Central funnel nodes-> manager channel
	ToManagerPipe chan protocol.Message // Output pipe leading to Lock Manager
	NodeRegistry  map[string]chan protocol.Message
}

// NewNetworkSimulator initializes our chaos engine with safe fallback defaults
func NewNetworkSimulator(ToManagerPipe chan protocol.Message) *NetworkSimulator {
	return &NetworkSimulator{
		dropRate:      DefaultDropRate,
		minDelay:      DefaultMinDelay,
		maxDelay:      DefaultMaxDelay,
		FromNodePipe:  make(chan protocol.Message, 100), //current network pipe channel is of length 100=> max 100 packets allowed to stay
		ToManagerPipe: ToManagerPipe,
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
		fmt.Println("[NODE->NETWORK] pipeline online. Monitoring channel traffic...")
		//A for range loop on a channel will not spin wildly or waste CPU if the channel is empty.
		// It will process messages as they arrive, and if there are no messages,
		// it will safely freeze and put that specific background thread to sleep until a new message is dropped into the pipe.
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
				fmt.Printf("[NETWORK CHAOS] !!! PACKET DROPPED !!! Discarded %d from %s\n", msg.Type, msg.NodeId)
				//add this event to NewtworkLogChannel
				select {
				case NetworkLogChannel <- NetworkEvent{
					Event:   "PACKET_DROPPED",
					Message: fmt.Sprintf("Packet from %s sending status %s to Lock manager was dropped", msg.NodeId[:5], protocol.MsgMap[msg.Type]),
				}:
				default:
				}
				continue //move onto next message packet if avail
			}

			// 3. Process Latency Calculations
			randomDelay := currentMinDelay
			if currentMaxDelay > currentMinDelay {
				delta := currentMaxDelay - currentMinDelay
				randomDelay = currentMinDelay + time.Duration(rand.Int63n(int64(delta))) //rnadom number from [0,delta-1]
			}

			// 4. Asynchronous Delivery
			//why this routine?
			//every pcket will have a diff delay , thsu the packet is sent to manager only after its delay
			//but the parent routine cant wait till the delay is completed
			//so we spin off a new rotine/thread for every packet that runs in the background
			go func(m protocol.Message, delay time.Duration) {
				select {
				case NetworkLogChannel <- NetworkEvent{
					Event:   "PACKET_DELAYED",
					Message: fmt.Sprintf("Packet from %s sending status %s to Lock manager is delayed by %s", msg.NodeId[:5], protocol.MsgMap[msg.Type], delay),
				}:
				default:
				}
				time.Sleep(delay)
				ns.ToManagerPipe <- m //feed the entire message into manager's channel
			}(msg, randomDelay)
		}

	}() //since its an anonymous function,we make the call here as well
}

// ReturnPipeline runs an independent background loop to intercept manager-to-node traffic.
// It applies the exact same symmetrical chaos delays and drop rates before hitting the node buffers.
func (ns *NetworkSimulator) ReturnPipeline(FromManager chan protocol.Message) {
	go func() {
		fmt.Println("[MANAGER->NETWORK] return pipeline online. ")

		// 1. Pull the packet OUT of the manager's outbound channel wire
		for msg := range FromManager {

			// 2. Read configuration safely using Read Lock (RLock)
			ns.mu.RLock()
			currentDropRate := ns.dropRate
			currentMinDelay := ns.minDelay
			currentMaxDelay := ns.maxDelay
			ns.mu.RUnlock()

			// 3. PROCESS RETURN DROPS: Manager replies can get lost in flight too!
			if rand.Float64() < currentDropRate {
				fmt.Printf("[NETWORK CHAOS] !!! RETURN PACKET DROPPED !!! Lost %d destined for %s\n", msg.Type, msg.NodeId)

				select {
				case NetworkLogChannel <- NetworkEvent{
					Event:   "PACKET_DROPPED",
					Message: fmt.Sprintf("Packet from Lock manager sending status %s to %s was dropped", protocol.MsgMap[msg.Type], msg.NodeId[:5]),
				}:
				default:
				}

				continue // Packet is deleted, move to next message
			}

			// 4. PROCESS RETURN LATENCY: Calculate random lag within bounds
			randomDelay := currentMinDelay
			if currentMaxDelay > currentMinDelay {
				delta := currentMaxDelay - currentMinDelay
				randomDelay = currentMinDelay + time.Duration(rand.Int63n(int64(delta)))
			}

			// 5. Asynchronous Delayed Delivery to the specific Node Buffer
			// We spin up a goroutine per packet so lagging packets don't queue block each other!
			go func(m protocol.Message, delay time.Duration) {
				select {
				case NetworkLogChannel <- NetworkEvent{
					Event:   "PACKET_DELAYED",
					Message: fmt.Sprintf("Packet from lock manager sending status %s to %s is delayed by %s", protocol.MsgMap[msg.Type], msg.NodeId[:5], delay),
				}:
				default:
				}
				time.Sleep(delay) // Simulated return flight wire transit time

				ns.mu.RLock()
				// Look up the target node's buffer channel pointer in our clipboard map
				targetNodeBufferChannel, exists := ns.NodeRegistry[m.NodeId]
				ns.mu.RUnlock()

				if exists {
					// THE SYMMETRICAL HANDOFF: Drop packet into node buffer after the delay!
					targetNodeBufferChannel <- m
					fmt.Printf("[NETWORK] Delivered %d to Node %s buffer after %v delay.\n", m.Type, m.NodeId, delay)
				} else {
					fmt.Printf("[NETWORK Error] Routing failed. Node %s is not in the registry map!\n", m.NodeId)
				}
			}(msg, randomDelay)
		}
	}()
}

func (ns *NetworkSimulator) AddToRegistry(nodeId string, nodeBuffer chan protocol.Message) {
	ns.mu.Lock() // Grab write lock to modify the map safely
	defer ns.mu.Unlock()

	ns.NodeRegistry[nodeId] = nodeBuffer
	fmt.Printf("[NETWORK REGISTRY] Successfully mapped route for %s\n", nodeId)
}
