package l2

import (
	"fmt"

	"github.com/vishvananda/netlink"
)

const (
	vpcBridgePattern = "br-vpc-%d"
)

func GetVPCBridge(vpcID int32) (netlink.Link, error) {
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
		if link.Type() == "bridge" && link.Attrs().Name == fmt.Sprintf(vpcBridgePattern, vpcID) {
			return link, nil
		}
	}
	return nil, nil
}

func HasVPCBridge(vpcID int32) (bool, error) {
	br, err := GetVPCBridge(vpcID)
	if err != nil {
		return false, err
	}
	return br != nil, nil
}

func CreateVPCBridge(vpcID int32) (*netlink.Bridge, error) {
	if ok, _ := HasVPCBridge(vpcID); ok {
		return nil, fmt.Errorf("vpc %d already has a bridge", vpcID)
	}

	la := netlink.NewLinkAttrs()
	la.Name = fmt.Sprintf(vpcBridgePattern, vpcID)
	br := &netlink.Bridge{LinkAttrs: la}

	if err := netlink.LinkAdd(br); err != nil {
		return nil, err
	}

	err := netlink.LinkSetUp(br)

	return br, err
}

func DeleteVPCBridge(vpcID int32) error {
	if ok, _ := HasVPCBridge(vpcID); !ok {
		return fmt.Errorf("vpc %d bridge does not exist", vpcID)
	}

	br, err := GetVPCBridge(vpcID)
	if err != nil {
		return err
	}

	return netlink.LinkDel(br)
}