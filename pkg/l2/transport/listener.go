package transport

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"sync"

	"github.com/songgao/packets/ethernet"
	"github.com/songgao/water"
)

//Listener holds all vtep VNIs
type Listener struct {
	mu         sync.Mutex
	vteps      map[uint32]*vtep
	packetConn net.PacketConn
	mtu        int
	tx         chan *Packet
	FDB        *FDB
}

//NewListener inits a new VTEP style listener
func NewListener(port uint32, mtu int) (*Listener, error) {
	pc, err := net.ListenPacket("udp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}

	lis := &Listener{
		packetConn: pc,
		vteps:      map[uint32]*vtep{},
		mtu:        mtu,
		tx:         make(chan *Packet),
		FDB:        NewFDB(),
	}

	return lis, nil
}

//AddVTEP adds a vtep handler via tun device
func (s *Listener) AddVTEP(vnid uint32, tun string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	vtep := &vtep{
		vnid:    vnid,
		out:     make(chan ethernet.Frame),
		in:      make(chan *Packet),
		lis:     s,
		mtu:     s.mtu,
		FDBMiss: make(chan net.HardwareAddr, 1024),
	}

	config := water.Config{
		DeviceType: water.TAP,
	}
	config.Name = tun
	ifce, err := water.New(config)
	if err != nil {
		return fmt.Errorf("failed to init tun dev: %s", err)
	}
	vtep.tuntap = ifce

	s.vteps[vnid] = vtep

	log.Println("registered new VTEP")

	go vtep.Handle()

	return nil
}

//DelVTEP stops & deletes VTEP activity
func (s *Listener) DelVTEP(vnid uint32) {
	s.vteps[vnid].Stop()
	delete(s.vteps, vnid)
}

//Start begins listening for UDP handling
func (s *Listener) Start() error {
	go s.handleOut()

	for {
		buf := make([]byte, s.mtu)
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

		addrs := []net.IP{}

		addr := s.FDB.LookupMac(packet.VNID, dst)

		if addr == nil {
			go func() {
				if dst.String() != "ff:ff:ff:ff:ff:ff" {
					s.vteps[packet.VNID].FDBMiss <- dst
				}
			}()
			addrs = s.FDB.ListBroadcast(packet.VNID)
		} else {
			addrs = append(addrs, addr)
		}

		for _, addr := range addrs {
			dst := &net.UDPAddr{IP: addr, Port: 4789, Zone: ""}

			_, err := s.packetConn.WriteTo(packet.Bytes(), dst)
			if err != nil {
				log.Printf("failed to send UDP: %s", err)
			}
		}
	}
}

//FDBMiss gets a readonly sub for FDB misses
func (s *Listener) FDBMiss(vnid uint32) (<-chan net.HardwareAddr, error) {
	vtep, ok := s.vteps[vnid]
	if !ok {
		return nil, fmt.Errorf("unknown vnid %d", vnid)
	}

	return vtep.FDBMiss, nil
}
