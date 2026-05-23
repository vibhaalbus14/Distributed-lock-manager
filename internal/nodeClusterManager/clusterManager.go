package cluster_manager

import (
	"distributed_lock_manager/internal/node"
	"sync"
)

var AllNodes []*node.Node
var RegMu sync.Mutex
var HmNodes = make(map[string]*node.Node) //nodeid:pointer

func RegisterNode(n *node.Node) {

	RegMu.Lock()
	defer RegMu.Unlock() //locks are needed here because when the all nodes is  being updated, we may get a call from
	//frontend for stataus , then inconsistent info will be given
	//to forgo it we need locks
	AllNodes = append(AllNodes, n)
	HmNodes[n.Id] = n
}

func GetAllNodeStatus() map[string]node.NodeStatus {
	RegMu.Lock()
	defer RegMu.Unlock()
	//create a hm of id:status
	hm := make(map[string]node.NodeStatus)

	for _, np := range AllNodes {
		np.Mu.Lock()
		hm[np.Id] = np.Status
		np.Mu.Unlock()
	}

	return hm
}
