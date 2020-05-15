package tap

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"sync"

	"github.com/google/gopacket/layers"
	"github.com/songgao/packets/ethernet"
	"github.com/tcfw/vpc/pkg/l2/controller"

	"github.com/tcfw/vpc/pkg/l2/transport/tap/protocol/vxlan"
	// "github.com/tcfw/vpc/pkg/l2/transport/tap/protocol/quic"

	"github.com/tcfw/vpc/pkg/l2/transport/tap/protocol"

	"github.com/tcfw/vpc/pkg/l2/transport"

	"github.com/vishvananda/netlink"
)

//Listener holds all vtep VNIs
type Listener struct {
	mu    sync.Mutex
	taps  map[uint32]*Tap
	conn  protocol.Handler
	mtu   int32
	FDB   *FDB
	sdn   controller.Controller
	vlans map[uint16]int
}

//NewListener inits a new VTEP style listener
func NewListener() (*Listener, error) {
	lis := &Listener{
		taps:  map[uint32]*Tap{},
		mtu:   1500,
		FDB:   NewFDB(),
		vlans: map[uint16]int{},
		conn:  vxlan.NewHandler(),
	}

	return lis, nil
}

//Start attaches the conn handler
func (s *Listener) Start() error {
	s.conn.SetHandler(s.handleIn)
	return s.conn.Start()
}

//SetMTU sets the MTU of the listening device
func (s *Listener) SetMTU(mtu int32) error {
	if s.conn != nil {
		return fmt.Errorf("cannot set mtu after already started")
	}

	s.mtu = mtu
	return nil
}

//SetSDN adds the SDN controller for ip lookups
func (s *Listener) SetSDN(sdn controller.Controller) error {
	s.sdn = sdn
	return nil
}

//AddEP adds a vtep handler to a bridge
func (s *Listener) AddEP(vnid uint32, br *netlink.Bridge) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	tapHandler, err := NewTap(s, vnid, int(s.mtu), br)
	if err != nil {
		return err
	}

	s.taps[vnid] = tapHandler

	log.Println("registered new VTEP")

	go tapHandler.Handle()
	// go tapHandler.HandlePCAP()

	return nil
}

//DelEP stops & deletes VTEP activity
func (s *Listener) DelEP(vnid uint32) error {
	s.taps[vnid].Stop()

	netlink.LinkDel(s.taps[vnid].iface)

	delete(s.taps, vnid)

	return nil
}

func (s *Listener) handleIn(ps []*protocol.Packet) {
	for _, p := range ps {
		//Ignore unknown VNIDs
		tap, ok := s.taps[p.VNID]
		if !ok {
			return
		}

		tap.Write(p.Frame)
	}
}

//Send forwards or handles interally a set of packets
func (s *Listener) Send(ps []*protocol.Packet) {
	for _, p := range ps {
		s.sendOne(p)
	}
}

func (s *Listener) sendOne(packet *protocol.Packet) error {
	var frame ethernet.Frame = packet.Frame

	//Ignore all non-tagged frames
	if frame.Tagging() == ethernet.NotTagged {
		return fmt.Errorf("mismatch tags")
	}

	etherType := frame.Ethertype()
	dst := frame.Destination()

	//ARP reducer
	if etherType == ethernet.ARP {
		go func() {
			if err := s.arpReduce(packet); err != nil {
				log.Printf("failed arp reduce: %s", err)
			}
		}()
		return nil
	}

	//ICMPv6 NDP reducer - 0x3a = ICMPv6
	if etherType == ethernet.IPv6 && packet.Frame[24] == 0x3a {
		icmp6Type := frame.Payload()[40]
		if icmp6Type == layers.ICMPv6TypeNeighborSolicitation {
			go func() {
				if err := s.icmp6NDPReduce(packet); err != nil {
					log.Printf("failed ICMPv6 NDP reduce: %s", err)
				}
			}()
			return nil
		}
	}

	broadcast := net.HardwareAddr{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}

	if bytes.Compare(dst, broadcast) != 0 {
		addr := s.FDB.LookupMac(packet.VNID, dst)

		if addr == nil {
			s.taps[packet.VNID].FDBMiss <- transport.ForwardingMiss{Type: transport.MissTypeEP, HwAddr: dst}
			return nil
		}

		if _, err := s.conn.Send([]*protocol.Packet{packet}, addr); err != nil {
			return err
		}
	} else {
		//Flood to all EPs
		for _, addr := range s.FDB.ListBroadcast(packet.VNID) {
			if _, err := s.conn.Send([]*protocol.Packet{packet}, addr); err != nil {
				return err
			}
		}
	}
	return nil
}

//ForwardingMiss gets a readonly sub for FDB misses
func (s *Listener) ForwardingMiss(vnid uint32) (<-chan transport.ForwardingMiss, error) {
	vtep, ok := s.taps[vnid]
	if !ok {
		return nil, fmt.Errorf("unknown vnid %d", vnid)
	}

	return vtep.FDBMiss, nil
}

//AddForwardEntry adds an entry to the FDB
func (s *Listener) AddForwardEntry(vnid uint32, mac net.HardwareAddr, ip net.IP) error {
	s.FDB.AddEntry(vnid, mac, ip)
	return nil
}

//Status gets the status of the TAP
func (s *Listener) Status(vnid uint32) transport.EPStatus {
	return transport.EPStatusUP
}
