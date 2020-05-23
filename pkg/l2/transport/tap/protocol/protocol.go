package protocol

import "net"

//Handler provides methods of forwarding frames to another endpoint
type Handler interface {
	Start() error
	Stop() error

	Send([]*Packet, net.IP) (int, error)
	SendOne(*Packet, net.IP) (int, error)
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

//TestHandler provides a way to track and drop output packets
//and simulate inbound packets
type TestHandler struct {
	recv          HandlerFunc
	handledCount  int
	handled       []*Packet
	StoreOutbound bool
}

//Stats provides the handled packet count and any stored packets
func (th *TestHandler) Stats() (int, []*Packet) {
	return th.handledCount, th.handled
}

//Start - Fulfil interface
func (th *TestHandler) Start() error {
	return nil
}

//Stop - Fulfil interface
func (th *TestHandler) Stop() error {
	return nil
}

//Send records to the handled
func (th *TestHandler) Send(packets []*Packet, rdst net.IP) (int, error) {
	for _, packet := range packets {
		th.SendOne(packet, rdst)
	}
	return 0, nil
}

//SendOne a single record to the handled
func (th *TestHandler) SendOne(packet *Packet, rdst net.IP) (int, error) {
	th.handledCount++
	if th.StoreOutbound {
		th.handled = append(th.handled, packet)
	}
	return 0, nil
}

//SetHandler sets the TX handler for l2 forwarding
func (th *TestHandler) SetHandler(h HandlerFunc) {
	th.recv = h
}

//Receive allows for packet injection into the tap as if coming from OS
func (th *TestHandler) Receive(p []*Packet) error {
	th.recv(p)
	return nil
}
