package l2

import (
	"fmt"

	"github.com/vishvananda/netlink"
)

const (
	vtepPattern = "vtep-vpc-%d"
)

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

func HasVTEP(vpcID int32) (bool, error) {
	vtep, err := GetVTEP(vpcID)
	if err != nil {
		return false, err
	}
	return vtep != nil, nil
}

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
		Learning:     false,
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

func DeleteVTEP(vpcID int32) error {
	if ok, _ := HasVTEP(vpcID); !ok {
		return fmt.Errorf("vpc %d vxlan tunnel does not exist", vpcID)
	}

	br, err := GetVTEP(vpcID)
	if err != nil {
		return err
	}

	return netlink.LinkDel(br)
}
