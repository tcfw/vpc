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

	// var buf bytes.Buffer
	var frame ethernet.Frame
	for {
		// buf.Reset()
		frame.Resize(v.mtu)

		// _, err := buf.ReadFrom(v.tuntap) //TODO(tcfw) metrics?
		n, err := v.tuntap.Read([]byte(frame))
		if err != nil {
			close(v.out)
			return
		}

		frame = frame[:n]

		// p := NewPacket(v.vnid, 0, buf.Bytes())
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
