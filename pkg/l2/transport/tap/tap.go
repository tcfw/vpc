package tap

import (
	"fmt"
	"log"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"

	"github.com/tcfw/vpc/pkg/l2/transport"
	"github.com/vishvananda/netlink"

	"github.com/google/gopacket/pcap"
	"github.com/songgao/packets/ethernet"
	"github.com/songgao/water"
)

const (
	queueLen = 5000
)

//Tap provides communication between taps in bridges and other endpoints
type Tap struct {
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

const (
	vtepPattern = "t-%d"
)

//NewTap creates a tap interface and starts listening to it
func NewTap(s *Listener, vnid uint32, mtu int, br *netlink.Bridge) (*Tap, error) {
	tapIface := &netlink.Tuntap{
		Mode: netlink.TUNTAP_MODE_TAP,
		LinkAttrs: netlink.LinkAttrs{
			Name:        fmt.Sprintf(vtepPattern, vnid),
			MTU:         mtu,
			MasterIndex: br.Index,
		},
		Flags: netlink.TUNTAP_MULTI_QUEUE_DEFAULTS | netlink.TUNTAP_TUN_EXCL,
	}

	if err := netlink.LinkAdd(tapIface); err != nil {
		return nil, err
	}

	if err := netlink.LinkSetUp(tapIface); err != nil {
		return nil, err
	}

	handler, err := pcap.OpenLive(tapIface.Name, 65535, false, pcap.BlockForever)
	if err != nil {
		return nil, err
	}

	handler.SetDirection(pcap.DirectionOut)

	tapHandler := &Tap{
		vnid:    vnid,
		out:     make(chan ethernet.Frame, queueLen),
		in:      make(chan *Packet, queueLen),
		lis:     s,
		iface:   tapIface,
		mtu:     int(s.mtu),
		FDBMiss: make(chan transport.ForwardingMiss, 10),
		handler: handler,
	}

	config := water.Config{
		DeviceType: water.TAP,
	}
	config.Name = tapIface.Name
	config.MultiQueue = true

	ifce, err := water.New(config)
	if err != nil {
		return nil, fmt.Errorf("failed to init tun dev: %s", err)
	}
	tapHandler.tuntap = ifce

	return tapHandler, nil
}

//Stop closes off comms
func (v *Tap) Stop() {
	close(v.FDBMiss)
	close(v.in)
}

//Handle listens for packets on the tap and forwards them to the endpoint handler
func (v *Tap) Handle() {
	go v.pipeIn()

	buf := make([]byte, v.mtu)
	for {
		n, err := v.tuntap.Read(buf)
		if err != nil {
			close(v.out)
			return
		}

		v.lis.tx <- NewPacket(v.vnid, 0, buf[:n])
	}
}

//HandlePCAP uses PCAP to listen for packets and forward them to the endpoint handler
func (v *Tap) HandlePCAP() {
	go v.pipeIn()

	src := gopacket.NewPacketSource(v.handler, layers.LayerTypeEthernet)
	for {
		frame, ok := <-src.Packets()
		if !ok {
			return
		}
		v.lis.tx <- NewPacket(v.vnid, 0, frame.Data())
	}
}

//pipeInPCAP takes packets from the endpoint handler and writes them to the tap interface using PCAP
func (v *Tap) pipeInPCAP() {
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

//pipeIn takes packets from the endpoint handler and writes them to the tap interface
func (v *Tap) pipeIn() {
	for {
		packet, ok := <-v.in
		if !ok {
			return
		}

		v.tuntap.Write(packet.InnerFrame)
	}
}
