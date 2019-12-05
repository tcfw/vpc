package l2

import (
	"fmt"

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

	// vtep := &netlink.Tuntap{
	// 	Mode: netlink.TUNTAP_MODE_TAP,
	// 	LinkAttrs: netlink.LinkAttrs{
	// 		Name:        fmt.Sprintf(vtepPattern, vpcID),
	// 		MTU:         1000,
	// 		MasterIndex: bridge.Index,
	// 	},
	// 	Flags: netlink.TUNTAP_MULTI_QUEUE_DEFAULTS,
	// }

	devLink, err := netlink.LinkByName(dev)
	if err != nil {
		return nil, fmt.Errorf("Cannot find vtep link dev %s", dev)
	}

	vtep := &netlink.Vxlan{
		LinkAttrs: netlink.LinkAttrs{
			Name:        fmt.Sprintf(vtepPattern, vpcID),
			MTU:         1000,
			MasterIndex: bridge.Index,
		},
		VxlanId:      int(vpcID),
		VtepDevIndex: devLink.Attrs().Index,
		L2miss:       true,  //TODO(tcfw) inject l2 req to FDB
		L3miss:       false, //TODO(tcfw) track l3 and response to neighbor reqs
		Learning:     false,
		Proxy:        false,
		Port:         4789,
	}

	if err := netlink.LinkAdd(vtep); err != nil {
		return nil, fmt.Errorf("failed to add link to netlink: %s", err)
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
