package l2

import (
	"fmt"

	"github.com/vishvananda/netlink"
)

const (
	nicPattern = "n-%s"
)

//CreateNIC creates a new tap device attached to a VPC bridge
//NOTE: this type of NIC can really only be suitable for containers
//TODO(tcfw) create new functio to generate macvtap
func CreateNIC(stack *Stack, id string) (netlink.Link, error) {
	la := netlink.NewLinkAttrs()
	la.Name = fmt.Sprintf(nicPattern, id)
	nic := &netlink.Tuntap{
		LinkAttrs: la,
		Mode:      netlink.TUNTAP_MODE_TAP,
	}
	if err := netlink.LinkAdd(nic); err != nil {
		return nil, fmt.Errorf("Failed to add tap device: %s", err)
	}

	if err := netlink.LinkSetMaster(nic, stack.Bridge); err != nil {
		return nil, fmt.Errorf("Failed to set master: %s", err)
	}

	err := netlink.LinkSetUp(nic)

	return nic, err
}

//GetNIC finds a tap interface given a stack and expected id
func GetNIC(stack *Stack, id string) (netlink.Link, error) {
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
		if link.Attrs().Name == fmt.Sprintf(nicPattern, id) {
			return link, nil
		}
	}
	return nil, nil
}

//HasNIC checks if a nic exists by trying to get it
func HasNIC(stack *Stack, id string) (bool, error) {
	nic, err := GetNIC(stack, id)
	if err != nil {
		return false, err
	}
	return nic != nil, nil
}

//DeleteNIC deletes the tap from
func DeleteNIC(stack *Stack, id string) error {
	if ok, _ := HasNIC(stack, id); !ok {
		return fmt.Errorf("nic %s does not exist", id)
	}

	nic, err := GetNIC(stack, id)
	if err != nil {
		return err
	}

	return netlink.LinkDel(nic)
}
