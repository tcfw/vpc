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

func getNicVlans(index int32) uint16 {
	links, _ := netlink.BridgeVlanList()
	for linkIndex, linkVlans := range links {
		if linkIndex == index && len(linkVlans) > 0 {
			return linkVlans[0].Vid
		}
	}

	return 1
}

func enableBridgeVlanFiltering(bridgeName string) error {
	return ioutil.WriteFile(fmt.Sprintf("/sys/devices/virtual/net/%s/bridge/vlan_filtering", bridgeName), []byte("1"), 0644)
}
