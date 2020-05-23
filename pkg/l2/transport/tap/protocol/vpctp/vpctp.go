package vpctp

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"net"
	"syscall"

	"github.com/tcfw/vpc/pkg/l2/transport/tap/protocol"
	"golang.org/x/sys/unix"
)

//NewHandler creates a new VxLAN handler
func NewHandler() *VPCTP {
	return &VPCTP{
		listenAddr: "",
	}
}

//VPCTP VPC tunnel protocol raw socket handler
type VPCTP struct {
	sock net.PacketConn
	recv protocol.HandlerFunc

	listenAddr string

	buf bytes.Buffer
}

//Start starts listening for remote packets coming in
func (vtp *VPCTP) Start() error {
	lisConfig := net.ListenConfig{
		Control: func(network string, address string, c syscall.RawConn) error {
			var err error
			c.Control(func(fd uintptr) {
				err = unix.SetsockoptInt(int(fd), unix.SOL_SOCKET, unix.SO_PRIORITY, 0)
				if err != nil {
					return
				}

				// err = unix.SetsockoptInt(int(fd), unix.SOL_SOCKET, unix.SO_REUSEPORT, 1)
				// if err != nil {
				// 	return
				// }

				// err = unix.SetNonblock(int(fd), true)
				// if err != nil {
				// 	return
				// }
			})
			return err
		},
	}
	conn, err := lisConfig.ListenPacket(context.Background(), "ip4:172", vtp.listenAddr) //raw IP sock on IANA protocol 172
	if err != nil {
		return fmt.Errorf("failed to open raw sock: %s", err)
	}
	vtp.sock = conn

	go vtp.listen()

	return nil
}

//Stop stop the listening socket
func (vtp *VPCTP) Stop() error {
	return vtp.sock.Close()
}

//Send sends a group of packets to the specified destination
func (vtp *VPCTP) Send(ps []*protocol.Packet, addr net.IP) (int, error) {
	for _, p := range ps {
		vtp.SendOne(p, addr)
	}
	return 0, nil
}

//SendOne sends a single packet to the specified destination
func (vtp *VPCTP) SendOne(p *protocol.Packet, addr net.IP) (int, error) {
	raddr := &net.IPAddr{IP: addr}

	return vtp.sock.WriteTo(vtp.toBytes(p), raddr)
}

//SetHandler sets the receiving handler callback func
func (vtp *VPCTP) SetHandler(h protocol.HandlerFunc) {
	vtp.recv = h
}

func (vtp *VPCTP) listen() {
	buf := make([]byte, 9500)
	for {
		n, _, err := vtp.sock.ReadFrom(buf[0:])
		if err != nil {
			return
		}
		vtp.recv([]*protocol.Packet{vtp.toPacket(buf[:n])})
	}
}

func (vtp *VPCTP) toBytes(p *protocol.Packet) []byte {
	vtp.buf.Reset()
	slen := uint16(len(p.Source))

	vtp.buf.WriteByte(byte(p.VNID))
	vtp.buf.WriteByte(byte(p.VNID >> 8))
	vtp.buf.WriteByte(byte(p.VNID >> 16))
	vtp.buf.WriteByte(byte(p.VNID >> 24))

	vtp.buf.WriteByte(byte(slen))
	vtp.buf.WriteByte(byte(slen >> 8))
	vtp.buf.Write(p.Source)

	vtp.buf.Write(p.Frame)

	return vtp.buf.Bytes()
}

func (vtp *VPCTP) toPacket(b []byte) *protocol.Packet {
	pack := &protocol.Packet{}

	//VNID
	pack.VNID = binary.LittleEndian.Uint32(b[:4])

	//Source
	slen := binary.LittleEndian.Uint16(b[4:6])
	pack.Source = b[6 : 6+slen]

	//Frame
	pack.Frame = b[6+slen:]

	return pack
}
