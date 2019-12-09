package tap

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"sync"

	"github.com/tcfw/vpc/pkg/l2/transport"

	"github.com/vishvananda/netlink"
)

//Listener holds all vtep VNIs
type Listener struct {
	mu         sync.Mutex
	vteps      map[uint32]*Tap
	packetConn net.PacketConn
	mtu        int32
	tx         chan *Packet
	FDB        *FDB
	port       int
	vlans      map[uint16]int
}

//NewListener inits a new VTEP style listener
func NewListener(port int) (*Listener, error) {
	lis := &Listener{
		port:  port,
		vteps: map[uint32]*Tap{},
		mtu:   1500,
		tx:    make(chan *Packet, 1000),
		FDB:   NewFDB(),
		vlans: map[uint16]int{},
	}

	return lis, nil
}

//SetMTU sets the MTU of the listening device
func (s *Listener) SetMTU(mtu int32) error {
	if s.packetConn != nil {
		return fmt.Errorf("cannot set mtu after already started")
	}

	s.mtu = mtu
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

	s.vteps[vnid] = tapHandler

	log.Println("registered new VTEP")

	// go tapHandler.Handle()
	go tapHandler.HandlePCAP()

	return nil
}

//DelEP stops & deletes VTEP activity
func (s *Listener) DelEP(vnid uint32) error {
	s.vteps[vnid].Stop()

	netlink.LinkDel(s.vteps[vnid].iface)

	delete(s.vteps, vnid)

	return nil
}

//Start begins listening for UDP handling
func (s *Listener) Start() error {
	pc, err := net.ListenPacket("udp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		return err
	}

	s.packetConn = pc

	go s.handleOut()

	buf := make([]byte, s.mtu)
	for {
		n, addr, err := s.packetConn.ReadFrom(buf)
		if err != nil {
			return err
		}
		s.handleIn(addr, buf[:n])
	}
}

func (s *Listener) handleIn(addr net.Addr, raw []byte) {
	p, err := FromBytes(bytes.NewBuffer(raw))
	if err != nil {
		log.Printf("invalid packet: %s\n", err)
	}

	//Ignore unknown VNIDs
	vtep, ok := s.vteps[p.VNID]
	if !ok {
		return
	}

	vtep.in <- p
}

func (s *Listener) handleOut() {
	var buf bytes.Buffer

	for {
		buf.Reset()
		packet, ok := <-s.tx
		if !ok {
			return
		}

		//TODO(tcfw) handle ARP Proxy and ICMPv6

		dst := packet.InnerFrame.Destination()
		broadcast := net.HardwareAddr{255, 255, 255, 255, 255, 255}

		n, err := packet.WriteTo(&buf)
		if err != nil {
			log.Fatalf("Failed to write packet: %s", err)
		}

		if bytes.Compare(dst, broadcast) != 0 {
			addr := s.FDB.LookupMac(packet.VNID, dst)

			if addr == nil {
				s.vteps[packet.VNID].FDBMiss <- transport.ForwardingMiss{Type: transport.MissTypeEP, HwAddr: dst}
				continue
			}

			s.sendPacket(buf.Bytes()[:n], addr)
		} else {
			//Flood to all EPs
			for _, addr := range s.FDB.ListBroadcast(packet.VNID) {
				s.sendPacket(buf.Bytes()[:n], addr)
			}
		}
	}
}

//sendPacket forwards a data to a given endpoint
func (s *Listener) sendPacket(data []byte, ip net.IP) error {
	dst := &net.UDPAddr{IP: ip, Port: 4789, Zone: ""}
	_, err := s.packetConn.WriteTo(data, dst)
	return err
}

//ForwardingMiss gets a readonly sub for FDB misses
func (s *Listener) ForwardingMiss(vnid uint32) (<-chan transport.ForwardingMiss, error) {
	vtep, ok := s.vteps[vnid]
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
