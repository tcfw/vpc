package l2

import (
	"fmt"

	"github.com/vishvananda/netlink"
)

//Stack stores references to the various devices for VPC connectivity
type Stack struct {
	VPCID  int32
	Bridge *netlink.Bridge
	Vtep   netlink.Link
	Nics   map[string]*VNic
}

//CreateVPCStack creates a linux bridge and vtep to construct the VPC
func CreateVPCStack(vpcID int32, vtepDev string) (*Stack, error) {
	br, err := CreateVPCBridge(vpcID)
	if err != nil {
		DeleteVPCBridge(vpcID)
		return nil, fmt.Errorf("failed to create bridge: %s", err)
	}

	vtep, err := CreateVTEP(vpcID, br, vtepDev)
	if err != nil {
		DeleteVPCBridge(vpcID)
		DeleteVTEP(vpcID)
		return nil, fmt.Errorf("failed to create vtep: %s", err)
	}

	return &Stack{VPCID: vpcID, Bridge: br, Vtep: vtep, Nics: map[string]*VNic{}}, nil
}

//GetVPCStack finds the linux bridge and vtep assocated with the VPC id
func GetVPCStack(vpcID int32) (*Stack, error) {
	stack := &Stack{VPCID: vpcID, Nics: map[string]*VNic{}}

	br, err := GetVPCBridge(vpcID)
	if err != nil {
		return nil, err
	}
	if br != nil {
		stack.Bridge = br.(*netlink.Bridge)
	}

	vtep, err := GetVTEP(vpcID)
	if err != nil {
		return nil, err
	}
	if vtep != nil {
		stack.Vtep = vtep.(*netlink.Vxlan)
	}

	return stack, nil
}

//DeleteVPCStack deletes both the linux bridge and vtep
func DeleteVPCStack(stack *Stack) error {
	if err := DeleteVTEP(stack.VPCID); err != nil {
		return err
	}

	if err := DeleteVPCBridge(stack.VPCID); err != nil {
		return err
	}

	return nil
}
