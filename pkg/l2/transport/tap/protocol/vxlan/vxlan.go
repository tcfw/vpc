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
	conn []net.PacketConn
	port int
	recv protocol.HandlerFunc

	connCount int
}

//NewHandler creates a new VxLAN handler
func NewHandler() *Handler {
	return &Handler{
		connCount: runtime.NumCPU(),
		conn:      []net.PacketConn{},
		recv:      func(_ []*protocol.Packet) {},
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

	p.handleIn()

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
func (p *Handler) Send(packets []*protocol.Packet, rdst net.IP) (int, error) {
	n := 0
	for _, packet := range packets {
		vxlanFrame := NewPacket(packet.VNID, 0, packet.Frame)
		addr := &net.UDPAddr{IP: rdst, Port: p.port}
		i := hash(packet, rdst, p.connCount)
		ni, err := p.conn[i].WriteTo(vxlanFrame.Bytes(), addr)
		if err != nil {
			return n, err
		}
		n += ni
	}
	return n, nil
}

//SetHandler sets the receiving callback
func (p *Handler) SetHandler(handle protocol.HandlerFunc) {
	p.recv = handle
}

//handleIn handles picking up pakcets from the underlying udp connection per thread
func (p *Handler) handleIn() {
	for _, conn := range p.conn {
		go func(c net.PacketConn) {
			buff := make([]byte, 81920)
			for {
				n, _, err := c.ReadFrom(buff)
				if err != nil {
					return
				}

				packet, _ := FromBytes(bytes.NewBuffer(buff[:n]))

				p.recv([]*protocol.Packet{protocol.NewPacket(packet.VNID, packet.InnerFrame)})
			}
		}(conn)
	}
}
