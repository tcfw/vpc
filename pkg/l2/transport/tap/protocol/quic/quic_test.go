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
	rdst := &net.UDPAddr{IP: net.ParseIP("::1"), Port: port}

	err := cliHandler.Send(sendPacket, rdst)
	if !assert.NoError(t, err) {
		return
	}

	recvPacket, err := handler.Recv()
	if assert.NoError(t, err) {
		assert.Equal(t, frame, []byte(recvPacket.Frame))
		assert.Equal(t, vnid, recvPacket.VNID)
	}
}
