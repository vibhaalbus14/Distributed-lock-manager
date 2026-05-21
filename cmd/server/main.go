package main

// Package Manager imports Package Network (to read the incoming channel).

// Package Network imports Package Manager (to handle the return pipeline processing).

// The moment you type go build, the Go compiler looks at this loop, throws its hands up, and spits out a hard fatal error:
// import cycle not allowed. Go strictly forbids packages from importing each other because it makes it mathematically impossible
// for the compiler to determine which package to compile first.

//thats why bothe channels to and from manager are created outside and passed on as arguments

import (
	"distributed_lock_manager/internal/manager"
	"distributed_lock_manager/internal/network"
	"distributed_lock_manager/internal/node"
	"distributed_lock_manager/internal/protocol"
	"fmt"
	"time"
)

func main() {

	FromManagerChannel := make(chan protocol.Message)
	ToManagerChannel := make(chan protocol.Message)

	//first create network
	networkSimulator := network.NewNetworkSimulator(ToManagerChannel)
	//create lock manager
	lockManager := manager.NewLockManager(5*time.Second, FromManagerChannel)

	//now start listeing loops in both nw and manager
	networkSimulator.StartPipeline()
	networkSimulator.ReturnPipeline(FromManagerChannel)
	lockManager.Start(ToManagerChannel)

	var n *node.Node
	//now im creating 4 nodes
	for i := 0; i < 1; i++ {
		bufferChan := make(chan protocol.Message, 50)
		n = node.CreateNode(networkSimulator.FromNodePipe, bufferChan)
		networkSimulator.AddToRegistry(n.Id, bufferChan) //initialise node registry
		n.NodeBufferIteration()
		n.RequestLock()
	}

	//networkSimulator.UpdateDropRate(0.4)
	//networkSimulator.UpdateLatencyBounds(3, 5)
	timer := time.NewTimer(10 * time.Second)
	fmt.Println(("60 sec timer started , after this node will be killed"))
	defer timer.Stop() // Clean up the timer resources when the function exits

	val := <-timer.C

	// 2. Since it successfully unblocked, the timer definitely fired.
	// A zero time check ensures the value is valid.
	if !val.IsZero() {
		fmt.Printf("[NODE %s] Work simulation lease expired. Automatically releasing lock...\n", n.Id)

		// 3. Clear your state resources
		n.ReleaseLock()

		// 4. If you want to transition immediately to a new 10-second wait phase,
		// you must block on a fresh timer channel right here:
		cooldownTimer := time.NewTimer(10 * time.Second)
		defer cooldownTimer.Stop()

		<-cooldownTimer.C
		fmt.Printf("[NODE %s] Cooldown phase complete.\n", n.Id)
	} //as there are no options, go makes the thread in which main is running to sleep,so that our background theread can continue else everything will pause as soon as end of main thread is reached
	//when all the nodes are idle, the channels aka routines are asleeep, main is also asleep, so when go detects this dedlock, it forcefully exits the prog
}
