package sbs

import (
	"bytes"
	"context"
	"encoding/binary"
	"net"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	volumesAPI "github.com/tcfw/vpc/pkg/api/v1/volumes"
	"github.com/tcfw/vpc/pkg/utils"
)

func TestNBDSendCmdReply(t *testing.T) {
	sConn, cConn := net.Pipe()

	s := &session{
		conn: sConn,
		log:  utils.DefaultLogger(),
	}

	go func() {
		err := s.sendCmdReply(&nbdWork{
			nbdReq: nbdRequest{
				NbdCommandType: NBDCMDWrite,
				NbdHandle:      1,
			},
		}, nil, 0)
		if err != nil {
			t.Fatal(err)
			return
		}
	}()

	buf := make([]byte, 8096)

	n, err := cConn.Read(buf)
	if err != nil {
		t.Fatal(err)
	}

	reply := &nbdReply{}
	binary.Read(bytes.NewReader(buf[:n]), binary.BigEndian, reply)

	assert.Equal(t, uint32(NBDMAGICReply), reply.NbdReplyMagic)
	assert.Equal(t, uint64(1), reply.NbdHandle)
	assert.Equal(t, uint32(0), reply.NbdError)
}

func TestNBDReadCmd(t *testing.T) {
	sConn, cConn := net.Pipe()

	volDesc := &volumesAPI.Volume{
		Id:   "vol-test",
		Size: 10, //GB
	}

	dir := os.TempDir()
	defer os.RemoveAll(dir + "/" + volDesc.Id)

	bs := NewBlockStore("", volDesc, dir, utils.DefaultLogger())

	data := make([]byte, BlockSize)

	data[0] = 0x1

	err := bs.write(&StoreCommand{
		Op:     OpWrite,
		Offset: 0,
		Length: BlockSize,
		Data:   data,
	})
	if err != nil {
		t.Fatal(err)
	}

	vol := &Volume{
		Blocks: bs,
	}

	s := &session{
		conn:     sConn,
		log:      utils.DefaultLogger(),
		volume:   vol,
		workCh:   make(chan *nbdWork, 2),
		bufAlloc: newBufAlloc(maxBlockAlloc),
	}

	go s.dispatch(context.Background(), 1)

	//Attempt multiple reads
	s.inFlight.Add(1)
	s.workCh <- &nbdWork{
		nbdReq: nbdRequest{
			NbdHandle:      1234,
			NbdCommandType: NBDCMDRead,
			NbdOffset:      0,
			NbdLength:      BlockSize,
		},
	}

	//Read reply
	buf := make([]byte, BlockSize)
	n, err := cConn.Read(buf)
	if err != nil {
		t.Fatal(err)
	}

	reply := &nbdReply{}
	binary.Read(bytes.NewReader(buf[:n]), binary.BigEndian, reply)
	assert.Equal(t, uint32(NBDMAGICReply), reply.NbdReplyMagic)

	//Read payload
	n, err = cConn.Read(buf)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, data, buf[:n])
}

func BenchmarkNBDReadCmd(b *testing.B) {
	sConn, cConn := net.Pipe()

	volDesc := &volumesAPI.Volume{
		Id:   "vol-test",
		Size: 10, //GB
	}

	dir := os.TempDir()
	defer os.RemoveAll(dir + "/" + volDesc.Id)

	bs := NewBlockStore("", volDesc, dir, utils.DefaultLogger())

	data := make([]byte, BlockSize)

	data[0] = 0x1

	err := bs.write(&StoreCommand{
		Op:     OpWrite,
		Offset: 0,
		Length: BlockSize,
		Data:   data,
	})
	if err != nil {
		b.Fatal(err)
	}

	vol := &Volume{
		Blocks: bs,
	}

	go func() {
		buf := make([]byte, BlockSize)
		//recv
		for {
			cConn.Read(buf)
			cConn.Read(buf)
		}
	}()

	b.Run("no-cached", func(b *testing.B) {
		b.SetBytes(BlockSize)
		s := &session{
			conn:     sConn,
			log:      utils.DefaultLogger(),
			volume:   vol,
			workCh:   make(chan *nbdWork, 200),
			bufAlloc: newBufAlloc(maxBlockAlloc),
		}

		b.RunParallel(func(pb *testing.PB) {
			go s.dispatch(context.Background(), 0)
			for pb.Next() {
				s.inFlight.Add(1)
				s.workCh <- &nbdWork{
					nbdReq: nbdRequest{
						NbdHandle:      1234,
						NbdCommandType: NBDCMDRead,
						NbdOffset:      0,
						NbdLength:      BlockSize,
					},
				}
			}
		})
	})

	b.Run("cached", func(b *testing.B) {
		b.SetBytes(BlockSize)
		s := &session{
			conn:       sConn,
			log:        utils.DefaultLogger(),
			volume:     vol,
			workCh:     make(chan *nbdWork, 200),
			bufAlloc:   newBufAlloc(maxBlockAlloc),
			blockCache: newBlockCache(),
		}

		b.RunParallel(func(pb *testing.PB) {
			go s.dispatch(context.Background(), 0)
			for pb.Next() {
				s.inFlight.Add(1)
				s.workCh <- &nbdWork{
					nbdReq: nbdRequest{
						NbdHandle:      1234,
						NbdCommandType: NBDCMDRead,
						NbdOffset:      0,
						NbdLength:      BlockSize,
					},
				}
			}
		})
	})
}
