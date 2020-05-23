package tap

import (
	"fmt"
	"hash/maphash"
	"log"
	"sync"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/vishvananda/netlink"

	"github.com/tcfw/vpc/pkg/l2/transport/tap/protocol"
)

const (
	vtepPattern = "t-%d"
	tapQueues   = 2
)

//Nic basic NIC that can send a receive packets to and from a compute service
type Nic interface {
	Stop() error
	Start()
	Write(b []byte) (int, error)
	Delete() error
	SetHandler(protocol.HandlerFunc)
	IFace() netlink.Link
}

//Tap provides communication between taps in bridges and other endpoints
type Tap struct {
	vnid        uint32
	tuntaps     []*tuntapDev
	pCapHandler *pcap.Handle
	iface       netlink.Link
	mtu         int
	tx          protocol.HandlerFunc
}

//NewTap creates a tap interface on the specified bridge
func NewTap(s *Listener, vnid uint32, mtu int, br *netlink.Bridge) (*Tap, error) {
	iface := &netlink.Tuntap{
		Mode: netlink.TUNTAP_MODE_TAP,
		LinkAttrs: netlink.LinkAttrs{
			Name:        fmt.Sprintf(vtepPattern, vnid),
			MTU:         mtu,
			MasterIndex: br.Index,
		},
		Flags: netlink.TUNTAP_MULTI_QUEUE_DEFAULTS,
	}

	if err := netlink.LinkAdd(iface); err != nil {
		return nil, err
	}

	if err := netlink.LinkSetTxQLen(iface, 10000); err != nil {
		return nil, err
	}

	tapHandler := &Tap{
		vnid:    vnid,
		iface:   iface,
		mtu:     int(s.mtu),
		tuntaps: []*tuntapDev{},
	}

	//Create 2 listening queues through multiqueue tap option
	for i := 0; i < tapQueues; i++ {
		ifce, err := tapHandler.openDev(iface.Name)
		if err != nil {
			log.Fatal(err)
		}
		tapHandler.tuntaps = append(tapHandler.tuntaps, ifce)
	}

	return tapHandler, nil
}

//Stop closes off comms
func (v *Tap) Stop() error {
	return netlink.LinkSetDown(v.iface)
}

//Start listeners for each tap fd created
func (v *Tap) Start() {
	netlink.LinkSetUp(v.iface)

	var wg sync.WaitGroup

	for _, tap := range v.tuntaps {
		go v.handleTapPipe(&wg, tap)
	}

	wg.Wait()
}

//handleTapPipe listens for packets on a tap and forwards them to the endpoint handler
func (v *Tap) handleTapPipe(wg *sync.WaitGroup, tap *tuntapDev) {
	wg.Add(1)
	defer wg.Done()

	buf := make([]byte, v.mtu+1024)
	for {
		n, err := tap.Read(buf)
		if err != nil {
			return
		}

		//TODO(tcfw) add to drop metrics on error
		v.tx([]*protocol.Packet{protocol.NewPacket(v.vnid, buf[:n])})
	}
}

//handlePCAP uses PCAP to listen for packets and forward them to the endpoint handler
func (v *Tap) handlePCAP() {
	src := gopacket.NewPacketSource(v.pCapHandler, layers.LayerTypeEthernet)
	for {
		frame, ok := <-src.Packets()
		if !ok {
			return
		}

		v.tx([]*protocol.Packet{protocol.NewPacket(v.vnid, frame.Data())})
	}
}

//SetHandler applies the listener callback for handling TX
func (v *Tap) SetHandler(h protocol.HandlerFunc) {
	v.tx = h
}

//Write sends a packet to the bridge
func (v *Tap) Write(b []byte) (int, error) {
	return v.tuntaps[v.woHash(b)].WriteRaw(b)
}

func (v *Tap) woHash(b []byte) int {
	h := maphash.Hash{}
	h.Write(b[:12]) //hash the src & dst l2
	return int(h.Sum64() % uint64(tapQueues))
}

//Delete removes the actual nic from the OS
func (v *Tap) Delete() error {
	return netlink.LinkDel(v.iface)
}

//IFace returns the underlying OS NIC
func (v *Tap) IFace() netlink.Link {
	return v.iface
}
