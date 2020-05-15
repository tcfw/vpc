package vxlan

import (
	"net"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tcfw/vpc/pkg/l2/transport/tap/protocol"
)

func TestSendRecv(t *testing.T) {
	handler := NewHandler()

	if !assert.NoError(t, handler.Start()) {
		return
	}

	var recvPacket *protocol.Packet

	cliHandler := &Handler{
		connCount: runtime.NumCPU() * 2,
		port:      4790,
	}
	if !assert.NoError(t, cliHandler.Start()) {
		return
	}

	handler.SetHandler(func(ps []*protocol.Packet) {
		if len(ps) == 0 {
			return
		}

		recvPacket = ps[0]
	})

	cliHandler.SetHandler(func(ps []*protocol.Packet) {
		if len(ps) == 0 {
			return
		}

		recvPacket = ps[0]
	})

	frame := []byte{0x10, 0x20}
	var vnid uint32 = 5
	sendPacket := &protocol.Packet{VNID: vnid, Frame: frame}
	rdst := net.ParseIP("::1")

	_, err := cliHandler.Send([]*protocol.Packet{sendPacket}, rdst)
	if !assert.NoError(t, err) {
		return
	}

	//Wait for a packet to arrive
	for {
		if recvPacket != nil {
			break
		}
	}

	assert.Equal(t, frame, []byte(recvPacket.Frame))
	assert.Equal(t, vnid, recvPacket.VNID)

}

func BenchmarkSend(b *testing.B) {
	handler := NewHandler()
	handler.Start()

	received := 0

	handler.SetHandler(func(ps []*protocol.Packet) {
		received += len(ps)
	})

	frame := []byte{0x10, 0x20}
	var vnid uint32 = 5
	sendPacket := []*protocol.Packet{&protocol.Packet{VNID: vnid, Frame: frame}}
	rdst := net.ParseIP("::1")

	b.SetBytes(1)
	b.ResetTimer()

	b.Run("send", func(b *testing.B) {
		c := 0
		for i := 0; i < b.N; i++ {
			n, _ := handler.Send(sendPacket, rdst)
			c += n
		}
		b.SetBytes(int64(c))
	})
}
