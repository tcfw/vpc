package l2

import (
	"fmt"

	"github.com/vishvananda/netlink"
)

//Stack stores references to the various devices for VPC connectivity
type Stack struct {
	VPCID  int32
	Bridge *netlink.Bridge
	Nics   map[string]*VNic
}

//createStack creates a linux bridge and vtep to construct the VPC
func createStack(vpcID int32) (*Stack, error) {
	br, err := createBridge(vpcID)
	if err != nil {
		deleteBridge(vpcID)
		return nil, fmt.Errorf("failed to create bridge: %s", err)
	}

	return &Stack{VPCID: vpcID, Bridge: br, Nics: map[string]*VNic{}}, nil
}

//getStack finds the linux bridge and vtep assocated with the VPC id
func getStack(vpcID int32) (*Stack, error) {
	stack := &Stack{VPCID: vpcID, Nics: map[string]*VNic{}}

	br, err := getBridge(vpcID)
	if err != nil {
		return nil, err
	}
	if br != nil {
		stack.Bridge = br.(*netlink.Bridge)
	}

	return stack, nil
}

//deleteStack deletes both the linux bridge and vtep
func deleteStack(stack *Stack) error {
	if err := deleteBridge(stack.VPCID); err != nil {
		return err
	}

	return nil
}
