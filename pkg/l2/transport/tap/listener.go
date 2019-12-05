package tap

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"sync"

	"github.com/tcfw/vpc/pkg/l2/transport"

	"github.com/songgao/packets/ethernet"
	"github.com/songgao/water"
	"github.com/vishvananda/netlink"
)

const (
	vtepPattern = "t-%d"
)

//Listener holds all vtep VNIs
type Listener struct {
	mu         sync.Mutex
	vteps      map[uint32]*tap
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
		vteps: map[uint32]*tap{},
		mtu:   1500,
		tx:    make(chan *Packet),
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

	// handler, err := pcap.OpenLive(tun, int32(s.mtu), true, pcap.BlockForever)
	// if err != nil {
	// 	return err
	// }

	tapIface := &netlink.Tuntap{
		Mode: netlink.TUNTAP_MODE_TAP,
		LinkAttrs: netlink.LinkAttrs{
			Name:        fmt.Sprintf(vtepPattern, vnid),
			MTU:         int(s.mtu),
			MasterIndex: br.Index,
		},
		Flags: netlink.TUNTAP_MULTI_QUEUE_DEFAULTS,
	}

	if err := netlink.LinkAdd(tapIface); err != nil {
		return err
	}

	if err := netlink.LinkSetUp(tapIface); err != nil {
		return err
	}

	tapHandler := &tap{
		vnid:    vnid,
		out:     make(chan ethernet.Frame),
		in:      make(chan *Packet),
		lis:     s,
		iface:   tapIface,
		mtu:     int(s.mtu),
		FDBMiss: make(chan transport.ForwardingMiss),
		// handler: handler,
	}

	config := water.Config{
		DeviceType: water.TAP,
	}
	config.Name = tapIface.Name
	config.MultiQueue = true

	ifce, err := water.New(config)
	if err != nil {
		return fmt.Errorf("failed to init tun dev: %s", err)
	}
	tapHandler.tuntap = ifce

	s.vteps[vnid] = tapHandler

	log.Println("registered new VTEP")

	go tapHandler.Handle()
	// go vtep.HandlePCAP()

	return nil
}

//DelEP stops & deletes VTEP activity
func (s *Listener) DelEP(vnid uint32) error {
	s.vteps[vnid].Stop()
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
	for {
		packet, ok := <-s.tx
		if !ok {
			return
		}

		//TODO(tcfw) handle ARP Proxy and ICMPv6

		dst := packet.InnerFrame.Destination()
		broadcast := net.HardwareAddr{255, 255, 255, 255, 255, 255}

		if bytes.Compare(dst, broadcast) != 0 {
			addr := s.FDB.LookupMac(packet.VNID, dst)

			if addr == nil {
				s.vteps[packet.VNID].FDBMiss <- transport.ForwardingMiss{Type: transport.MissTypeEP, HwAddr: dst}
				continue
			}

			s.sendPacket(packet, addr)
		} else {
			//Flood to all VTEPs
			for _, addr := range s.FDB.ListBroadcast(packet.VNID) {
				s.sendPacket(packet, addr)
			}
		}
	}
}

func (s *Listener) sendPacket(packet *Packet, ip net.IP) error {
	dst := &net.UDPAddr{IP: ip, Port: 4789, Zone: ""}
	_, err := s.packetConn.WriteTo(packet.Bytes(), dst)
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
