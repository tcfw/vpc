package sbs

import (
	"context"
	"fmt"
	"io"
	"net"
	"reflect"
	"strings"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/kpango/fastime"
	"github.com/lucas-clemente/quic-go"

	"github.com/sirupsen/logrus"

	volumesAPI "github.com/tcfw/vpc/pkg/api/v1/volumes"
)

const (
	peerMsgBufSize = 18 * 1024
	rpcTimeout     = 25 * time.Second
)

//handlePeer listens to new peer streams after negotiation
func (s *Server) handlePeer(p *Peer) {
	s.connCount.Add(1)
	defer s.connCount.Done()

	for {
		stream, err := p.conn.AcceptStream(context.Background())
		if err != nil {
			if err.(net.Error).Timeout() {
				p.Status = PeerStatusReconnecting
				s.reconnectPeer(p)
				return
			} else if qErr, ok := err.(quicError); ok && qErr.IsApplicationError() {
				s.log.WithField("peer", p.PeerID).Warn("Peer gracefully shutting down")
				s.removePeer(p)
				return
			}

			s.log.WithError(err).Error("failed to accept new streams")

			//Wait a little while before resuming
			time.Sleep(5 * time.Second)
			continue
		}

		go s.handlePeerStream(p, stream)
	}
}

//handlePeerStream handles messages sent via a single peer stream
func (s *Server) handlePeerStream(p *Peer, stream quic.Stream) {
	buff := make([]byte, peerMsgBufSize)
	defer stream.Close()

	for {
		n, err := stream.Read(buff)
		if err != nil {
			streamErr, isStreamError := err.(quic.StreamError)
			netErr, isNetError := err.(net.Error)
			if isStreamError && streamErr.Canceled() {
				s.log.Trace("stream remotely closed")
				return
			}
			if isNetError && netErr.Timeout() {
				s.log.Debug("stream timeout")
				return
			}
			if err == io.EOF || strings.Contains(err.Error(), "Application error") {
				s.log.Trace("stream closed")
				return
			}

			s.log.WithError(err).WithFields(logrus.Fields{"peer": p.PeerID}).Warn("read failure")
			continue
		}

		msg := &volumesAPI.PeerRPC{}

		if err := proto.Unmarshal(buff[:n], msg); err != nil {
			s.log.WithError(err).WithFields(logrus.Fields{"peer": p.PeerID}).Warn("msg marshal failed")
			continue
		}

		s.log.WithFields(logrus.Fields{"peer": p.PeerID, "async": msg.Async}).Trace("rpc from peer")

		if msg.Async {
			go s.handleRPC(p, stream, msg)
			continue
		}

		if err := s.handleRPC(p, stream, msg); err != nil {
			s.log.WithError(err).WithFields(logrus.Fields{"peer": p.PeerID}).Warn("failed rpc handle")
		}
	}
}

func (s *Server) handleRPC(p *Peer, stream quic.Stream, msg *volumesAPI.PeerRPC) error {
	switch msgType := msg.RPC.(type) {
	case *volumesAPI.PeerRPC_RaftRPC:
		return s.handleRaftRPC(p, stream, msg.GetRaftRPC())
	case *volumesAPI.PeerRPC_PeerPing:
		return s.handlePing(p, stream, msg.GetPeerPing())
	case *volumesAPI.PeerRPC_BlockCommand:
		return s.handleBlockCommand(p, stream, msg.GetBlockCommand(), msg)
	default:
		return fmt.Errorf("unknown msg type from peer (%s): %s", p.PeerID, msgType)
	}
}

//SendMsg send a protobuf msg into a stream
func (s *Server) SendMsg(stream quic.Stream, rpc proto.Message) error {
	msg := rpc
	if _, isRPC := rpc.(*volumesAPI.PeerRPC); !isRPC {
		msg = s.rpcReq(rpc)
		if msg == nil {
			return fmt.Errorf("trying to send unknown rpc msg type %s", reflect.TypeOf(rpc))
		}
	}

	msgBytes, err := proto.Marshal(msg)
	if err != nil {
		return err
	}

	_, err = stream.Write(msgBytes)
	return err
}

//sendReq sends a single protobuf message and waits for a response
func (s *Server) sendReq(p *Peer, m proto.Message) (proto.Message, error) {
	if p.conn == nil || p.Status > PeerStatusConnected {
		return nil, fmt.Errorf("peer not connected")
	}

	msg := m

	if _, isRPC := m.(*volumesAPI.PeerRPC); !isRPC {
		msg = s.rpcReq(m)
		if msg == nil {
			return nil, fmt.Errorf("trying to send unknown rpc msg type %s", reflect.TypeOf(m))
		}
	}

	msgBytes, err := proto.Marshal(msg)
	if err != nil {
		return nil, err
	}

	errCh := make(chan error)
	respCh := make(chan *volumesAPI.PeerRPC)
	ctx, cancel := context.WithTimeout(context.Background(), rpcTimeout)

	go func() {
		defer cancel()
		stream, err := p.conn.OpenStream()
		if err != nil {
			errCh <- fmt.Errorf("open stream failed: %s", err)
			return
		}

		defer stream.Close()

		if _, err := stream.Write(msgBytes); err != nil {
			errCh <- fmt.Errorf("msg write failed: %s", err)
			return
		}

		buf := make([]byte, peerMsgBufSize)
		n, err := stream.Read(buf)
		if err != nil {
			errCh <- fmt.Errorf("failed resp read: %s", err)
			return
		}

		resp := &volumesAPI.PeerRPC{}
		if err := proto.Unmarshal(buf[:n], resp); err != nil {
			errCh <- err
			return
		}

		respCh <- resp
	}()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case err := <-errCh:
		return nil, err
	case resp := <-respCh:
		return s.rpcResp(resp), nil
	}
}

//SendReqViaChannel sends via a channel/persistent stream
//if the channel doesn't exist, it will be created
func (s *Server) SendReqViaChannel(p *Peer, channel string, m proto.Message) (proto.Message, error) {
	msg := m
	if _, isRPC := m.(*volumesAPI.PeerRPC); !isRPC {
		msg = s.rpcReq(m)
		if msg == nil {
			return nil, fmt.Errorf("recv unknown rpc msg type")
		}
	}

	chanStream, ok := p.Channels[channel]
	if !ok {
		stream, err := p.conn.OpenStream()
		if err != nil {
			return nil, err
		}

		peerChannel := &streamChannel{
			syncBuff: make([]byte, peerMsgBufSize),
			stream:   stream,
		}

		p.mu.Lock()
		p.Channels[channel] = peerChannel
		p.mu.Unlock()
		chanStream = peerChannel
	}

	if err := s.SendMsg(chanStream.stream, msg); err != nil {
		return nil, err
	}

	buff, err := chanStream.SyncRead()
	if err != nil {
		return nil, fmt.Errorf("failed resp read: %s", err)
	}

	resp := &volumesAPI.PeerRPC{}

	if err := proto.Unmarshal(buff, resp); err != nil {
		return nil, err
	}

	return s.rpcResp(resp), nil
}

//rpcReq converts a supported RPC to a PeerRPC msg
func (s *Server) rpcReq(m proto.Message) *volumesAPI.PeerRPC {
	rpc := &volumesAPI.PeerRPC{}

	switch m.(type) {
	case *volumesAPI.PeerPing:
		rpc.RPC = &volumesAPI.PeerRPC_PeerPing{PeerPing: m.(*volumesAPI.PeerPing)}
	case *volumesAPI.RaftRPC:
		rpc.RPC = &volumesAPI.PeerRPC_RaftRPC{RaftRPC: m.(*volumesAPI.RaftRPC)}
	case *volumesAPI.BlockCommandRequest:
		rpc.RPC = &volumesAPI.PeerRPC_BlockCommand{BlockCommand: m.(*volumesAPI.BlockCommandRequest)}
	case *volumesAPI.BlockCommandResponse:
		rpc.RPC = &volumesAPI.PeerRPC_BlockCommandResponse{BlockCommandResponse: m.(*volumesAPI.BlockCommandResponse)}
	default:
		return nil
	}

	return rpc
}

//rpcResp converts a PeerRPC msg to a supported RPC
func (s *Server) rpcResp(m *volumesAPI.PeerRPC) proto.Message {
	switch m.RPC.(type) {
	case *volumesAPI.PeerRPC_PeerPing:
		return m.GetPeerPing()
	case *volumesAPI.PeerRPC_RaftRPC:
		return m.GetRaftRPC()
	case *volumesAPI.PeerRPC_BlockCommand:
		return m.GetBlockCommand()
	case *volumesAPI.PeerRPC_BlockCommandResponse:
		return m.GetBlockCommandResponse()
	default:
		return nil
	}
}

//handlePing pings back
func (s *Server) handlePing(p *Peer, stream quic.Stream, ping *volumesAPI.PeerPing) error {
	return s.SendMsg(stream, s.rpcReq(&volumesAPI.PeerPing{Ts: fastime.Now().Unix()}))
}

//handleRaftRPC forwards raft RPC to relevant volume raft
func (s *Server) handleRaftRPC(p *Peer, stream quic.Stream, rpc *volumesAPI.RaftRPC) error {
	vol, ok := s.volumes[rpc.Volume]
	if !ok {
		return fmt.Errorf("recv raft cmd for unknown volume %s", rpc.Volume)
	}

	return vol.Transport.HandleCommand(stream, rpc)
}

func (s *Server) handleBlockCommand(p *Peer, stream quic.Stream, rpc *volumesAPI.BlockCommandRequest, msg *volumesAPI.PeerRPC) error {
	vol, ok := s.volumes[rpc.Volume]
	if !ok {
		return fmt.Errorf("recv raft cmd for unknown volume %s", rpc.Volume)
	}

	isLeader := vol.Raft.VerifyLeader().Error() == nil
	requiresLeader := rpc.Cmd == volumesAPI.BlockCommandRequest_WRITE ||
		rpc.Cmd == volumesAPI.BlockCommandRequest_ZERO

	//
	if ((msg.Metadata != nil && msg.Metadata.LeaderOnly) || requiresLeader) && !isLeader {
		leaderTries := 5
		var leader string
		for leaderTries > 0 {
			leader = string(vol.Raft.Leader())
			if leader != "" {
				break
			}
			leaderTries--
			time.Sleep(10 * time.Millisecond)
		}
		if leader == "" {
			s.log.Info("failed to find leader to forward to")
			return s.SendMsg(stream, &volumesAPI.BlockCommandResponse{Volume: vol.ID(), Error: "no leader to forward to yet"})
		}
		if !msg.Metadata.AllowForwarding {
			return s.SendMsg(stream, &volumesAPI.BlockCommandResponse{Volume: vol.ID(), RetryAt: leader})
		}

		msg.Metadata.TTL--
		if msg.Metadata.TTL <= 0 {
			return fmt.Errorf("TTL expired")
		}

		peer, err := s.discoverPeer(leader)
		if err != nil {
			return fmt.Errorf("failed to forward to leader - not in peer store")
		}

		resp, err := s.SendReqViaChannel(peer, "forward", msg)
		if err != nil {
			return err
		}

		err = s.SendMsg(stream, s.rpcReq(resp))
		if err != nil {
			s.SendMsg(stream, s.rpcReq(&volumesAPI.BlockCommandResponse{Volume: vol.ID(), Error: err.Error()}))
		}
		return err
	}

	switch rpc.Cmd {
	case volumesAPI.BlockCommandRequest_READ:
		buf := make([]byte, rpc.Length)
		n, err := vol.Blocks.ReadAt(buf, uint64(rpc.Offset))
		if err != nil {
			return err
		}

		resp := &volumesAPI.BlockCommandResponse{Volume: vol.ID(), Data: buf[:n]}
		return s.SendMsg(stream, resp)
	case volumesAPI.BlockCommandRequest_WRITE, volumesAPI.BlockCommandRequest_ZERO:
		logCmd := &StoreCommand{
			Op:     OpWrite,
			Offset: uint64(rpc.Offset),
			Length: uint32(rpc.Length),
			Data:   rpc.Data,
		}
		if rpc.Cmd == volumesAPI.BlockCommandRequest_ZERO {
			logCmd.Op = OpZeros
			logCmd.Data = nil
		}

		logCmdBytes, _ := logCmd.Encode()

		apply := vol.Raft.Apply(logCmdBytes, 20*time.Second)
		err := apply.Error()
		if err != nil {
			s.SendMsg(stream, s.rpcReq(&volumesAPI.BlockCommandResponse{Volume: vol.ID(), Error: err.Error()}))
			return err
		}
		resp := apply.Response()
		if err, ok := resp.(error); ok {
			s.SendMsg(stream, s.rpcReq(&volumesAPI.BlockCommandResponse{Volume: vol.ID(), Error: err.Error()}))
			return err
		}

		return s.SendMsg(stream, s.rpcReq(&volumesAPI.BlockCommandResponse{Volume: vol.ID()}))
	default:
		return fmt.Errorf("unknown cmd type")
	}
}
