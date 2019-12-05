package l2

import (
	"fmt"
	"io/ioutil"

	"github.com/vishvananda/netlink"
)

//UpdateVTEPVlans sets all required vlans on the vtep
func UpdateVTEPVlans(stack *Stack) error {
	vlans := listTrunkVlans(stack.Bridge)
	for _, vlan := range vlans {
		if err := netlink.BridgeVlanAdd(stack.Bridge, vlan, true, true, false, false); err != nil {
			return err
		}
	}

	return nil
}

//listTrunkVlans lists all vlans currently on the bridge
func listTrunkVlans(bridge *netlink.Bridge) []uint16 {
	vlans := []uint16{}
	links, _ := netlink.BridgeVlanList()

	for index, linkVlans := range links {
		link, _ := netlink.LinkByIndex(int(index))

		if link.Attrs().MasterIndex == bridge.Index {
			for _, vlan := range linkVlans {
				vlans = append(vlans, vlan.Vid)
			}
		}
	}

	return vlans
}

//clearNICVlans clears all vlan nics on the bridge
func clearNICVlans(link netlink.Link) error {
	links, _ := netlink.BridgeVlanList()
	for linkIndex, linkVlans := range links {
		if linkIndex == int32(link.Attrs().Index) && len(linkVlans) > 0 {
			for _, vlan := range linkVlans {
				netlink.BridgeVlanDel(link, vlan.Vid, false, false, false, false)
			}
		}
	}
	return nil
}

//UpdateVLANTrunks reapplies required VLANs to VTEP for trunking
func (s *Server) UpdateVLANTrunks(stack *Stack) error {
	if err := clearNICVlans(stack.Vtep); err != nil {
		return fmt.Errorf("failed to clear trunk VLANs: %s", err)
	}

	for _, nic := range stack.Nics {
		if err := netlink.BridgeVlanAdd(stack.Vtep, nic.vlan, false, false, false, false); err != nil {
			return err
		}
	}

	return nil
}
