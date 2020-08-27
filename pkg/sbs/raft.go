package sbs

import (
	"bytes"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/hashicorp/raft"
	"github.com/lucas-clemente/quic-go"
	"github.com/sirupsen/logrus"
	"github.com/tcfw/vpc/pkg/utils"
	"github.com/vmihailenco/msgpack"

	volumesAPI "github.com/tcfw/vpc/pkg/api/v1/volumes"
)

const (
	rpcAppendEntries int32 = iota
	rpcRequestVote
	rpcInstallSnapshot
	rpcTimeoutNow

	rpcResp              int32 = 99
	pipelineMaxInFlight  int   = 256
	raftTimingMultiplier       = 3
)

//NewRaft creats a new raft instance
func NewRaft(s *Server, bs *BlockStore, ls raft.LogStore, trans raft.Transport) (*raft.Raft, error) {
	config := raft.DefaultConfig()
	config.LocalID = raft.ServerID(s.PeerID())
	config.LogOutput = utils.DefaultLogger().WriterLevel(logrus.DebugLevel)

	config.HeartbeatTimeout = raftTimingMultiplier * 1000 * time.Millisecond
	config.ElectionTimeout = raftTimingMultiplier * 1000 * time.Millisecond
	config.CommitTimeout = raftTimingMultiplier * 50 * time.Millisecond
	config.LeaderLeaseTimeout = raftTimingMultiplier * 500 * time.Millisecond

	metastore := raft.NewInmemStore()

	dir := fmt.Sprintf("%s", bs.BaseDir)
	snaps, _ := raft.NewFileSnapshotStore(dir, 2, utils.DefaultLogger().Writer())

	r, err := raft.NewRaft(
		config,
		bs,
		ls,
		metastore,
		snaps,
		trans)
	if err != nil {
		return nil, err
	}

	return r, nil
}

//NewRaftTransport creates a new transport compatible with raft
func NewRaftTransport(s *Server, volumeID string) *RaftTransport {
	return &RaftTransport{
		server:    s,
		volumeID:  volumeID,
		consumeCh: make(chan raft.RPC),
		log:       s.log,
	}
}

//RaftTransport allows for raft rpcs via the main p2p comms
//loosely based around raft.NetworkTransport
type RaftTransport struct {
	server    *Server
	volumeID  string
	peersLock sync.Mutex
	log       *logrus.Logger

	consumeCh chan raft.RPC

	heartbeatFn     func(raft.RPC)
	heartbeatFnLock sync.Mutex
}

//HandleCommand handles a remote Raft RPC command
func (r *RaftTransport) HandleCommand(s quic.Stream, p *volumesAPI.RaftRPC) error {
	// r.log.WithField("type", p.Type).Trace("got raft rpc")

	respCh := make(chan raft.RPCResponse, 1)
	rpc := raft.RPC{
		RespChan: respCh,
	}

	isHeartbeat := false
	switch p.Type {
	case rpcAppendEntries:
		var req raft.AppendEntriesRequest
		if err := msgpack.Unmarshal(p.Command, &req); err != nil {
			return err
		}
		rpc.Command = &req

		// Check if this is a heartbeat
		if req.Term != 0 && req.Leader != nil &&
			req.PrevLogEntry == 0 && req.PrevLogTerm == 0 &&
			len(req.Entries) == 0 && req.LeaderCommitIndex == 0 {
			isHeartbeat = true
		}

	case rpcRequestVote:
		var req raft.RequestVoteRequest
		if err := msgpack.Unmarshal(p.Command, &req); err != nil {
			return err
		}
		rpc.Command = &req

	case rpcInstallSnapshot:
		var req raft.InstallSnapshotRequest
		if err := msgpack.Unmarshal(p.Command, &req); err != nil {
			return err
		}
		rpc.Command = &req
		rpc.Reader = io.LimitReader(bytes.NewReader(p.Command), req.Size)

	case rpcTimeoutNow:
		var req raft.TimeoutNowRequest
		if err := msgpack.Unmarshal(p.Command, &req); err != nil {
			return err
		}
		rpc.Command = &req

	default:
		return fmt.Errorf("unknown rpc type %d", p.Type)
	}

	// Check for heartbeat fast-path
	if isHeartbeat {
		r.heartbeatFnLock.Lock()
		fn := r.heartbeatFn
		r.heartbeatFnLock.Unlock()
		if fn != nil {
			fn(rpc)
			goto RESP
		}
	}

	// Dispatch the RPC
	select {
	case r.consumeCh <- rpc:
	}

RESP:
	select {
	case resp := <-respCh:
		// Send the error first
		respErr := ""
		if resp.Error != nil {
			respErr = resp.Error.Error()
		}

		pResp := &volumesAPI.RaftRPC{
			Volume: r.volumeID,
			Type:   rpcResp,
			Error:  respErr,
		}

		cmdBytes, err := msgpack.Marshal(resp.Response)
		if err != nil {
			return err
		}
		pResp.Command = cmdBytes

		if err := r.server.SendMsg(s, pResp); err != nil {
			return err
		}
	}

	return nil
}

// SetHeartbeatHandler is used to setup a heartbeat handler
// as a fast-pass. This is to avoid head-of-line blocking from
// disk IO.
func (r *RaftTransport) SetHeartbeatHandler(cb func(rpc raft.RPC)) {
	r.heartbeatFnLock.Lock()
	defer r.heartbeatFnLock.Unlock()

	r.heartbeatFn = cb
}

// Consumer returns a channel that can be used to
// consume and respond to RPC requests.
func (r *RaftTransport) Consumer() <-chan raft.RPC {
	return r.consumeCh
}

// LocalAddr is used to return our local address to distinguish from our peers.
func (r *RaftTransport) LocalAddr() raft.ServerAddress {
	return raft.ServerAddress(r.server.PeerID())
}

// AppendEntries sends the appropriate RPC to the target node.
func (r *RaftTransport) AppendEntries(id raft.ServerID, target raft.ServerAddress, args *raft.AppendEntriesRequest, resp *raft.AppendEntriesResponse) error {
	return r.genericRPC(id, target, rpcAppendEntries, args, resp)
}

// RequestVote sends the appropriate RPC to the target node.
func (r *RaftTransport) RequestVote(id raft.ServerID, target raft.ServerAddress, args *raft.RequestVoteRequest, resp *raft.RequestVoteResponse) error {
	return r.genericRPC(id, target, rpcRequestVote, args, resp)
}

// TimeoutNow is used to start a leadership transfer to the target node.
func (r *RaftTransport) TimeoutNow(id raft.ServerID, target raft.ServerAddress, args *raft.TimeoutNowRequest, resp *raft.TimeoutNowResponse) error {
	return r.genericRPC(id, target, rpcTimeoutNow, args, resp)
}

func (r *RaftTransport) genericRPC(id raft.ServerID, target raft.ServerAddress, rpcType int32, args interface{}, resp interface{}) error {
	peer := r.server.GetPeer(string(id))
	if peer == nil {
		return fmt.Errorf("unknown peer")
	}

	// r.log.WithFields(logrus.Fields{
	// 	"type": rpcType,
	// 	"peer": id,
	// }).Trace("sending raft rpc")

	req := &volumesAPI.RaftRPC{
		Volume: r.volumeID,
		Type:   rpcType,
	}

	cmdBytes, err := msgpack.Marshal(args)
	if err != nil {
		return fmt.Errorf("failed to marshal raft rpc: %s", err)
	}
	req.Command = cmdBytes

	m, err := r.server.SendReqViaChannel(peer, "raft", req)
	if err != nil {
		return fmt.Errorf("raft - failed resp from peer: %s", err)
	}

	rpcResp := m.(*volumesAPI.RaftRPC)
	if rpcResp.Error != "" {
		r.log.WithField("volume", r.volumeID).Errorf("Raft: %s", err)
	}

	return msgpack.Unmarshal(rpcResp.Command, resp)
}

// AppendEntriesPipeline returns an interface that can be used to pipeline
// AppendEntries requests.
func (r *RaftTransport) AppendEntriesPipeline(id raft.ServerID, target raft.ServerAddress) (raft.AppendPipeline, error) {
	peer := r.server.GetPeer(string(id))
	if peer == nil {
		return nil, fmt.Errorf("unknown peer")
	}

	return newRaftPipeline(r.server, peer, r.volumeID)
}

// InstallSnapshot is used to push a snapshot down to a follower. The data is read from
// the ReadCloser and streamed to the client.
func (r *RaftTransport) InstallSnapshot(id raft.ServerID, target raft.ServerAddress, args *raft.InstallSnapshotRequest, resp *raft.InstallSnapshotResponse, data io.Reader) error {
	return nil
}

// EncodePeer is used to serialize a peer's address.
func (r *RaftTransport) EncodePeer(id raft.ServerID, addr raft.ServerAddress) []byte {
	return []byte(addr)
}

// DecodePeer is used to deserialize a peer's address.
func (r *RaftTransport) DecodePeer(p []byte) raft.ServerAddress {
	return raft.ServerAddress(p)
}

//If only this was exposed
type appendFuture struct {
	start time.Time
	args  *raft.AppendEntriesRequest
	resp  *raft.AppendEntriesResponse

	err        error
	errCh      chan error
	responded  bool
	ShutdownCh chan struct{}
}

func (d *appendFuture) init() {
	d.errCh = make(chan error, 1)
}

func (d *appendFuture) Error() error {
	if d.err != nil {
		// Note that when we've received a nil error, this
		// won't trigger, but the channel is closed after
		// send so we'll still return nil below.
		return d.err
	}
	if d.errCh == nil {
		panic("waiting for response on nil channel")
	}
	select {
	case d.err = <-d.errCh:
	case <-d.ShutdownCh:
		d.err = fmt.Errorf("shutdown")
	}
	return d.err
}

func (d *appendFuture) respond(err error) {
	if d.errCh == nil {
		return
	}
	if d.responded {
		return
	}
	d.errCh <- err
	close(d.errCh)
	d.responded = true
}

func (d *appendFuture) Start() time.Time {
	return d.start
}

func (d *appendFuture) Request() *raft.AppendEntriesRequest {
	return d.args
}

func (d *appendFuture) Response() *raft.AppendEntriesResponse {
	return d.resp
}

type raftPipeline struct {
	server   *Server
	conn     quic.Stream
	volumeID string

	doneCh       chan raft.AppendFuture
	inprogressCh chan *appendFuture

	shutdown     bool
	shutdownCh   chan struct{}
	shutdownLock sync.Mutex
}

func newRaftPipeline(s *Server, p *Peer, v string) (*raftPipeline, error) {
	stream, err := p.Conn().OpenStream()
	if err != nil {
		return nil, fmt.Errorf("failed to open peer stream: %s", err)
	}

	pipeline := &raftPipeline{
		conn:         stream,
		volumeID:     v,
		shutdownCh:   make(chan struct{}),
		doneCh:       make(chan raft.AppendFuture, pipelineMaxInFlight),
		inprogressCh: make(chan *appendFuture, pipelineMaxInFlight),
	}

	go pipeline.work()

	return pipeline, nil
}

func (rp *raftPipeline) work() {
	buf := make([]byte, peerMsgBufSize)

	for {
		select {
		case <-rp.shutdownCh:
			return
		case future := <-rp.inprogressCh:
			rpcResp := &volumesAPI.PeerRPC{}
			var raftResp *volumesAPI.RaftRPC
			n, err := rp.conn.Read(buf)
			if err != nil {
				future.respond(err)
				rp.Close()
				goto DONE
			}

			if err := proto.Unmarshal(buf[:n], rpcResp); err != nil {
				future.respond(err)
				goto DONE
			}

			raftResp = rpcResp.GetRaftRPC()
			if raftResp.Volume != rp.volumeID {
				future.respond(fmt.Errorf("received mismatch volume rpc response"))
				goto DONE
			}

			future.respond(fmt.Errorf(raftResp.Error))
		DONE:
			select {
			case rp.doneCh <- future:
			case <-rp.shutdownCh:
				return
			}

		}
	}
}

// AppendEntries is used to pipeline a new append entries request.
func (rp *raftPipeline) AppendEntries(args *raft.AppendEntriesRequest, resp *raft.AppendEntriesResponse) (raft.AppendFuture, error) {
	future := &appendFuture{
		start: time.Now(),
		args:  args,
		resp:  resp,
	}
	future.init()

	cmd, _ := msgpack.Marshal(future.args)

	err := rp.server.SendMsg(rp.conn, &volumesAPI.PeerRPC{RPC: &volumesAPI.PeerRPC_RaftRPC{RaftRPC: &volumesAPI.RaftRPC{
		Volume:  rp.volumeID,
		Type:    rpcAppendEntries,
		Command: cmd,
	}}})
	if err != nil {
		return nil, err
	}

	select {
	case rp.inprogressCh <- future:
		return future, nil
	case <-rp.shutdownCh:
		return nil, fmt.Errorf("pipeline is shutdown")
	}
}

// Consumer returns a channel that can be used to consume complete futures.
func (rp *raftPipeline) Consumer() <-chan raft.AppendFuture {
	return rp.doneCh
}

// Closed is used to shutdown the pipeline connection.
func (rp *raftPipeline) Close() error {
	rp.shutdownLock.Lock()
	defer rp.shutdownLock.Unlock()
	if rp.shutdown {
		return nil
	}

	rp.conn.Close()

	rp.shutdown = true
	close(rp.shutdownCh)
	return nil
}
