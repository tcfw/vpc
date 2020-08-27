package sbs

import (
	"net"
	"os"
	"sync"

	"github.com/lucas-clemente/quic-go"
	"github.com/sirupsen/logrus"

	volumesAPI "github.com/tcfw/vpc/pkg/api/v1/volumes"
	"github.com/tcfw/vpc/pkg/sbs/control"
	"github.com/tcfw/vpc/pkg/utils"
)

const (
	//DefaultListenPort main peer listening port
	DefaultListenPort = 32546
)

//Server Simple Block Storage server instance
type Server struct {
	listenPort int
	pConn      net.PacketConn
	listener   quic.Listener
	controller control.Controller
	peerID     string

	peers   map[string]*Peer
	volumes map[string]*Volume

	mu         sync.Mutex
	connCount  sync.WaitGroup
	log        *logrus.Logger
	shutdownCh chan struct{}

	nbd *NBDServer
}

//NewServer constructs a new SBS server
func NewServer() *Server {
	s := &Server{
		peerID:     utils.MachineID(),
		listenPort: DefaultListenPort,
		log:        utils.DefaultLogger(),
		peers:      map[string]*Peer{},
		volumes:    map[string]*Volume{},
		shutdownCh: make(chan struct{}),
	}

	return s
}

//NewServerWithPort constructs a new SBS server on a specific listening port
func NewServerWithPort(port int) *Server {
	s := NewServer()
	s.listenPort = port

	return s
}

//SetPeerID overwrites the local peer id
func (s *Server) SetPeerID(id string) {
	s.peerID = id
}

//PeerID returns the current peer ID of the server
func (s *Server) PeerID() string {
	return s.peerID
}

//SetController provides the server with a controller for volume assignments
func (s *Server) SetController(cont control.Controller) {
	s.controller = cont
}

//Shutdown gracefully
func (s *Server) Shutdown() {
	s.log.Warnln("Starting graceful shutdown. Press ctrl+c again to force shutdown")
	go func() {
		utils.BlockUntilSigTerm()
		os.Exit(1)
	}()

	//NBD first
	if s.nbd != nil {
		s.nbd.Shutdown()
	}

	//Tell peers we're shutting down
	s.log.Debug("removing peers from store via disconnect")
	for _, peer := range s.peers {
		s.disconnectPeer(peer)
	}

	s.log.Debug("closing all block stores")
	for _, vol := range s.volumes {
		vol.Shutdown()
	}

	close(s.shutdownCh)

	s.log.Infoln("Bye!")
}

//PeerIDs provides a list of peer IDs
func (s *Server) PeerIDs() []string {
	ids := make([]string, 0, len(s.peers))
	for id := range s.peers {
		ids = append(ids, id)
	}
	return ids
}

//LocalAddr provides the server listening address
func (s *Server) LocalAddr() *net.UDPAddr {
	return s.listener.Addr().(*net.UDPAddr)
}

//Volume returns the volume associated with the volume ID
func (s *Server) Volume(id string) *Volume {
	if vol, ok := s.volumes[id]; ok {
		return vol
	}

	return nil
}

//AddVolume attaches a volume to the server
func (s *Server) AddVolume(d *volumesAPI.Volume, placementPeers []string) error {
	if _, ok := s.volumes[d.Id]; ok {
		return nil
	}

	vol, err := NewVolume(s, d, placementPeers)
	if err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.volumes[d.Id] = vol

	return nil
}

//Export exposes a volume directly via NBD without attachment
func (s *Server) Export(volume string) {
	if s.nbd == nil {
		s.nbd = NewNBDServer(s, NBDDefaultPort, s.controller)
	}

	peers := s.GetPeers(s.PeerIDs())

	if vol, isLocal := s.volumes[s.volumes[volume].ID()]; isLocal {
		err := s.nbd.Attach(vol, peers)
		if err != nil {
			s.log.WithError(err).Error("failed to attach volume")
		}

		s.log.WithField("vol", volume).Info("Started NBD exposure")
		return
	}

	s.log.WithField("vol", volume).Warn("Refusing to export a volume we're not a member of")
}

func (s *Server) newLocalConn() (*Peer, error) {

	tlsconfig, _ := tlsConfig()
	sess, err := quic.DialAddr(s.LocalAddr().String(), tlsconfig, defaultQuicConfig())
	if err != nil {
		return nil, err
	}

	p := &Peer{
		PeerID:     s.peerID,
		Status:     PeerStatusConnected,
		RemoteAddr: s.LocalAddr(),
		Channels:   map[string]*streamChannel{},
		conn:       sess,
	}

	return p, nil
}
