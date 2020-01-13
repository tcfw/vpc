package tap

import (
	"fmt"
	"log"
	"runtime"
	"sync"

	"github.com/tcfw/vpc/pkg/l2/transport/tap/protocol"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"

	"github.com/tcfw/vpc/pkg/l2/transport"
	"github.com/vishvananda/netlink"

	"github.com/google/gopacket/pcap"
	"github.com/songgao/packets/ethernet"
)

const (
	queueLen    = 1000
	vtepPattern = "t-%d"
)

//Tap provides communication between taps in bridges and other endpoints
type Tap struct {
	out  chan *ethernet.Frame  //from bridge (rx)
	in   chan *protocol.Packet //from protocol endpoint (tx)
	vnid uint32
	// tuntaps []*water.Interface
	tuntaps []*tuntapDev
	handler *pcap.Handle
	iface   netlink.Link
	mtu     int
	lis     *Listener
	FDBMiss chan transport.ForwardingMiss
}

//NewTap creates a tap interface and starts listening to it
func NewTap(s *Listener, vnid uint32, mtu int, br *netlink.Bridge) (*Tap, error) {
	tapIface := &netlink.Tuntap{
		Mode: netlink.TUNTAP_MODE_TAP,
		LinkAttrs: netlink.LinkAttrs{
			Name:        fmt.Sprintf(vtepPattern, vnid),
			MTU:         mtu,
			MasterIndex: br.Index,
			TxQLen:      1000,
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
		out:     make(chan *ethernet.Frame, queueLen),
		in:      make(chan *protocol.Packet, queueLen),
		lis:     s,
		iface:   tapIface,
		mtu:     int(s.mtu),
		FDBMiss: make(chan transport.ForwardingMiss, 10),
		tuntaps: []*tuntapDev{},
	}

	//Create 2 listening queues through multiqueue tap option
	for range [2]int{} {
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
	close(v.in)
}

//Handle starts listeners for each tap fd created
func (v *Tap) Handle() {
	go v.pipeIn()

	var wg sync.WaitGroup

	for _, tap := range v.tuntaps {
		go v.handleTapPipe(&wg, tap)
	}

	wg.Wait()
	close(v.out)
}

//handleTapPipe listens for packets on a tap and forwards them to the endpoint handler
func (v *Tap) handleTapPipe(wg *sync.WaitGroup, tap *tuntapDev) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	wg.Add(1)
	defer wg.Done()

	buf := make([]byte, v.mtu+1024)
	for {
		n, err := tap.Read(buf)
		if err != nil {
			return
		}

		frame := buf[:n]

		v.lis.tx <- protocol.NewPacket(v.vnid, frame)
	}
}

//HandlePCAP uses PCAP to listen for packets and forward them to the endpoint handler
func (v *Tap) HandlePCAP() {
	go v.pipeIn()
	defer close(v.out)

	src := gopacket.NewPacketSource(v.handler, layers.LayerTypeEthernet)
	for {
		frame, ok := <-src.Packets()
		if !ok {
			return
		}
		v.lis.tx <- protocol.NewPacket(v.vnid, frame.Data())
	}
}

//pipeIn takes packets from the endpoint handler and writes them to the tap interface
func (v *Tap) pipeIn() {
	for {
		packet, ok := <-v.in
		if !ok {
			return
		}

		v.Write(packet.Frame)
	}
}

//Write sends a packet to the bridge
func (v *Tap) Write(b []byte) (int, error) {
	return v.tuntaps[0].WriteRaw(b)
}
