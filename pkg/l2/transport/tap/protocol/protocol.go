package protocol

import (
	"net"

	"github.com/songgao/packets/ethernet"
)

type Handler interface {
	Start() error
	Stop() error

	Send(packet *Packet, rdst net.Addr) error
	Recv() (*Packet, error)
}

type Packet struct {
	VNID  uint32
	Frame ethernet.Frame
}

func NewPacket(vnid uint32, frame []byte) *Packet {
	return &Packet{VNID: vnid, Frame: frame}
}
