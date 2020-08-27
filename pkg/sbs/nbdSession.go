package sbs

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/sirupsen/logrus"
	volumesAPI "github.com/tcfw/vpc/pkg/api/v1/volumes"
)

const (
	maxInFlight          = 1024
	defaultWorkers       = 24
	maxBlockAlloc  int32 = 16000

	//Currently supported funcs
	txSupportedFuncs = NBDFLAGSendFlush | NBDFLAGSendFUA | NBDFLAGSendWriteZeroes | NBDFLAGMultiCon
)

type nbdWork struct {
	nbdReq   nbdRequest
	payload  *bytes.Buffer
	cacheHit bool
}

func newSession(n *NBDServer, s *Server, conn net.Conn, l *logrus.Logger) *session {
	return &session{
		server:     s,
		nbd:        n,
		conn:       conn,
		log:        l,
		shutdownCh: make(chan struct{}),
		bufAlloc:   newBufAlloc(maxBlockAlloc),
		blockCache: newBlockCache(),
	}
}

type session struct {
	server *Server
	nbd    *NBDServer
	conn   net.Conn
	log    *logrus.Logger

	volume        *Volume
	peers         []*Peer
	currentLeader *Peer

	txMu       sync.Mutex
	workCh     chan *nbdWork
	shutdownCh chan struct{}
	shutdown   bool

	inFlight   sync.WaitGroup
	bufAlloc   *bufAlloc
	blockCache *blockCache

	useStructuredReplies bool
}

func (s *session) handle(ctx context.Context) {
	s.workCh = make(chan *nbdWork, 2048)

	defer func() {
		s.kill()
	}()

	if err := s.negotiate(ctx); err != nil {
		s.log.WithError(err).Error("negotiation failed")
		return
	}

	//start Working
	go s.rx(ctx)
	for i := 0; i < defaultWorkers; i++ {
		go s.dispatch(ctx, i)
	}

	select {
	case <-s.shutdownCh:
		s.log.Infof("Worker forced close")
	case <-ctx.Done():
		s.log.Infof("Parent forced close")
	}
}

func (s *session) rx(ctx context.Context) {
	defer func() {
		s.kill()
	}()
	rxPayloadBuf := make([]byte, BlockSize*10) //some clients dont ahere to block sizes...
	for {
		req := &nbdWork{}
		if err := binary.Read(s.conn, binary.BigEndian, &req.nbdReq); err != nil {
			if err == io.EOF {
				return
			}
			//TODO(tcfw) log error
			s.log.WithError(err).Error("RX read close")

			return
		}

		if req.nbdReq.NbdRequestMagic != NBDMAGICRequest {
			s.log.Warnf("Client %s had bad magic number in request: %d", s.volume.id, req.nbdReq.NbdRequestMagic)
			return
		}

		shouldHalt := false

		s.log.Tracef("NBD REQ %+v", req.nbdReq)

		switch req.nbdReq.NbdCommandType {
		case NBDCMDDisc:
			shouldHalt = true
			fallthrough
		case NBDCMDFlush, NBDCMDRead, NBDCMDTrim, NBDCMDWriteZeroes:
			//TODO(tcfw) validate bounds
		case NBDCMDWrite:
			n, err := io.ReadFull(s.conn, rxPayloadBuf[:req.nbdReq.NbdLength])
			if err != nil {
				s.log.WithError(err).Error("failed to read payload")
				return
			}
			req.payload = s.bufAlloc.get()
			if _, err := req.payload.Write(rxPayloadBuf[:n]); err != nil {
				s.log.WithError(err).Error("failed to copy payload to buf")
				return
			}
		default:
			s.log.WithField("vol", s.volume.id).Error("received unknown command")
			return
		}

		s.inFlight.Add(1)
		select {
		case s.workCh <- req:
		case <-ctx.Done():
			return
		}

		if shouldHalt {
			//Wait for all requests to be fulfilled then start shutdown
			s.inFlight.Wait()
			return
		}

	}
}

func (s *session) dispatch(ctx context.Context, id int) {
	var replyErr error
	var replyErrNo uint32

	s.log.WithField("wid", id).Tracef("listening for NBD commands")

	for req := range s.workCh {
		replyErrNo = 0
		useLocal := len(s.peers) == 0

		// s.log.WithField("vol", s.volume.ID()).Tracef("NBD REQ %+v", req.nbdReq)

		//do something
		switch req.nbdReq.NbdCommandType {
		case NBDCMDRead:
			//TODO(tcfw) read from block cache
			offset := req.nbdReq.NbdOffset
			length := req.nbdReq.NbdLength
			var peer *Peer
			if len(s.peers) > 0 {
				peer = s.peers[rand.Intn(len(s.peers))]
			}

			if s.blockCache != nil {
				cache := s.blockCache.get(offset)
				if cache != nil && uint32(len(cache.data)) == length {
					s.log.Tracef("cache hit")
					req.payload = bytes.NewBuffer(cache.data)
					req.cacheHit = true
					goto REPLY
				}
				// s.blockCache.invalidate(offset)
				s.log.Tracef("cache miss")
			}

			if req.payload == nil {
				req.payload = s.bufAlloc.get()
			}

			req.payload.Grow(int(length))

			buf := make([]byte, BlockSize)
			for length > 0 {
				bLen := uint32(BlockSize)
				if bLen > length {
					bLen = length
					buf = buf[:bLen]
				}
				if !useLocal {
					cmd := &volumesAPI.BlockCommandRequest{
						Volume: s.volume.ID(),
						Cmd:    volumesAPI.BlockCommandRequest_READ,
						Offset: int64(offset),
						Length: int32(bLen),
					}
					resp, err := s.server.sendReq(peer, cmd)
					if err != nil {
						replyErrNo = NBDEIO
						replyErr = err
						s.log.WithError(err).Error("vol cluster rpc failed")
						goto REPLY
					}

					bcr := resp.(*volumesAPI.BlockCommandResponse)
					if bcr.Error != "" {
						err = errors.New(bcr.Error)
						replyErrNo = NBDEIO
						replyErr = err
						s.log.WithError(err).Errorf("vol cluster err")
						goto REPLY
					}

					// s.log.Tracef("got %d back from vol cluster", len(bcr.Data))

					_, err = req.payload.Write(bcr.Data)
					if err != nil {
						s.log.WithError(err).Error("failed to write back to buffer")
						return
					}
					if uint32(len(bcr.Data)) != bLen {
						//assume end-of-block read
						bLen = uint32(len(bcr.Data))
					}
				} else {
					_, err := s.volume.Blocks.ReadAt(buf, offset)
					if err != nil {
						replyErr = err
						s.log.WithError(err).Errorf("local vol err")
						goto REPLY
					}
					req.payload.Write(buf)
				}
				offset += uint64(bLen)
				length -= bLen

			}

			if req.payload.Len() != int(req.nbdReq.NbdLength) {
				s.log.Warn("invalid payload len")
			}

			//update cache for next read
			if s.blockCache != nil {
				s.blockCache.set(req.nbdReq.NbdOffset, req.payload.Bytes())
			}

			// s.log.Trace("handled read")

		case NBDCMDWrite, NBDCMDWriteZeroes:
			//update cache
			if s.blockCache != nil {
				s.blockCache.set(req.nbdReq.NbdOffset, req.payload.Bytes())
				s.log.Tracef("cache set")
			}
			// s.log.WithFields(logrus.Fields{"offset": req.nbdReq.NbdOffset, "len": req.payload.Len()}).Trace("got write request")
			if useLocal {
				cmd := &StoreCommand{
					Op:     OpZeros,
					Offset: req.nbdReq.NbdOffset,
					Length: req.nbdReq.NbdLength,
				}
				if req.nbdReq.NbdCommandType == NBDCMDWrite {
					cmd.Op = OpWrite
					cmd.Data = req.payload.Bytes()
				}
				cmdBytes, _ := cmd.Encode()
				var err error
				future := s.volume.Raft.Apply(cmdBytes, 10*time.Second)
				if req.nbdReq.NbdCommandFlags&NBDCMDFlagFUA != 0 {
					//wait for apply
					err = future.Error()
				}
				if err != nil {
					s.log.WithError(err).Error("failed to apply log entry")
					replyErr = err
					goto REPLY
				}

			} else {
				if s.currentLeader == nil {
					s.currentLeader = s.peers[rand.Intn(len(s.peers))]
				}
				//Allow 2 retries, once if not master, and 2nd for actaul failure
				for tries := 2; tries > 0; tries-- {
					rpcCmd := &volumesAPI.PeerRPC{
						RPC: &volumesAPI.PeerRPC_BlockCommand{
							BlockCommand: &volumesAPI.BlockCommandRequest{
								Volume: s.volume.ID(),
								Cmd:    volumesAPI.BlockCommandRequest_READ,
								Offset: int64(req.nbdReq.NbdOffset),
								Length: int32(req.payload.Len()),
								Data:   req.payload.Bytes(),
							},
						},
					}
					m, err := s.server.sendReq(s.currentLeader, rpcCmd)
					if err != nil {
						replyErr = err
						s.log.WithError(err).Errorf("failed leader rpc")
						goto REPLY
					}
					resp := m.(*volumesAPI.BlockCommandResponse)
					if resp.RetryAt != "" {
						s.currentLeader = s.server.GetPeer(resp.RetryAt)
						s.log.WithField("vol", s.volume.id).Tracef("updating leader")
						continue
					}
					if resp.Error != "" {
						replyErr = errors.New(resp.Error)
						s.log.WithError(err).Errorf("rpc error")
						goto REPLY
					}
					break
				}
			}

		case NBDCMDFlush:
			if useLocal {
				cmd := &StoreCommand{
					Op: OpFlush,
				}
				cmdBytes, _ := cmd.Encode()
				future := s.volume.Raft.Apply(cmdBytes, 5*time.Second)
				if err := future.Error(); err != nil {
					replyErr = err
					goto REPLY
				}
			}
		case NBDCMDTrim:
			//TODO(tcfw) validate bounds
		}
	REPLY:
		//send reply
		if err := s.sendCmdReply(req, replyErr, replyErrNo); err != nil {
			return
		}

		s.inFlight.Done()
		if req.payload != nil && !req.cacheHit {
			s.bufAlloc.put(req.payload)
		}
	}

	s.log.WithField("wid", id).Debug("listening closed")

}

func (s *session) sendCmdReply(req *nbdWork, replyErr error, replyErrNo uint32) error {
	if replyErr != nil && replyErrNo == 0 {
		replyErrNo = NBDEINVAL
	}

	if replyErrNo != 0 {
		s.log.WithError(replyErr).Errorf("cmd reply err [%d]: %s", replyErrNo, replyErr)
	}

	if s.useStructuredReplies {
		return s.sendCmdStructuredReply(req, replyErr, replyErrNo)
	}

	// s.log.Tracef("sending reply for %d", req.nbdReq.NbdHandle)

	reply := &nbdReply{
		NbdReplyMagic: NBDMAGICReply,
		NbdHandle:     req.nbdReq.NbdHandle,
		NbdError:      replyErrNo,
	}

	s.txMu.Lock()
	defer s.txMu.Unlock()

	if err := binary.Write(s.conn, binary.BigEndian, reply); err != nil {
		s.log.WithField("vol", s.volume.ID()).WithError(err).Error("failed to write reply")
		return err
	}

	if req.nbdReq.NbdCommandType == NBDCMDRead && req.payload != nil && replyErrNo == 0 && replyErr == nil {
		//send payload
		var length int64 = int64(req.payload.Len())
		var nn int64
		for length > 0 {
			n, err := req.payload.WriteTo(s.conn)
			if err != nil {
				s.log.WithField("vol", s.volume.ID()).WithError(err).Error("failed to write payload")
				return err
			}
			nn += n
			length -= n
		}
		// s.log.Tracef("wrote %d bytes for req %d (bsize: %d)", nn, req.nbdReq.NbdHandle, req.payload.Len())
	}
	// s.log.Tracef("sent reply for %d", req.nbdReq.NbdHandle)

	return nil
}

func (s *session) sendCmdStructuredReply(req *nbdWork, err error, replyErr uint32) error {
	// s.log.Tracef("sending structured reply for %d", req.nbdReq.NbdHandle)

	reply := &nbdStructuredReply{
		NbdReplyMagic: NBDMAGICStructuredReply,
		NbdHandle:     req.nbdReq.NbdHandle,
		NbdFlags:      NBDReplyFlagDone,
	}
	var replyErrPayload nbdStructuredError
	var replyErrMsg string

	if err != nil {
		reply.NbdType = NBDREPLYTYPEError
		if replyErr != 0 {
			replyErrPayload.NbdError = replyErr
		} else {
			replyErrPayload.NbdError = NBDEINVAL
		}
		s.log.WithError(err).Errorf("cmd failed :%d", replyErr)
	}

	if replyErrPayload.NbdError != 0 {
		replyErrPayload.NbdLength = uint16(len(replyErrMsg))

		errPayload := bytes.NewBuffer(nil)
		binary.Write(errPayload, binary.BigEndian, replyErrPayload)
		binary.Write(errPayload, binary.BigEndian, replyErrMsg)
		req.payload = errPayload
	}

	if req.payload != nil {
		reply.NbdLength = uint32(req.payload.Len())
	}

	s.txMu.Lock()
	defer s.txMu.Unlock()

	if err := binary.Write(s.conn, binary.BigEndian, reply); err != nil {
		s.txMu.Unlock()
		s.log.WithField("vol", s.volume.ID()).WithError(err).Error("failed to write reply")
		return err
	}

	if req.nbdReq.NbdCommandType == NBDCMDRead && req.payload != nil && reply.NbdType != NBDREPLYTYPEError {
		bLen := req.nbdReq.NbdLength
		if req.nbdReq.NbdOffset+uint64(bLen) > uint64(s.volume.desc.Size*VolDescSizeMultiplier) {
			bLen = uint32(uint64(s.volume.desc.Size*VolDescSizeMultiplier) - req.nbdReq.NbdOffset)
		}

		//send payload
		var offset uint32
		nn := 0
		for bLen > 0 {
			n, err := s.conn.Write(req.payload.Bytes()[:offset+bLen])
			if err != nil {
				s.log.WithField("vol", s.volume.ID()).WithError(err).Error("failed to write payload")
				return err
			}
			bLen -= uint32(n)
			offset += uint32(n)
			nn += n
		}
		// s.log.Tracef("write %d bytes for req %d", nn, req.nbdReq.NbdHandle)
	}
	s.log.Debugf("sent reply for %d", req.nbdReq.NbdHandle)
	return nil
}

type bufAlloc struct {
	pool  chan *bytes.Buffer
	inUse int32
}

func newBufAlloc(max int32) *bufAlloc {
	return &bufAlloc{
		pool: make(chan *bytes.Buffer, max),
	}
}

func (ba *bufAlloc) get() *bytes.Buffer {
	defer atomic.AddInt32(&ba.inUse, 1)

	select {
	case b := <-ba.pool:
		return b
	default:
		return bytes.NewBuffer(make([]byte, 0, BlockSize))
	}
}

func (ba *bufAlloc) put(b *bytes.Buffer) {
	b.Reset()

	if b.Cap() > BlockSize {
		b = bytes.NewBuffer(make([]byte, 0, BlockSize))
	}

	atomic.AddInt32(&ba.inUse, -1)

	select {
	case ba.pool <- b:
	default:
		//gc
	}
}

func (s *session) kill() {
	if !s.shutdown {
		close(s.shutdownCh)
		close(s.workCh)
		s.shutdown = true
	}
}

func (s *session) negotiate(ctx context.Context) error {
	s.conn.SetDeadline(time.Now().Add(5 * time.Second))

	//initial phase of negotiation
	hello := nbdNewStyleHeader{
		NbdMagic:       NBDMAGIC,
		NbdOptsMagic:   NBDMAGICOpts,
		NbdGlobalFlags: NBDFLAGFixedNewstyle,
	}
	if err := binary.Write(s.conn, binary.BigEndian, hello); err != nil {
		return errors.New("Cannot write magic header")
	}
	s.log.Trace("NBD sent hello")

	//hi flags
	var clientFlags nbdClientFlags
	if err := binary.Read(s.conn, binary.BigEndian, &clientFlags); err != nil {
		return errors.New("Cannot read client flags")
	}
	s.log.Trace("NBD received client flags")

	//option haggling
	done := false
	for !done {
		var opt nbdClientOpt
		if err := binary.Read(s.conn, binary.BigEndian, &opt); err != nil {
			return errors.New("Cannot read option (perhaps client dropped the connection)")
		}
		s.log.WithField("opt", opt).Trace("NBD received client opt")

		//validate option
		if opt.NbdOptMagic != NBDMAGICOpts {
			return errors.New("Bad option magic")
		}

		//bound check
		if opt.NbdOptLen > 65536 {
			return errors.New("Option is too long")
		}

		switch opt.NbdOptID {
		case NBDOPTExportName, NBDOPTInfo, NBDOPTGo:
			var name []byte

			// clientSupportsBlockSizeConstraints := false

			if opt.NbdOptID == NBDOPTExportName {
				name = make([]byte, opt.NbdOptLen)
				n, err := io.ReadFull(s.conn, name)
				if err != nil {
					return err
				}

				if uint32(n) != opt.NbdOptLen {
					return errors.New("Incomplete name")
				}
			} else {
				var nameLength uint32
				if err := binary.Read(s.conn, binary.BigEndian, &nameLength); err != nil {
					return errors.New("Bad export name length")
				}

				if nameLength > 4096 {
					return errors.New("Name is too long")
				}

				name = make([]byte, nameLength)
				n, err := io.ReadFull(s.conn, name)
				if err != nil {
					return err
				}

				if uint32(n) != nameLength {
					return errors.New("Incomplete name")
				}

				s.log.Tracef("NBD name reqd: %s", string(name))

				var numInfoElements uint16
				if err := binary.Read(s.conn, binary.BigEndian, &numInfoElements); err != nil {
					return errors.New("Bad number of info elements")
				}

				for i := uint16(0); i < numInfoElements; i++ {
					var infoElement uint16
					if err := binary.Read(s.conn, binary.BigEndian, &infoElement); err != nil {
						return errors.New("Bad number of info elements")
					}
					switch infoElement {
					case NBDINFOBlockSize:
						// clientSupportsBlockSizeConstraints = true
					}
				}
				l := 2 + 2*uint32(numInfoElements) + 4 + uint32(nameLength)
				if opt.NbdOptLen > l {
					if err := skip(s.conn, opt.NbdOptLen-l); err != nil {
						return err
					}
				} else if opt.NbdOptLen < l {
					return errors.New("Option length too short")
				}
			}

			//force TLS
			// if _, isTLS := s.conn.(*tls.Conn); !isTLS {
			// 	s.log.WithField("name", string(name)).Tracef("NBD requesting TLS upgrade")

			// 	or := nbdOptReply{
			// 		NbdOptReplyMagic:  NBDMAGICRep,
			// 		NbdOptID:          opt.NbdOptID,
			// 		NbdOptReplyType:   NBDREPErrTLSReqd,
			// 		NbdOptReplyLength: 0,
			// 	}
			// 	if err := binary.Write(s.conn, binary.BigEndian, or); err != nil {
			// 		return errors.New("Cannot send info error")
			// 	}
			// 	break
			// }

			//lookup volume
			vol, err := s.lookupVolume(string(name))
			if err != nil {
				or := nbdOptReply{
					NbdOptReplyMagic:  NBDMAGICRep,
					NbdOptID:          opt.NbdOptID,
					NbdOptReplyType:   NBDREPErrUnknown,
					NbdOptReplyLength: 0,
				}
				if err := binary.Write(s.conn, binary.BigEndian, or); err != nil {
					return errors.New("Cannot send info error")
				}
				s.log.WithField("name", string(name)).WithError(err).Tracef("NBD invalid vol. Killing connection")
				break
			}

			s.log.WithField("name", string(name)).Tracef("NBD attaching vol to session")

			if opt.NbdOptID == NBDOPTExportName {
				// this option has a unique reply format
				ed := nbdExportDetails{
					NbdExportSize:  uint64(vol.desc.Size * VolDescSizeMultiplier),
					NbdExportFlags: NBDFLAGHasFlags | txSupportedFuncs,
				}
				if err := binary.Write(s.conn, binary.BigEndian, ed); err != nil {
					return errors.New("Cannot write export details")
				}
			} else {
				s.log.WithField("name", string(name)).Tracef("Sending block info")
				// Send NBD_INFO_EXPORT
				or := nbdOptReply{
					NbdOptReplyMagic:  NBDMAGICRep,
					NbdOptID:          opt.NbdOptID,
					NbdOptReplyType:   NBDREPInfo,
					NbdOptReplyLength: 12,
				}
				if err := binary.Write(s.conn, binary.BigEndian, or); err != nil {
					return errors.New("Cannot write info export pt1")
				}

				ir := nbdInfoExport{
					NbdInfoType:          NBDINFOExport,
					NbdExportSize:        uint64(vol.desc.Size * VolDescSizeMultiplier),
					NbdTransmissionFlags: NBDFLAGHasFlags | txSupportedFuncs,
				}
				if err := binary.Write(s.conn, binary.BigEndian, ir); err != nil {
					return errors.New("Cannot write info export pt2")
				}

				// Send NBD_INFO_NAME
				or = nbdOptReply{
					NbdOptReplyMagic:  NBDMAGICRep,
					NbdOptID:          opt.NbdOptID,
					NbdOptReplyType:   NBDREPInfo,
					NbdOptReplyLength: uint32(2 + len(name)),
				}
				if err := binary.Write(s.conn, binary.BigEndian, or); err != nil {
					return errors.New("Cannot write info name pt1")
				}
				if err := binary.Write(s.conn, binary.BigEndian, uint16(NBDINFOName)); err != nil {
					return errors.New("Cannot write name id")
				}
				if err := binary.Write(s.conn, binary.BigEndian, name); err != nil {
					return errors.New("Cannot write name")
				}

				// Send NBD_INFO_BLOCK_SIZE
				or = nbdOptReply{
					NbdOptReplyMagic:  NBDMAGICRep,
					NbdOptID:          opt.NbdOptID,
					NbdOptReplyType:   NBDREPInfo,
					NbdOptReplyLength: 14,
				}
				if err := binary.Write(s.conn, binary.BigEndian, or); err != nil {
					return errors.New("Cannot write info block size pt1")
				}
				ir2 := nbdInfoBlockSize{
					NbdInfoType:           NBDINFOBlockSize,
					NbdMinimumBlockSize:   uint32(1),
					NbdPreferredBlockSize: uint32(BlockSize),
					NbdMaximumBlockSize:   uint32(BlockSize),
				}
				if err := binary.Write(s.conn, binary.BigEndian, ir2); err != nil {
					return errors.New("Cannot write info block size pt2")
				}

				replyType := NBDREPAck

				// if !clientSupportsBlockSizeConstraints {
				// 	replyType = NBDREPErrBlockSizeReqd
				// 	s.log.WithField("name", string(name)).Trace("NBD backing out of connection. Client did not specify block size")
				// }

				// Send ACK or error
				or = nbdOptReply{
					NbdOptReplyMagic:  NBDMAGICRep,
					NbdOptID:          opt.NbdOptID,
					NbdOptReplyType:   replyType,
					NbdOptReplyLength: 0,
				}
				if err := binary.Write(s.conn, binary.BigEndian, or); err != nil {
					return errors.New("Cannot info ack")
				}
				if opt.NbdOptID == NBDOPTInfo || or.NbdOptReplyType&NBDREPFlagError != 0 {
					break
				}

				s.log.WithField("name", string(name)).Tracef("NBD sent block info")
			}

			if clientFlags.NbdClientFlags&NBDFLAGCNoZeroes == 0 && opt.NbdOptID == NBDOPTExportName {
				// send 124 bytes of zeroes.
				zeroes := make([]byte, 124, 124)
				if err := binary.Write(s.conn, binary.BigEndian, zeroes); err != nil {
					return errors.New("Cannot write zeroes")
				}
			}

			s.volume = vol
			peerIDs := []string{}
			for p := range vol.PlacementPeers {
				peerIDs = append(peerIDs, p)
			}
			s.peers = s.server.GetPeers(peerIDs)

			done = true

		case NBDOPTStartTLS:
			or := nbdOptReply{
				NbdOptReplyMagic:  NBDMAGICRep,
				NbdOptID:          opt.NbdOptID,
				NbdOptReplyType:   NBDREPAck,
				NbdOptReplyLength: 0,
			}
			if err := binary.Write(s.conn, binary.BigEndian, or); err != nil {
				return errors.New("Cannot send TLS ack")
			}
			if err := s.startTLS(); err != nil {
				return fmt.Errorf("TLS failed to start: %s", err)
			}

		case NBDOPTAbort:
			or := nbdOptReply{
				NbdOptReplyMagic:  NBDMAGICRep,
				NbdOptID:          opt.NbdOptID,
				NbdOptReplyType:   NBDREPAck,
				NbdOptReplyLength: 0,
			}
			if err := binary.Write(s.conn, binary.BigEndian, or); err != nil {
				return errors.New("Cannot send abort ack")
			}
			return errors.New("Connection aborted by client")

		// case NBDOPTStructuredReply:
		// 	s.log.Info("NBD Using structured replies")

		// 	s.useStructuredReplies = true
		// 	or := nbdOptReply{
		// 		NbdOptReplyMagic:  NBDMAGICRep,
		// 		NbdOptID:          opt.NbdOptID,
		// 		NbdOptReplyType:   NBDREPAck,
		// 		NbdOptReplyLength: 0,
		// 	}
		// 	if err := binary.Write(s.conn, binary.BigEndian, or); err != nil {
		// 		return errors.New("Cannot send abort ack")
		// 	}

		default:
			if err := skip(s.conn, opt.NbdOptLen); err != nil {
				return err
			}

			s.log.Warnf("skipped client option: %d", opt.NbdOptID)

			// say it's unsuppported
			or := nbdOptReply{
				NbdOptReplyMagic:  NBDMAGICRep,
				NbdOptID:          opt.NbdOptID,
				NbdOptReplyType:   NBDREPErrUnsup,
				NbdOptReplyLength: 0,
			}
			if err := binary.Write(s.conn, binary.BigEndian, or); err != nil {
				return errors.New("Cannot reply to unsupported option")
			}
		}
	}

	s.log.WithField("vol", s.volume.id).Info("NBD volume attached")

	s.conn.SetDeadline(time.Time{})
	return nil
}

//startTLS replaces the current connection with a TLS connection
func (s *session) startTLS() error {
	//ignore if already TLS
	if _, isTLS := s.conn.(*tls.Conn); isTLS {
		return nil
	}

	tlsConfig, err := tlsConfig()
	if err != nil {
		return err
	}

	tls := tls.Server(s.conn, tlsConfig)
	s.conn = tls

	s.log.Trace("NBD upgrading to TLS")

	return tls.Handshake()
}

func (s *session) lookupVolume(name string) (*Volume, error) {
	vol, exists := s.nbd.attachments[name]
	if !exists {
		return nil, errors.New("not attached")
	}

	return vol, nil
}

func skip(r io.Reader, n uint32) error {
	for n > 0 {
		l := n
		if l > 1024 {
			l = 1024
		}
		b := make([]byte, l)
		if nr, err := io.ReadFull(r, b); err != nil {
			return err
		} else if nr != int(l) {
			return errors.New("skip returned short read")
		}
		n -= l
	}
	return nil
}
