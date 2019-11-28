package l2

import (
	"fmt"
	"log"
	"net"
	"time"

	"golang.org/x/sys/unix"

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
	for _, link := range links {
		if link.Type() == "vxlan" && link.Attrs().Name == fmt.Sprintf(vtepPattern, vpcID) {
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
func CreateVTEP(vpcID int32, bridge *netlink.Bridge, dev string) (*netlink.Vxlan, error) {
	if ok, _ := HasVTEP(vpcID); ok {
		return nil, fmt.Errorf("vpc %d already has a vtep", vpcID)
	}

	devLink, err := netlink.LinkByName(dev)
	if err != nil {
		return nil, fmt.Errorf("Cannot find vtep link dev %s", dev)
	}

	la := netlink.NewLinkAttrs()
	la.Name = fmt.Sprintf(vtepPattern, vpcID)

	vtep := &netlink.Vxlan{
		LinkAttrs:    la,
		VxlanId:      int(vpcID),
		VtepDevIndex: devLink.Attrs().Index,
		L2miss:       false, //TODO(tcfw) inject l2 req to FDB
		L3miss:       false, //TODO(tcfw) track l3 and response to neighbor reqs
		Learning:     false,
		Proxy:        false,
		Port:         4789,
	}

	if err := netlink.LinkAdd(vtep); err != nil {
		return nil, err
	}

	if err := netlink.LinkSetMaster(vtep, bridge); err != nil {
		netlink.LinkDel(vtep)
		return nil, err
	}

	err = netlink.LinkSetUp(vtep)

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
	vtepLink, err := GetVTEP(vpcID)
	if err != nil {
		return
	}

	vtep := vtepLink.(*netlink.Vxlan)

	go func() {
		ticker := time.Tick(10 * time.Second)
		for {
			select {
			case <-ticker:
				s.handleVTEPPoll(uint32(vpcID))
				break
			}
		}
	}()

	if vtep.Proxy {
		updateCh := make(chan netlink.NeighUpdate)
		doneCh := make(chan struct{})

		if err := netlink.NeighSubscribe(updateCh, doneCh); err != nil {
			log.Println(err)
			return
		}

		for {
			change, ok := <-updateCh
			if !ok {
				break
			}
			if change.LinkIndex == vtep.Attrs().Index {
				//Only interested in neighbor requests
				if change.Type != unix.RTM_GETNEIGH {
					continue
				}

				log.Printf("%++v\n", change)

				if change.State == netlink.NUD_STALE && change.LLIPAddr == nil {
					if change.IP == nil && change.HardwareAddr.String() == net.HardwareAddr(nil).String() {
						s.handleVTEPPoll(uint32(vpcID))
					} else if change.IP == nil {
						s.handleMacReq(uint32(vpcID), change.HardwareAddr)
					}
				} else if change.State == netlink.NUD_FAILED || change.State == netlink.NUD_NOARP {
					log.Println("failed arp")
				} else {
					log.Println("unknown neigh state")
				}
			}
		}
	}
}

func (s *Server) handleVTEPPoll(vpcID uint32) {
	log.Printf("attempt to find vxlan vtep links")
	if err := s.bgp.ApplyVTEPAD(vpcID); err != nil {
		log.Printf("VTEP linking failed: %s", err)
	}
}

func (s *Server) handleMacReq(vpcID uint32, hwaddr net.HardwareAddr) {
	ip, err := s.bgp.LookupMac(uint32(vpcID), hwaddr)
	if err != nil || ip == nil {
		log.Printf("Failed to find VTEP for %+v %s\n", hwaddr, err)
		return
	}
}

//UpdateVLANTrunks reapplies required VLANs to VTEP for trunking
func (s *Server) UpdateVLANTrunks(stack *Stack) error {
	vlans := []uint16{}
	for _, nic := range stack.Nics {
		vlans = append(vlans, nic.vlan)

		if err := netlink.BridgeVlanAdd(stack.Vtep, nic.vlan, false, false, false, false); err != nil {
			return err
		}
	}

	//TODO(tcfw) clean unused vlans from trunk

	return nil
}
