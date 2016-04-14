package network

import (
	"time"

	"github.com/josephlewis42/historia/storage"
)

const (
	heartbeatTimeout = 3 * time.Second
)

type Quorum struct {
	net   *Network
	alive map[network.Node]time.Time
	store storage.Storage
}

func New(network *Network, store storage.Storage) *Quorum {
	q := Quorum{
		net:   network,
		alive: make(map[Node]time.Time),
		store: store,
	}

	go q.Run()
	return &q
}

func (this *Quorum) Run() {
	for _, node := range this.net.GetNodes() {
		n.alive[node] = time.Unix(0, 0)
	}

}

func (n *Quorum) GetAliveHosts() []Node {
	var keys []Node

	currentTime := time.Now()

	for node, lastseen := range n.alive {
		if currentTime.Sub(lastseen) < heartbeatTimeout {
			keys = append(keys, node)
		}
	}

	return keys
}

/**
// Attempts to get a list of all hosts, if they aren't all alive
// we abort
func (n *Network) GetAllHosts() ([]Node, error) {
	alive := n.GetAliveHosts()

	if len(alive) == len(n.hosts) {
		return alive, nil
	}

	return nil, NotEnoughAliveHosts
}

func (n *Network) GetQuorum() ([]Node, error) {
	alive := n.GetAliveHosts()

	// We get a 2/3 quorum
	// TODO, do this more intelligently by ranzoming other hosts and always
	// choosing us
	if len(alive) >= n.QuorumSize() {
		return alive[:n.QuorumSize()], nil
	}

	return nil, NotEnoughAliveHosts
}

func (n *Network) QuorumSize() int {
	return int(math.Ceil((2.0 / 3.0) * float64(len(n.hosts))))
}

// Sends heartbeat messages to tell others we're alive.
func (n *Network) heartbeat() {
	for {
		for k, _ := range n.alive {
			nm := NetworkPacket{Type: networkHeartbeat}
			n.send(k, nm)
		}

		time.Sleep(1 * time.Second)
	}
}
**/
