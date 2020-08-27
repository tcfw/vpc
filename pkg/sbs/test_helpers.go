package sbs

import (
	"crypto/sha256"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/hashicorp/raft"
	"github.com/tcfw/vpc/pkg/sbs/config"
	"github.com/tcfw/vpc/pkg/sbs/control"
)

//randomID generates a random ID to be used for a server
func randomID() string {
	h := sha256.New()
	h.Write([]byte(time.Now().String()))
	return fmt.Sprintf("%x", h.Sum(nil))
}

//getLeader waits for a nd gets the current raft leader
func getLeader(servers []*Server) *Server {
	var id string

upper:
	for {
		for _, srv := range servers {
			leaderID := srv.Volume("vol-test").Raft.Leader()
			if leaderID != "" {
				id = string(leaderID)
				break upper
			}
		}
		time.Sleep(10 * time.Millisecond)
	}

	for _, srv := range servers {
		if srv.PeerID() == id {
			return srv
		}
	}

	return nil
}

//getNotLeader waits for a leader and then returns the first server
//that is not the leader
func getNotLeader(servers []*Server) *Server {
	var id string

	//Actually find a leader or at least make sure one exists
upper:
	for {
		for _, srv := range servers {
			leaderID := srv.Volume("vol-test").Raft.Leader()
			if leaderID != "" {
				id = string(leaderID)
				break upper
			}
		}
		time.Sleep(10 * time.Millisecond)
	}

	for _, srv := range servers {
		if srv.PeerID() != id {
			return srv
		}
	}

	return nil
}

//waitForNPeers waits until the required number of raft peers is connected together
func waitForNPeers(n int, obsCh <-chan raft.Observation) {
	atleastOne := make(chan struct{})

	go func() {
		for ch := range obsCh {
			if !ch.Data.(raft.PeerObservation).Removed {
				atleastOne <- struct{}{}
			}
		}
	}()

	for i := n; i > 0; i-- {
		<-atleastOne
	}
}

//makePeerObserver constructs a new raft observer that is filtered to peer observations
func makePeerObserver() (<-chan raft.Observation, *raft.Observer) {
	obsCh := make(chan raft.Observation, 10)
	observer := raft.NewObserver(obsCh, false, func(o *raft.Observation) bool {
		_, isPeerOb := o.Data.(raft.PeerObservation)
		return isPeerOb
	})

	return obsCh, observer
}

//makeTestCluster creates a small set of servers on random ports and connects them all together
//in a star topology
func makeTestCluster(n int, controller control.Controller) ([]string, []*Server, error) {
	ids, servers, err := makeTestDisconnectedCluster(n, controller)
	if err != nil {
		return nil, nil, err
	}

	//Fully connect all peers
	for _, srv := range servers {
		for _, peerSrv := range servers {
			if srv.PeerID() == peerSrv.PeerID() {
				continue
			}
			if err := srv.AddPeer(peerSrv.PeerID(), peerSrv.LocalAddr()); err != nil {
				return nil, nil, err
			}
		}
	}

	return ids, servers, nil
}

//makeTestDisconnectedCluster creates a small set of servers on random ports, but does not connect them to each other
func makeTestDisconnectedCluster(n int, controller control.Controller) ([]string, []*Server, error) {
	ids := []string{}
	servers := []*Server{}

	for i := 0; i < n; i++ {
		id := randomID()
		srv := NewServerWithPort(rand.Intn(1750) + 34000)
		srv.SetPeerID(id)
		srv.SetController(controller)

		if err := srv.Listen(); err != nil {
			return nil, nil, err
		}

		ids = append(ids, id)
		servers = append(servers, srv)
	}

	return ids, servers, nil
}

//cleanUpServerData deletes all local data related to a test cluster
func cleanUpServerData(servers []*Server) {
	for _, serv := range servers {
		os.RemoveAll(fmt.Sprintf("%s/%s", config.BlockStoreDir(), serv.PeerID()))
	}
}
