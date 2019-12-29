package vxlan

import (
	"bytes"
	"fmt"
	"net"

	"github.com/tcfw/vpc/pkg/l2/transport/tap/protocol"
)

//Handler VxLAN endpoint protocol
type Handler struct {
	conn  net.PacketConn
	in    chan []byte
	inBuf []byte
	port  int
}

//NewHandler creates a new VxLAN handler
func NewHandler() *Handler {
	return &Handler{
		inBuf: make([]byte, 1500),
		in:    make(chan []byte, 1000),
		port:  4789,
	}
}

//Start opens a UDP endpoint for VxLAN packets
func (p *Handler) Start() error {
	pc, err := net.ListenPacket("udp", fmt.Sprintf(":%d", p.port))
	if err != nil {
		return err
	}

	p.conn = pc

	go p.handleIn()

	return nil
}

//Stop ends the UDP endpoint
func (p *Handler) Stop() error {
	return p.conn.Close()
}

//Send sends a single packet to a VxLAN endpoinp
func (p *Handler) Send(packet *protocol.Packet, rdst net.IP) (int, error) {
	vxlanFrame := NewPacket(packet.VNID, 0, packet.Frame)
	addr := &net.UDPAddr{IP: rdst, Port: p.port}
	n, err := p.conn.WriteTo(vxlanFrame.Bytes(), addr)
	return n, err
}

func (p *Handler) handleIn() {
	for {
		n, _, err := p.conn.ReadFrom(p.inBuf)
		if err != nil {
			close(p.in)
			return
		}

		p.in <- p.inBuf[:n]
	}
}

//Recv wait for a new packet
func (p *Handler) Recv() (*protocol.Packet, error) {
	raw, ok := <-p.in
	if !ok {
		return nil, fmt.Errorf("channel closed")
	}

	packet, err := FromBytes(bytes.NewBuffer(raw))
	if err != nil {
		return nil, err
	}

	return protocol.NewPacket(packet.VNID, packet.InnerFrame), nil
}
