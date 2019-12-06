package tap

import (
	"log"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"

	"github.com/tcfw/vpc/pkg/l2/transport"
	"github.com/vishvananda/netlink"

	"github.com/google/gopacket/pcap"
	"github.com/songgao/packets/ethernet"
	"github.com/songgao/water"
)

type tap struct {
	out     chan ethernet.Frame //from bridge
	in      chan *Packet        //from tap
	vnid    uint32
	tuntap  *water.Interface
	handler *pcap.Handle
	iface   netlink.Link
	mtu     int
	lis     *Listener
	FDBMiss chan transport.ForwardingMiss
}

func (v *tap) Stop() {
	close(v.FDBMiss)
	close(v.in)
}

func (v *tap) Handle() {
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

func (v *tap) HandlePCAP() {
	go v.pipeIn()

	src := gopacket.NewPacketSource(v.handler, layers.LayerTypeEthernet)

	for {
		// frame, _, err := v.handler.ZeroCopyReadPacketData()
		frame, ok := <-src.Packets()
		if !ok {
			return
		}
		v.lis.tx <- NewPacket(v.vnid, 0, frame.Data())
	}
}

func (v *tap) pipeInPCAP() {
	for {
		packet, ok := <-v.in
		if !ok {
			return
		}

		if err := v.handler.WritePacketData(packet.InnerFrame); err != nil {
			log.Printf("failed to write packet")
		}
	}
}

func (v *tap) pipeIn() {
	for {
		packet, ok := <-v.in
		if !ok {
			return
		}

		v.tuntap.Write(packet.InnerFrame)
	}
}
