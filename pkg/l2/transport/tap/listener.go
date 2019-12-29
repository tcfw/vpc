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

	// "github.com/tcfw/vpc/pkg/l2/transport/tap/protocol/vxlan"
	"github.com/tcfw/vpc/pkg/l2/transport/tap/protocol/quic"

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
	tx    chan *protocol.Packet
	FDB   *FDB
	sdn   controller.Controller
	vlans map[uint16]int
}

//NewListener inits a new VTEP style listener
func NewListener() (*Listener, error) {
	lis := &Listener{
		taps:  map[uint32]*Tap{},
		mtu:   1500,
		tx:    make(chan *protocol.Packet, 1000),
		FDB:   NewFDB(),
		vlans: map[uint16]int{},
		conn:  quic.NewHandler(),
	}

	return lis, nil
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

//Start begins listening for UDP handling
func (s *Listener) Start() error {
	if err := s.conn.Start(); err != nil {
		return err
	}

	go s.handleOut()

	for {
		packet, err := s.conn.Recv()
		if err != nil {
			return err
		}
		s.handleIn(packet)
	}
}

func (s *Listener) handleIn(p *protocol.Packet) {
	//Ignore unknown VNIDs
	tap, ok := s.taps[p.VNID]
	if !ok {
		return
	}

	tap.in <- p
}

func (s *Listener) handleOut() {
	for {
		packet, ok := <-s.tx
		if !ok {
			return
		}

		var frame ethernet.Frame = packet.Frame

		//Ignore all non-tagged frames
		if frame.Tagging() == ethernet.NotTagged {
			continue
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
			continue
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
				continue
			}
		}

		broadcast := net.HardwareAddr{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}

		if bytes.Compare(dst, broadcast) != 0 {
			addr := s.FDB.LookupMac(packet.VNID, dst)

			if addr == nil {
				s.taps[packet.VNID].FDBMiss <- transport.ForwardingMiss{Type: transport.MissTypeEP, HwAddr: dst}
				continue
			}

			if _, err := s.conn.Send(packet, addr); err != nil {
				log.Printf("Error sending: %s\n", err)
				return
			}
		} else {
			//Flood to all EPs
			for _, addr := range s.FDB.ListBroadcast(packet.VNID) {
				if _, err := s.conn.Send(packet, addr); err != nil {
					log.Printf("Error sending: %s\n", err)
					return
				}
			}
		}
	}
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
