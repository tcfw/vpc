package sbs

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"

	"github.com/sirupsen/logrus"
	"github.com/tcfw/vpc/pkg/sbs/control"
	"github.com/tcfw/vpc/pkg/utils"
)

//NBDServer provides Linux NBD connectivity to a SBS cluster
type NBDServer struct {
	volumeServer *Server
	log          *logrus.Logger
	listenPort   int
	listener     net.Listener
	controller   control.Controller
	attachments  map[string]*Volume
	mu           sync.Mutex

	ctx context.Context

	sessCount sync.WaitGroup
}

//NewNBDServer provisions a new NBD listening server attached to the volume server
func NewNBDServer(s *Server, port int, c control.Controller) *NBDServer {
	nbd := &NBDServer{
		volumeServer: s,
		log:          utils.DefaultLogger(),
		listenPort:   port,
		controller:   c,
		attachments:  map[string]*Volume{},
		ctx:          context.Background(),
	}

	go nbd.Listen()

	return nbd
}

//Listen starts listening for and accepts new connections
func (s *NBDServer) Listen() error {
	nli, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", s.listenPort))
	if err != nil {
		return fmt.Errorf("failed to listen: %s", err)
	}

	defer nli.Close()

	s.listener = nli

	s.log.WithField("port", s.listenPort).Info("Listening for NBD connections")

	for {
		select {
		case <-s.ctx.Done():
			return nil
		default:
		}
		if conn, err := s.listener.Accept(); err != nil {
			if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
				continue
			}
			s.log.WithError(err).Error("failed to accept connection")
		} else {
			go s.accept(conn)
		}
	}
}

//accept new connections
func (s *NBDServer) accept(conn net.Conn) {
	s.sessCount.Add(1)
	defer s.sessCount.Done()

	sess := newSession(s, s.volumeServer, conn, s.log)
	sess.handle(s.ctx)
}

//Shutdown terminates the NBD listener
func (s *NBDServer) Shutdown() {

}

//Attach adds a volume as a NBD export
func (s *NBDServer) Attach(vol *Volume, peers []*Peer) error {
	if _, exists := s.attachments[vol.ID()]; exists {
		return errors.New("already exists")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.attachments[vol.ID()] = vol

	return nil
}
