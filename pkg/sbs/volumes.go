package sbs

import (
	"fmt"
	"sync"

	badger "github.com/dgraph-io/badger/v2"
	badgerOptions "github.com/dgraph-io/badger/v2/options"
	"github.com/hashicorp/raft"
	"github.com/sirupsen/logrus"

	volumesAPI "github.com/tcfw/vpc/pkg/api/v1/volumes"
	"github.com/tcfw/vpc/pkg/sbs/config"
	"github.com/tcfw/vpc/pkg/utils"
)

//Volume instance
type Volume struct {
	id             string
	PlacementPeers map[string]struct{}
	Blocks         *BlockStore
	Raft           *raft.Raft
	Transport      *RaftTransport
	server         *Server
	desc           *volumesAPI.Volume
	logStore       *badger.DB

	mu  sync.Mutex
	log *logrus.Logger
}

//NewVolume constructs a raft-backed distributed volume
func NewVolume(s *Server, d *volumesAPI.Volume, peers []string) (*Volume, error) {
	transport := NewRaftTransport(s, d.Id)
	bs := NewBlockStore(s.PeerID(), d, config.BlockStoreDir(), utils.DefaultLogger())

	options := badger.DefaultOptions(fmt.Sprintf("%s/log", bs.BaseDir))
	options.SyncWrites = false
	options.ZSTDCompressionLevel = 0
	options.BlockSize = 16 * 1 << 10
	options.WithCompression(badgerOptions.None)
	options.WithMaxCacheSize(6400)
	options.WithMaxBfCacheSize(6400)
	logBadger, err := badger.Open(options)
	if err != nil {
		return nil, err
	}
	// logStore := NewLogStore(d.Id, logBadger)
	logStore := raft.NewInmemStore()
	cachedLogStore, err := raft.NewLogCache(3000, logStore)
	if err != nil {
		return nil, err
	}

	r, err := NewRaft(s, bs, cachedLogStore, transport)
	if err != nil {
		return nil, err
	}

	vol := &Volume{
		id:             d.Id,
		PlacementPeers: map[string]struct{}{},
		desc:           d,
		Blocks:         bs,
		Transport:      transport,
		server:         s,
		logStore:       logBadger,
		Raft:           r,
		log:            utils.DefaultLogger(),
	}

	if err := vol.init(peers); err != nil {
		vol.Raft.Shutdown()
		return nil, err
	}

	go vol.watchLeaderChange()

	return vol, nil
}

//ID of the volume
func (v *Volume) ID() string {
	return v.id
}

//Size of the volume in GB
func (v *Volume) Size() int64 {
	return v.desc.Size
}

func (v *Volume) watchLeaderChange() {
	for l := range v.Raft.LeaderCh() {
		if l {
			v.log.Info("Became leader of ", v.id)
		} else {
			v.log.Info("Lost leadership of ", v.id)
		}
	}
}

//Shutdown closes the volume
func (v *Volume) Shutdown() error {
	v.Raft.Shutdown().Error()
	v.logStore.Close()
	v.Blocks.Close()
	return nil
}

func (v *Volume) init(peers []string) error {
	bootConfig := raft.Configuration{
		Servers: []raft.Server{{
			Suffrage: raft.Voter,
			ID:       raft.ServerID(v.server.PeerID()),
			Address:  raft.ServerAddress(v.server.PeerID()),
		}},
	}

	for _, peer := range peers {
		v.server.GetPeer(peer)
		if peer == v.server.PeerID() {
			continue
		}

		v.PlacementPeers[peer] = struct{}{}

		bootConfig.Servers = append(bootConfig.Servers, raft.Server{
			Suffrage: raft.Voter,
			ID:       raft.ServerID(peer),
			Address:  raft.ServerAddress(peer),
		})

		v.log.WithField("vol", v.id).Debugf("added peer %s", peer)
	}

	//Ignore bootstrap errors
	v.Raft.BootstrapCluster(bootConfig).Error()

	return nil
}

func (v *Volume) writePTable() error {

	return nil
}

//AddPeer adds a placment peer where the volume is expected to reside
func (v *Volume) AddPeer(p *Peer) error {
	v.mu.Lock()
	defer v.mu.Unlock()

	f := v.Raft.AddVoter(raft.ServerID(p.ID()), raft.ServerAddress(""), 0, 0)
	if err := f.Error(); err != nil {
		return err
	}

	v.PlacementPeers[p.ID()] = struct{}{}
	return nil
}

//RemovePeer adds a placment peer where the volume is expected to reside
func (v *Volume) RemovePeer(p *Peer) error {
	v.mu.Lock()
	defer v.mu.Unlock()

	f := v.Raft.RemovePeer(raft.ServerAddress(p.ID()))
	if err := f.Error(); err != nil {
		return err
	}

	delete(v.PlacementPeers, p.ID())
	return nil
}
