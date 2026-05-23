package startup

import (
	"distributed_lock_manager/internal/manager"
	"distributed_lock_manager/internal/network"
	"distributed_lock_manager/internal/protocol"
	"os"
	"time"
)

var (

	// Global channel to signal main.go to shut down
	ShutdownChan     chan os.Signal
	networkSimulator *network.NetworkSimulator
	lockManager      *manager.LockManager
)

func init() {
	ShutdownChan = make(chan os.Signal, 1)
	fromManagerChannel := make(chan protocol.Message, 250)
	toManagerChannel := make(chan protocol.Message, 250)

	//first create network
	networkSimulator = network.NewNetworkSimulator(toManagerChannel)
	//create lock manager
	lockManager = manager.NewLockManager(5*time.Second, fromManagerChannel)

	//now start listeing loops in both nw and manager
	go networkSimulator.StartPipeline()
	go networkSimulator.ReturnPipeline(fromManagerChannel)
	go lockManager.Start(toManagerChannel)
}

func PointerToNS() *network.NetworkSimulator {
	return networkSimulator
}
func PointerToLM() *manager.LockManager {
	return lockManager
}
