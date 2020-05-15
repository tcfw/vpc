package protocol

import "net"

//Handler provides methods of forwarding frames to another endpoint
type Handler interface {
	Start() error
	Stop() error

	Send(packet []*Packet, rdst net.IP) (int, error)
	SetHandler(HandlerFunc)
}

//Packet holds the ethernet frame and the desired VNID
type Packet struct {
	VNID   uint32
	Source []byte
	Frame  []byte
}

//NewPacket constructs a new packet give an ethernet frame and desired VNID
func NewPacket(vnid uint32, frame []byte) *Packet {
	return &Packet{VNID: vnid, Frame: frame}
}

//HandlerFunc allows callbacks when receiving packets from a remote source
type HandlerFunc func([]*Packet)
