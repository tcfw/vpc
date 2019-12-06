package l2

import (
	"fmt"
	"io/ioutil"

	"github.com/vishvananda/netlink"
)

const (
	bridgePattern = "b-%d"
)

//getBridge finds the bridge associated with a VPC
func getBridge(vpcID int32) (netlink.Link, error) {
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

	ifaceName := fmt.Sprintf(bridgePattern, vpcID)

	for _, link := range links {
		if link.Type() == "bridge" && link.Attrs().Name == ifaceName {
			return link, nil
		}
	}
	return nil, nil
}

//hasBridge checks if a bridge exists by trying to get it
func hasBridge(vpcID int32) (bool, error) {
	br, err := getBridge(vpcID)
	if err != nil {
		return false, err
	}
	return br != nil, nil
}

//createBridge creates a new linux bridge for a VPC
func createBridge(vpcID int32) (*netlink.Bridge, error) {
	if ok, _ := hasBridge(vpcID); ok {
		return nil, fmt.Errorf("vpc %d already has a bridge", vpcID)
	}

	la := netlink.NewLinkAttrs()
	la.Name = fmt.Sprintf(bridgePattern, vpcID)
	la.MTU = 2000

	br := &netlink.Bridge{LinkAttrs: la}

	if err := netlink.LinkAdd(br); err != nil {
		return nil, err
	}

	if err := enableBridgeVlanFiltering(br.Name); err != nil {
		return nil, fmt.Errorf("Failed to enable VLAN filtering on bridge: %s", err)
	}

	err := netlink.LinkSetUp(br)

	return br, err
}

//enableBridgeVlanFiltering sets vlan filtering on the bridge
func enableBridgeVlanFiltering(bridgeName string) error {
	return ioutil.WriteFile(fmt.Sprintf("/sys/devices/virtual/net/%s/bridge/vlan_filtering", bridgeName), []byte("1"), 0644)
}

//deleteBridge deletes a linux bridge for a VPC
func deleteBridge(vpcID int32) error {
	if ok, _ := hasBridge(vpcID); !ok {
		return fmt.Errorf("vpc %d bridge does not exist", vpcID)
	}

	br, err := getBridge(vpcID)
	if err != nil {
		return err
	}

	return netlink.LinkDel(br)
}

func getBridgeLinks(brIndex int, vpcID int32) ([]netlink.Link, error) {
	slaveLinks := []netlink.Link{}

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

	vtepName := fmt.Sprintf("t-%d", vpcID)

	for _, link := range links {
		if link.Attrs().MasterIndex == brIndex && link.Attrs().Name != vtepName {
			slaveLinks = append(slaveLinks, link)
		}
	}

	return slaveLinks, nil
}
