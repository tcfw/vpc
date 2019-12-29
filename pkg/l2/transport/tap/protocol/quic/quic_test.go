package quic

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tcfw/vpc/pkg/l2/transport/tap/protocol"
)

func TestSendRecv(t *testing.T) {
	handler := NewHandler()

	cliHandler := &Handler{
		epConns: map[string]*epConn{},
		in:      make(chan *protocol.Packet, 1000),
		port:    18443,
	}

	if !assert.NoError(t, handler.Start()) {
		return
	}

	frame := []byte{0x10, 0x20}
	var vnid uint32 = 5
	sendPacket := &protocol.Packet{VNID: vnid, Frame: frame}
	rdst := net.ParseIP("::1")

	_, err := cliHandler.Send(sendPacket, rdst)
	if !assert.NoError(t, err) {
		return
	}

	recvPacket, err := handler.Recv()
	if assert.NoError(t, err) {
		assert.Equal(t, frame, []byte(recvPacket.Frame))
		assert.Equal(t, vnid, recvPacket.VNID)
	}
}

func BenchmarkQuicSend(b *testing.B) {
	handler := NewHandler()
	handler.Start()

	cliHandler := &Handler{
		epConns: map[string]*epConn{},
		in:      make(chan *protocol.Packet, 1000),
		port:    18443,
	}

	frame := []byte{
		0xff, 0x1, 0x2, 0x3, 0x1, 0x2, 0x3, 0x1, 0x2, 0x3, 0x1, 0x2, 0x3, 0x1, 0x2, 0x3,
		0xff, 0x1, 0x2, 0x3, 0x1, 0x2, 0x3, 0x1, 0x2, 0x3, 0x1, 0x2, 0x3, 0x1, 0x2, 0x3,
	}
	packet := &protocol.Packet{VNID: 5, Frame: frame}
	rdst := net.ParseIP("::1")

	go func() {
		//ensure we're still receiving
		for {
			handler.Recv()
		}
	}()

	b.Run("send", func(b *testing.B) {
		c := 0
		for i := 0; i < b.N; i++ {
			n, _ := cliHandler.Send(packet, rdst)
			c += n
		}
		b.SetBytes(int64(c))
	})
}
