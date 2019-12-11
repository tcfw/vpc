package vxlan

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tcfw/vpc/pkg/l2/transport/tap/protocol"
)

func TestSendRecv(t *testing.T) {
	handler := NewHandler()

	if !assert.NoError(t, handler.Start()) {
		return
	}

	cliHandler := &Handler{
		port: 4790,
	}
	if !assert.NoError(t, cliHandler.Start()) {
		return
	}

	frame := []byte{0x10, 0x20}
	var vnid uint32 = 5
	sendPacket := &protocol.Packet{VNID: vnid, Frame: frame}
	rdst := &net.UDPAddr{IP: net.ParseIP("::1"), Port: handler.port}

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
