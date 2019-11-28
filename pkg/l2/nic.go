package l2

import (
	"fmt"

	"github.com/vishvananda/netlink"
)

const (
	nicPattern = "n-%s"
)

//VNic holds the state of known NICs
type VNic struct {
	id     string
	vlan   uint16
	link   netlink.Link
	manual bool
}

//CreateNIC creates a new tap device attached to a VPC bridge
//NOTE: this type of NIC can really only be suitable for containers
//TODO(tcfw) create new functio to generate macvtap
func CreateNIC(stack *Stack, id string, subnetVlan uint16) (netlink.Link, error) {
	if subnetVlan > 4096 {
		return nil, fmt.Errorf("subnet out of range")
	}

	la := netlink.NewLinkAttrs()
	la.Name = fmt.Sprintf(nicPattern, id)
	la.MTU = 1000
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

	if err := netlink.BridgeVlanAdd(nic, subnetVlan, true, true, false, false); err != nil {
		return nil, fmt.Errorf("Failed to add VLAN to nic: %s", err)
	}

	if err := netlink.LinkSetUp(nic); err != nil {
		return nic, err
	}

	err := UpdateVTEPVlans(stack)

	stack.Nics[id] = &VNic{id: id, vlan: subnetVlan, link: nic}

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

	if _, ok := stack.Nics[id]; ok {
		delete(stack.Nics, id)
	}

	if err := netlink.LinkDel(nic); err != nil {
		return err
	}

	return UpdateVTEPVlans(stack)
}
