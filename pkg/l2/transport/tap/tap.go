package tap

import (
	"fmt"
	"log"
	"sync"

	"github.com/tcfw/vpc/pkg/l2/transport/tap/protocol"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"

	"github.com/tcfw/vpc/pkg/l2/transport"
	"github.com/vishvananda/netlink"

	"github.com/google/gopacket/pcap"

	"hash/maphash"
)

const (
	vtepPattern = "t-%d"
	tapQueues   = 2
)

//Tap provides communication between taps in bridges and other endpoints
type Tap struct {
	vnid uint32
	// tuntaps []*water.Interface
	tuntaps     []*tuntapDev
	pCapHandler *pcap.Handle
	iface       netlink.Link
	mtu         int
	lis         *Listener
	FDBMiss     chan transport.ForwardingMiss
}

//NewTap creates a tap interface and starts listening to it
func NewTap(s *Listener, vnid uint32, mtu int, br *netlink.Bridge) (*Tap, error) {
	tapIface := &netlink.Tuntap{
		Mode: netlink.TUNTAP_MODE_TAP,
		LinkAttrs: netlink.LinkAttrs{
			Name:        fmt.Sprintf(vtepPattern, vnid),
			MTU:         mtu,
			MasterIndex: br.Index,
			TxQLen:      10000,
		},
		Flags: netlink.TUNTAP_MULTI_QUEUE_DEFAULTS,
	}

	if err := netlink.LinkAdd(tapIface); err != nil {
		return nil, err
	}

	if err := netlink.LinkSetUp(tapIface); err != nil {
		return nil, err
	}

	tapHandler := &Tap{
		vnid:    vnid,
		lis:     s,
		iface:   tapIface,
		mtu:     int(s.mtu),
		FDBMiss: make(chan transport.ForwardingMiss, 10),
		tuntaps: []*tuntapDev{},
	}

	//Create 2 listening queues through multiqueue tap option
	for i := 0; i < tapQueues; i++ {
		ifce, err := tapHandler.openDev(tapIface.Name)
		if err != nil {
			log.Fatal(err)
		}
		tapHandler.tuntaps = append(tapHandler.tuntaps, ifce)
	}

	return tapHandler, nil
}

//Stop closes off comms
func (v *Tap) Stop() {
	close(v.FDBMiss)
}

//Handle starts listeners for each tap fd created
func (v *Tap) Handle() {
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

		frame := buf[:n]

		//TODO(tcfw) add to drop metrics on error
		v.lis.sendOne(protocol.NewPacket(v.vnid, frame))
	}
}

//HandlePCAP uses PCAP to listen for packets and forward them to the endpoint handler
func (v *Tap) HandlePCAP() {
	src := gopacket.NewPacketSource(v.pCapHandler, layers.LayerTypeEthernet)
	for {
		frame, ok := <-src.Packets()
		if !ok {
			return
		}

		v.lis.sendOne(protocol.NewPacket(v.vnid, frame.Data()))
	}
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
