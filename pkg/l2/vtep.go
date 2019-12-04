package l2

import (
	"fmt"
	"log"
	"net"
	"time"

	"github.com/vishvananda/netlink"
)

const (
	vtepPattern = "t-%d"
)

//GetVTEP finds the vtep for the given VPC
func GetVTEP(vpcID int32) (netlink.Link, error) {
	handle, err := netlink.NewHandle(netlink.FAMILY_ALL)
	if err != nil {
		return nil, err
	}
	defer func() {
		handle.Delete()
	}()

	links, err := handle.LinkList()
	if err != nil {
		return nil, err
	}

	ifaceName := fmt.Sprintf(vtepPattern, vpcID)

	for _, link := range links {
		if link.Type() == "tuntap" && link.Attrs().Name == ifaceName {
			return link, nil
		}
	}
	return nil, nil
}

//HasVTEP checks if the vtep exists by getting it
func HasVTEP(vpcID int32) (bool, error) {
	vtep, err := GetVTEP(vpcID)
	if err != nil {
		return false, err
	}

	br, err := GetVPCBridge(vpcID)
	if err != nil {
		return false, err
	}

	if br != nil && vtep != nil && vtep.Attrs().MasterIndex != br.Attrs().Index {
		return false, fmt.Errorf("VTEP is not mastered by VPC Bridge")
	}

	return vtep != nil, nil
}

//CreateVTEP creates a new vtep using the linux VxLAN device
//in nolearning mode (use MP-BGP-EVPN/L2VPN for learning)
//NOTE: the actual VxLAN traffic port is used insteado the linux default
func CreateVTEP(vpcID int32, bridge *netlink.Bridge, dev string) (netlink.Link, error) {
	if ok, _ := HasVTEP(vpcID); ok {
		return nil, fmt.Errorf("vpc %d already has a vtep", vpcID)
	}

	/* MACVTAP */
	// vtep := &netlink.Macvtap{
	// 	Macvlan: netlink.Macvlan{
	// 		LinkAttrs: netlink.LinkAttrs{
	// 			Name:        fmt.Sprintf(vtepPattern, vpcID),
	// 			MTU:         1000,
	// 			ParentIndex: bridge.Index,
	// 		},
	// 		Mode: netlink.MACVLAN_MODE_BRIDGE,
	// 	},
	// }

	vtep := &netlink.Tuntap{
		Mode: netlink.TUNTAP_MODE_TAP,
		LinkAttrs: netlink.LinkAttrs{
			Name:        fmt.Sprintf(vtepPattern, vpcID),
			MTU:         1000,
			MasterIndex: bridge.Index,
		},
		Flags: netlink.TUNTAP_MULTI_QUEUE_DEFAULTS,
	}

	if err := netlink.LinkAdd(vtep); err != nil {
		return nil, fmt.Errorf("failed to add link to netlink: %s", err)
	}

	err := netlink.LinkSetUp(vtep)

	return vtep, err
}

//DeleteVTEP delets the linux vxlan device
func DeleteVTEP(vpcID int32) error {
	if ok, _ := HasVTEP(vpcID); !ok {
		return fmt.Errorf("vpc %d vxlan tunnel does not exist", vpcID)
	}

	vtep, err := GetVTEP(vpcID)
	if err != nil {
		return err
	}

	return netlink.LinkDel(vtep)
}

//SubscribeL2Updates monitors for subscription errors
func (s *Server) SubscribeL2Updates(vpcID int32) {
	miss, err := s.transport.FDBMiss(uint32(vpcID))
	if err != nil {
		log.Fatalf("uninited vtep miss")
	}

	ticker := time.Tick(10 * time.Second)
	for {
		select {
		case <-ticker:
			go s.handleVTEPPoll(uint32(vpcID))
			break
		case mac, ok := <-miss:
			if !ok {
				return
			}
			go s.handleMacLookup(uint32(vpcID), mac)
			break
		}
	}
}

//handleVTEPPoll applies zero'd hwaddrs to bridge forwarding
func (s *Server) handleVTEPPoll(vpcID uint32) {
	endpoints, err := s.bgp.BroadcastEndpoints(vpcID)
	if err != nil {
		log.Printf("VTEP linking failed: %s", err)
	}

	for _, endpoint := range endpoints {
		hwaddr, _ := net.ParseMAC("00:00:00:00:00:00")
		s.transport.FDB.AddEntry(vpcID, hwaddr, endpoint)
	}
}

//handleIPLookup looks up a potential mac for a given IP and adds to FDB
func (s *Server) handleIPLookup(vpcID uint32, vlan uint16, ip net.IP) {
	hwaddr, gw, err := s.bgp.LookupIP(vpcID, vlan, ip)
	if err != nil || ip == nil {
		log.Printf("Failed to find VTEP for %+v %s\n", ip, err)
		return
	}

	log.Println("Found HWADDR for", ip, "at", hwaddr, "via", gw)
}

func (s *Server) handleMacLookup(vpcID uint32, mac net.HardwareAddr) {
	gw, err := s.bgp.LookupMac(vpcID, mac)
	if err != nil || gw == nil {
		log.Printf("Failed to find VTEP for %s", mac)
		return
	}

	s.transport.FDB.AddEntry(vpcID, mac, gw)
}
