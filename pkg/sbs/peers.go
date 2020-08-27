package sbs

import (
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/lucas-clemente/quic-go"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/proto"

	volumesAPI "github.com/tcfw/vpc/pkg/api/v1/volumes"
)

const (
	reconnectRetryDuration         = 5 * time.Second
	reconnectRetryCooldownDuration = 10 * time.Second
	negotationTimeout              = 5 * time.Second
	bootPeerRetryCount             = 5

	//PeerStatusConnected Peer is successfully connected
	PeerStatusConnected PeerStatus = iota

	//PeerStatusReconnecting peer is current being reconnected
	PeerStatusReconnecting
)

//PeerStatus connection status of peer
type PeerStatus uint32

//Peer remote volume peer
type Peer struct {
	PeerID     string
	RemoteAddr net.Addr
	Status     PeerStatus

	conn     quic.Session
	Channels map[string]*streamChannel

	mu sync.Mutex
}

//ID peer ID
func (p *Peer) ID() string {
	return p.PeerID
}

//Conn remote connection to the peer
func (p *Peer) Conn() quic.Session {
	return p.conn
}

type streamChannel struct {
	stream   quic.Stream
	syncBuff []byte
	srMu     sync.Mutex
}

func (c *streamChannel) Write(p []byte) (int, error) {
	return c.stream.Write(p)
}

func (c *streamChannel) Read(p []byte) (int, error) {
	return c.stream.Read(p)
}

//SyncRead allows for reading using a preallocated buffer
func (c *streamChannel) SyncRead() ([]byte, error) {
	c.srMu.Lock()
	defer c.srMu.Unlock()

	n, err := c.stream.Read(c.syncBuff)
	if err != nil {
		return nil, err
	}

	return c.syncBuff[:n], nil
}

//AddPeer connects with a remote peer
func (s *Server) AddPeer(peerID string, remoteAddr *net.UDPAddr) error {
	//Peer already exists
	if _, ok := s.peers[peerID]; ok {
		return nil
	}

	sess, err := s.connectPeer(peerID, remoteAddr)
	if err != nil {
		return err
	}

	return s.negotiate(sess, true)
}

//AddPeerWithRetry tries to connect with a peer for given number of times
func (s *Server) AddPeerWithRetry(peerID string, remoteAddr *net.UDPAddr, retryCount int) error {
	tries := retryCount

	for tries > 0 {
		if err := s.AddPeer(peerID, remoteAddr); err != nil {
			tries--
			s.log.WithError(err).WithFields(logrus.Fields{"triesLeft": tries, "peer": peerID}).Warnf("Trying again in %s", reconnectRetryDuration.String())
			time.Sleep(reconnectRetryDuration)
			continue
		}
		return nil
	}

	return fmt.Errorf("failed after %d tries", retryCount)
}

//BootPeers bootstraps a list of peers in the format peerID@addr:port
func (s *Server) BootPeers(peerInfo []string) {
	for _, peer := range peerInfo {
		peerParts := strings.Split(peer, "@")
		if len(peerParts) != 2 {
			s.log.WithFields(logrus.Fields{"peerinfo": peer}).Warn("Invalid boot peer info")
			continue
		}

		addr, err := net.ResolveUDPAddr("udp", peerParts[1])
		if err != nil {
			s.log.WithError(err).WithFields(logrus.Fields{"peerinfo": peer}).Warn("Failed to parse peer remote addr from info")
		}

		peerID := peerParts[0]

		s.log.WithFields(logrus.Fields{"peer": peerID, "addr": addr.String()}).Info("Added boot peer")
		go func() {
			if err := s.AddPeerWithRetry(peerID, addr, bootPeerRetryCount); err != nil {
				s.log.WithError(err).WithField("peer", peerID).Warn("Failed to connect with boot peer")
			}
		}()
	}
}

//discoverPeer uses the controller to find the endpoint of a peer and connect with it
func (s *Server) discoverPeer(peer string) (*Peer, error) {
	if peer, exists := s.peers[peer]; exists {
		return peer, nil
	}

	if s.controller == nil {
		return nil, fmt.Errorf("no controller to find peers with")
	}

	addr, err := s.controller.Discover(peer)
	if err != nil {
		return nil, err
	}

	if err := s.AddPeerWithRetry(peer, addr, 5); err != nil {
		return nil, err
	}

	return s.peers[peer], nil
}

//connectPeer opens up a QUIC session with the remote peer and
//sends a partial handshake
func (s *Server) connectPeer(peerID string, remoteAddr *net.UDPAddr) (quic.Session, error) {
	tls, err := tlsConfig()
	if err != nil {
		return nil, fmt.Errorf("failed tls config: %s", err)
	}

	if peer, exists := s.peers[peerID]; exists {
		return peer.conn, nil
	}

	sess, err := quic.DialAddr(remoteAddr.String(), tls, defaultQuicConfig())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to remote peer: %s", err)
	}

	if err := s.sendHello(sess, peerID); err != nil {
		return nil, fmt.Errorf("failed to send hello: %s", err)
	}

	return sess, nil
}

//reconnectPeer attempts to reconnect with the peer every x seconds
//if the peer reconnects on it's own the reconnect is cancelled
func (s *Server) reconnectPeer(p *Peer) {
	s.log.WithField("peer", p.PeerID).Warnf("Reconnecting with peer ")

	p.mu.Lock()
	defer p.mu.Unlock()

	p.conn = nil

	if p.Status != PeerStatusReconnecting {
		//Assuming the peer reconnected on their end
		return
	}

	remoteAddr, _ := net.ResolveUDPAddr("udp", p.RemoteAddr.String())
	sess, err := s.connectPeer(p.PeerID, remoteAddr)
	if err != nil {
		s.log.WithError(err).Warnf("Failed to reconnect with peer. Trying again in %s", reconnectRetryCooldownDuration.String())
		time.Sleep(reconnectRetryCooldownDuration)
		s.reconnectPeer(p)
		return
	}

	p.conn = sess
	if err := s.negotiate(p.conn, true); err != nil {
		s.removePeer(p)
	}
}

//disconnectPeer sends a disconnect to the peer before removing
//the peer from the store
func (s *Server) disconnectPeer(p *Peer) {
	p.conn.CloseWithError(errDisconnect, "")
	s.removePeer(p)
	s.log.WithFields(logrus.Fields{"peer": p.PeerID}).Info("diconnected from peer")
}

//removePeer deletes the peer from the peer store
func (s *Server) removePeer(p *Peer) {
	if _, ok := s.peers[p.PeerID]; ok {
		s.mu.Lock()
		defer s.mu.Unlock()

		p.conn = nil

		delete(s.peers, p.PeerID)
		s.log.WithFields(logrus.Fields{"peer": p.PeerID}).Debug("removed peer from store")
	}
}

//negotiate waits for a handshake from the initiator before accepting
//other streams
func (s *Server) negotiate(sess quic.Session, initiator bool) error {
	ctx, cancel := context.WithTimeout(sess.Context(), negotationTimeout)
	peerInfo := make(chan *volumesAPI.PeerNegotiate)
	errCh := make(chan error)

	go func() {
		defer cancel()

		stream, err := sess.AcceptUniStream(ctx)
		if err != nil {
			s.log.WithFields(logrus.Fields{"addr": sess.RemoteAddr().String()}).Error("negotiation with peer failed waiting for hello")
			errCh <- err
			return
		}

		reqBytes, err := ioutil.ReadAll(stream)
		if err != nil {
			s.log.WithError(err).Error("negotiation read failed")
			errCh <- err
			sess.CloseWithError(0, err.Error())
			return
		}

		req := &volumesAPI.PeerNegotiate{}
		if err := proto.Unmarshal(reqBytes, req); err != nil {
			s.log.WithError(err).Error("failed to unmarshal handshake")
			sess.CloseWithError(errHandshakeRejected, "Invalid req")
			errCh <- err
			return
		}

		peerInfo <- req
	}()

	select {
	case <-ctx.Done():
		s.log.WithError(ctx.Err()).Debug("negotiation timeout")
		return ctx.Err()
	case err := <-errCh:
		s.log.WithError(err).Warn("negotiation failed")
		return err
	case info := <-peerInfo:
		if info.YouAre != s.peerID {
			s.log.WithFields(logrus.Fields{"peer": info.IAm, "addr": sess.RemoteAddr().String()}).Warn("Invalid YouAre - peer rejected")
			sess.CloseWithError(errHandshakeRejected, "Invalid YouAre")
			return fmt.Errorf("peer rejected")
		}
		if !initiator {
			s.sendHello(sess, info.IAm)
		}

		var peer *Peer
		if existingPeer, ok := s.peers[info.IAm]; !ok {
			peer = &Peer{
				PeerID:     info.IAm,
				RemoteAddr: sess.RemoteAddr(),
				conn:       sess,
				Channels:   map[string]*streamChannel{},
				Status:     PeerStatusConnected,
			}

			s.mu.Lock()
			s.peers[peer.PeerID] = peer
			s.mu.Unlock()

			s.log.WithFields(logrus.Fields{"peer": peer.PeerID}).Info("Connected with peer")
		} else {
			peer = existingPeer
			peer.mu.Lock()
			if peer.Status != PeerStatusConnected {
				peer.conn = sess
				peer.Status = PeerStatusConnected
				s.log.WithField("peer", peer.PeerID).Info("Reconnected with peer")
			}
			peer.mu.Lock()
		}

		go s.handlePeer(peer)
	}

	return nil
}

func (s *Server) negotiateLocalPeer(sess quic.Session) {
	p := &Peer{
		PeerID:     s.PeerID(),
		RemoteAddr: sess.RemoteAddr(),
		conn:       sess,
		Channels:   map[string]*streamChannel{},
		Status:     PeerStatusConnected,
	}

	go s.handlePeer(p)
}

//sendHello sends a handshake request
func (s *Server) sendHello(sess quic.Session, remotePeerID string) error {
	resp := &volumesAPI.PeerNegotiate{
		IAm:    s.peerID,
		YouAre: remotePeerID,
	}

	respBytes, err := proto.Marshal(resp)
	if err != nil {
		s.log.WithError(err).Error("failed to marshal handshake resp")
		return err
	}

	sendStream, err := sess.OpenUniStream()
	if err != nil {
		s.log.WithError(err).Error("failed to open resp handshake")
		return err
	}
	if _, err := sendStream.Write(respBytes); err != nil {
		s.log.WithError(err).Error("failed to write resp handshake")
		return err
	}

	return sendStream.Close()
}

//GetPeer sends back a peer given it's ID. If the peer does not exist,
//it will try to discover it via the controller
func (s *Server) GetPeer(id string) *Peer {
	if peer, ok := s.peers[id]; ok {
		return peer
	}

	peer, err := s.discoverPeer(id)
	if err == nil {
		return peer
	}

	return nil
}

//GetPeers discovers and returns a set of peers
func (s *Server) GetPeers(ids []string) []*Peer {
	peers := []*Peer{}

	for _, id := range ids {
		var peer *Peer
		if p, ok := s.peers[id]; ok {
			peer = p
		} else {
			if p, err := s.discoverPeer(id); err == nil {
				peer = p
			}
		}
		if peer != nil {
			peers = append(peers, peer)
		}
	}

	return peers
}
