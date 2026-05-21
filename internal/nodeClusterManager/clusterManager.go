package cluster_manager

import (
	"distributed_lock_manager/internal/node"
	"sync"
)

var AllNodes []*node.Node
var regMu sync.Mutex

func RegisterNode(n *node.Node) {

	regMu.Lock()
	defer regMu.Unlock() //locks are needed here because when the all nodes is  being updated, we may get a call from
	//frontend for stataus , then inconsistent info will be given
	//to forgo it we need locks
	AllNodes = append(AllNodes, n)
}

func GetAllNodeStatus() map[string]node.NodeStatus {
	regMu.Lock()
	defer regMu.Unlock()
	//create a hm of id:status
	hm := make(map[string]node.NodeStatus)

	for _, np := range AllNodes {
		np.Mu.Lock()
		hm[np.Id] = np.Status
		np.Mu.Unlock()
	}

	return hm
}
