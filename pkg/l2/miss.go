package l2

import (
	"log"
	"net"
	"time"

	"github.com/tcfw/vpc/pkg/l2/transport"
)

//HandleMisses monitors for subscription errors
func (s *Server) HandleMisses(vpcID uint32, miss <-chan transport.ForwardingMiss) {
	ticker := time.Tick(10 * time.Second)
	for {
		select {
		case <-ticker:
			go s.handleEPPoll(uint32(vpcID))
			break
		case miss, ok := <-miss:
			if !ok {
				return
			}
			go s.handleMacLookup(uint32(vpcID), miss.HwAddr)
			break
		}
	}
}

//handleEPPoll applies zero'd hwaddrs to bridge forwarding
func (s *Server) handleEPPoll(vpcID uint32) {
	endpoints, err := s.sdn.BroadcastEndpoints(vpcID)
	if err != nil {
		log.Printf("VTEP linking failed: %s", err)
	}

	for _, endpoint := range endpoints {
		hwaddr, _ := net.ParseMAC("00:00:00:00:00:00")
		s.transport.AddForwardEntry(vpcID, hwaddr, endpoint)
	}
}

//handleMacLookup looks up mac addresses from FDB misses and adds them back into the FDB
func (s *Server) handleMacLookup(vpcID uint32, mac net.HardwareAddr) {
	gw, err := s.sdn.LookupMac(vpcID, mac)
	if err != nil || gw == nil {
		log.Printf("Failed to find VTEP for %s", mac)
		return
	}

	s.transport.AddForwardEntry(vpcID, mac, gw)
}
