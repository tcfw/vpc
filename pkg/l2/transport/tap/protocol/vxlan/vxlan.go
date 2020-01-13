package vxlan

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"runtime"
	"syscall"

	"golang.org/x/sys/unix"

	"github.com/tcfw/vpc/pkg/l2/transport/tap/protocol"
)

//Handler VxLAN endpoint protocol
type Handler struct {
	conn  []net.PacketConn
	in    chan []byte
	inBuf []byte
	port  int

	connCount int
}

//NewHandler creates a new VxLAN handler
func NewHandler() *Handler {
	return &Handler{
		connCount: runtime.NumCPU() * 2,
		conn:      []net.PacketConn{},
		inBuf:     make([]byte, 1500),
		in:        make(chan []byte, 1000),
		port:      4789,
	}
}

//Start opens a UDP endpoint for VxLAN packets
func (p *Handler) Start() error {
	for i := 0; i < p.connCount; i++ {
		lisConfig := net.ListenConfig{
			Control: func(network string, address string, c syscall.RawConn) error {
				var err error
				c.Control(func(fd uintptr) {
					err = unix.SetsockoptInt(int(fd), unix.SOL_SOCKET, unix.SO_REUSEADDR, 1)
					if err != nil {
						return
					}

					err = unix.SetsockoptInt(int(fd), unix.SOL_SOCKET, unix.SO_REUSEPORT, 1)
					if err != nil {
						return
					}
				})
				return err
			},
		}

		pc, err := lisConfig.ListenPacket(context.Background(), "udp", fmt.Sprintf(":%d", p.port))
		if err != nil {
			return fmt.Errorf("failed to add packet listener; %s", err)
		}

		p.conn = append(p.conn, pc)
	}

	go p.handleIn()

	return nil
}

//Stop ends the UDP endpoint
func (p *Handler) Stop() error {
	for _, conn := range p.conn {
		if err := conn.Close(); err != nil {
			return err
		}
	}
	return nil
}

//Send sends a single packet to a VxLAN endpoinp
func (p *Handler) Send(packet *protocol.Packet, rdst net.IP) (int, error) {
	vxlanFrame := NewPacket(packet.VNID, 0, packet.Frame)
	addr := &net.UDPAddr{IP: rdst, Port: p.port}
	i := hash(packet, rdst, p.connCount)
	n, err := p.conn[i].WriteTo(vxlanFrame.Bytes(), addr)
	return n, err
}

func (p *Handler) handleIn() {
	for _, conn := range p.conn {
		go func(c net.PacketConn) {
			for {
				n, _, err := c.ReadFrom(p.inBuf)
				if err != nil {
					close(p.in)
					return
				}

				p.in <- p.inBuf[:n]
			}
		}(conn)
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
