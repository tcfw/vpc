package transport

import (
	"net"

	"github.com/songgao/packets/ethernet"
	"github.com/songgao/water"
)

type vtep struct {
	out     chan ethernet.Frame //from bridge
	in      chan *Packet        //from vtep
	vnid    uint32
	tuntap  *water.Interface
	mtu     int
	lis     *Listener
	FDBMiss chan net.HardwareAddr
}

func (v *vtep) Stop() {
	close(v.FDBMiss)
	close(v.in)
}

func (v *vtep) Handle() {
	go v.pipeIn()

	for {
		frame := make([]byte, v.mtu)
		_, err := v.tuntap.Read(frame) //TODO(tcfw) metrics?
		if err != nil {
			close(v.out)
			return
		}

		p := NewPacket(v.vnid, 0, frame)
		v.lis.tx <- p
	}
}

func (v *vtep) pipeIn() {
	for {
		packet, ok := <-v.in
		if !ok {
			return
		}

		v.tuntap.Write(packet.InnerFrame)
	}
}
