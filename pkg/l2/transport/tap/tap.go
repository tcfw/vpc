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
	"github.com/tcfw/vpc/pkg/l2/xdp"
)

const (
	vtepPattern = "t-%d"
	tapQueues   = 1
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
	xdp         []*xdp.Tap
}

//NewTap creates a tap interface on the specified bridge
func NewTap(s *Listener, vnid uint32, mtu int, br *netlink.Bridge) (*Tap, error) {
	// iface := &netlink.Tuntap{
	// 	Mode: netlink.TUNTAP_MODE_TAP,
	// 	LinkAttrs: netlink.LinkAttrs{
	// 		Name: fmt.Sprintf(vtepPattern, vnid),
	// 		MTU:  mtu,
	// 	},
	// 	Flags:  netlink.TUNTAP_MULTI_QUEUE_DEFAULTS,
	// 	Queues: tapQueues,
	// }
	iface := &netlink.Veth{
		LinkAttrs: netlink.LinkAttrs{
			Name: fmt.Sprintf(vtepPattern, vnid),
			MTU:  mtu,
		},
		PeerName: fmt.Sprintf(vtepPattern+"-p", vnid),
	}

	if err := netlink.LinkAdd(iface); err != nil {
		return nil, fmt.Errorf("failed to add link: %s", err)
	}

	ifacePeer, err := netlink.LinkByName(iface.PeerName)
	if err != nil {
		return nil, err
	}

	if br != nil {
		if err := netlink.LinkSetMaster(iface, br); err != nil {
			return nil, fmt.Errorf("failed to set link master: %s", err)
		}
	}

	//Attempt to set higher queue size, ignore failures
	netlink.LinkSetTxQLen(iface, 10000)
	// netlink.SetPromiscOn(iface)

	if err := netlink.LinkSetUp(iface); err != nil {
		return nil, fmt.Errorf("failed to bring up tap: %s", err)
	}

	netlink.LinkSetUp(ifacePeer)

	tapHandler := &Tap{
		vnid:    vnid,
		iface:   iface,
		mtu:     mtu,
		tuntaps: []*tuntapDev{},
		xdp:     []*xdp.Tap{},
	}

	//Create i queues through multiqueue tap option
	for i := 0; i < tapQueues; i++ {
		// ifce, err := tapHandler.openDev(iface.Name)
		// if err != nil {
		// 	log.Fatal(err)
		// }
		// tapHandler.tuntaps = append(tapHandler.tuntaps, ifce)
		xdp, err := xdp.NewTap(ifacePeer, i, nil)
		if err != nil {
			netlink.LinkDel(iface)
			return nil, err
		}
		tapHandler.xdp = append(tapHandler.xdp, xdp)
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

	for _, xdp := range v.xdp {
		go v.handleBatchXDP(&wg, xdp)
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
		v.tx([]protocol.Packet{protocol.NewPacket(v.vnid, buf[:n])})
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

		v.tx([]protocol.Packet{protocol.NewPacket(v.vnid, frame.Data())})
	}
}

func (v *Tap) handleXDP(wg *sync.WaitGroup, xdpt *xdp.Tap) {
	wg.Add(1)
	defer wg.Done()

	descs := make([]xdp.BatchDesc, 64)
	for _, desc := range descs {
		desc.Data = make([]byte, 2048)
	}

	packs := make([]protocol.Packet, 64)

	for {
		_, n, err := xdpt.BatchRead(descs)
		if err != nil {
			log.Printf("Failed to xdp read: %s", err)
			return
		}

		for i := 0; i < n; i++ {
			packs[i] = protocol.NewPacket(v.vnid, descs[i].Data[:descs[i].Len])
		}

		//TODO(tcfw) add to drop metrics on error
		v.tx(packs[:n])
	}
}

func (v *Tap) handleBatchXDP(wg *sync.WaitGroup, xdpt *xdp.Tap) {
	wg.Add(1)
	defer wg.Done()

	//Preallocate
	descs := make([]xdp.BatchDesc, 64)
	for i := range descs {
		descs[i] = xdp.BatchDesc{
			Data: make([]byte, v.mtu),
		}
	}
	packs := make([]protocol.Packet, 2048)
	for i := range packs {
		packs[i] = protocol.Packet{
			VNID:  v.vnid,
			Frame: make([]byte, v.mtu),
		}
	}

	var i int
	for {
		_, dn, err := xdpt.BatchRead(descs)
		if err != nil {
			return
		}

		for i = 0; i < dn; i++ {
			packs[i].Frame = descs[i].Data[:descs[i].Len]
		}

		v.tx(packs[:dn])
	}
}

//SetHandler applies the listener callback for handling TX
func (v *Tap) SetHandler(h protocol.HandlerFunc) {
	v.tx = h
}

//Write sends a packet to the bridge
func (v *Tap) Write(b []byte) (int, error) {
	return v.xdp[v.woHash(b)].Write(b)
}

//WriteBatch writes batches of packets to a single XDP queue
//Queue selection is based on the first packet provided
func (v *Tap) WriteBatch(bs []xdp.BatchDesc) (int, error) {
	return v.xdp[v.woHash(bs[0].Data[:bs[0].Len])].BatchWrite(bs)
}

func (v *Tap) woHash(b []byte) int {
	h := maphash.Hash{}
	h.Write(b[:12]) //hash the src & dst l2
	return int(h.Sum64() % uint64(tapQueues))
}

//Delete removes the actual nic from the OS
func (v *Tap) Delete() error {
	return netlink.LinkDel(v.iface)
	// return nil
}

//IFace returns the underlying OS NIC
func (v *Tap) IFace() netlink.Link {
	return v.iface
}
