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
	// "github.com/tcfw/vpc/pkg/l2/transport/tap/protocol/vpctp"
	// "github.com/tcfw/vpc/pkg/l2/transport/tap/protocol/quic"

	"github.com/tcfw/vpc/pkg/l2/transport/tap/protocol"

	"github.com/tcfw/vpc/pkg/l2/transport"

	"github.com/vishvananda/netlink"
)

//Listener holds all vtep VNIs
type Listener struct {
	mu     sync.Mutex
	taps   map[uint32]Nic
	misses map[uint32]chan transport.ForwardingMiss
	conn   protocol.Handler
	mtu    int32
	FDB    *FDB
	sdn    controller.Controller
	vlans  map[uint16]int
}

//NewListener inits a new VTEP style listener
func NewListener() (*Listener, error) {
	lis := &Listener{
		taps:   map[uint32]Nic{},
		mtu:    1500,
		FDB:    NewFDB(),
		vlans:  map[uint16]int{},
		misses: map[uint32]chan transport.ForwardingMiss{},
		conn:   vxlan.NewHandler(),
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

	tapHandler.SetHandler(s.Send)

	s.taps[vnid] = tapHandler
	s.misses[vnid] = make(chan transport.ForwardingMiss, 10)

	log.Println("registered new VTEP")

	go tapHandler.Start()
	// go tapHandler.HandlePCAP()

	return nil
}

//DelEP stops & deletes VTEP activity
func (s *Listener) DelEP(vnid uint32) error {
	tap, ok := s.taps[vnid]
	if !ok {
		return fmt.Errorf("failed to find tap for vnid")
	}

	tap.Stop()
	tap.Delete()
	delete(s.taps, vnid)
	delete(s.misses, vnid)

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
	frame := ethernet.Frame(packet.Frame)

	//Ignore all non-tagged frames
	if frame.Tagging() != ethernet.Tagged {
		return fmt.Errorf("mismatch vlan tags")
	}

	etherType := frame.Ethertype()
	dst := frame.Destination()

	if etherType == ethernet.ARP { //ARP reduce
		go func() {
			if err := s.arpReduce(packet); err != nil {
				log.Printf("failed arp reduce: %s", err)
			}
		}()
		return nil
	} else if etherType == ethernet.IPv6 && packet.Frame[24] == 0x3a { //ICMPv6 NDP reducer - 0x3a = ICMPv6
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

	if bytes.Compare(dst, net.HardwareAddr{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}) != 0 {
		addr := s.FDB.LookupMac(packet.VNID, dst)

		if addr == nil {
			go func() {
				s.misses[packet.VNID] <- transport.ForwardingMiss{Type: transport.MissTypeEP, HwAddr: dst}
			}()
			return nil
		}

		if _, err := s.conn.SendOne(packet, addr); err != nil {
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
	missCh, ok := s.misses[vnid]
	if !ok {
		return nil, fmt.Errorf("unknown vnid %d", vnid)
	}

	return missCh, nil
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
