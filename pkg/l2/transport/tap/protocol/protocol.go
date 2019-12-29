package protocol

import "net"

type Handler interface {
	Start() error
	Stop() error

	Send(packet *Packet, rdst net.IP) (int, error)
	Recv() (*Packet, error)
}

type Packet struct {
	VNID  uint32
	Frame []byte
}

func NewPacket(vnid uint32, frame []byte) *Packet {
	return &Packet{VNID: vnid, Frame: frame}
}
