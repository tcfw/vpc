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

	handler.SetHandler(func(ps []protocol.Packet) {
		if len(ps) == 0 {
			return
		}

		recvPacket = &ps[0]
	})

	cliHandler.SetHandler(func(ps []protocol.Packet) {
		if len(ps) == 0 {
			return
		}

		recvPacket = &ps[0]
	})

	frame := []byte{0x10, 0x20}
	var vnid uint32 = 5
	sendPacket := protocol.Packet{VNID: vnid, Frame: frame}
	rdst := net.ParseIP("::1")

	_, err := cliHandler.Send([]protocol.Packet{sendPacket}, rdst)
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

	handler.SetHandler(func(ps []protocol.Packet) {
		received += len(ps)
	})

	frame := []byte{
		0xCC, 0x03, 0x04, 0xDC, 0x00, 0x10, 0xFF, 0xAA, 0xFA, 0xAA, 0xFF, 0xAA, 0x81, 0x00, 0x00,
		0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x7F, 0x00, 0x00, 0x01, 0x08, 0x08, 0x08, 0x08, 0x00, 0x08, 0x00, 0x00, 0x00, 0x01, 0x00,
		0x01, 0x30, 0x31, 0x32, 0x33, 0x34, 0x35, 0x36, 0x37, 0x38, 0x39, 0x61, 0x62, 0x63, 0x64,
	}

	var vnid uint32 = 5
	sendPacket := []protocol.Packet{{VNID: vnid, Frame: frame}}
	rdst := net.ParseIP("::1")

	b.Run("send", func(b *testing.B) {
		b.SetBytes(int64(len(frame) + 10)) //See header format
		for i := 0; i < b.N; i++ {
			handler.Send(sendPacket, rdst)
		}
	})
}
