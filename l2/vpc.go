package l2

import "github.com/vishvananda/netlink"

const (
	vtepDev = "wlp0s20f3"
)

type Stack struct {
	VPCID  int32
	Bridge *netlink.Bridge
	Vtep   *netlink.Vxlan
}

func CreateVPCStack(vpcID int32) (*Stack, error) {
	br, err := CreateVPCBridge(vpcID)
	if err != nil {
		return nil, err
	}

	vtep, err := CreateVTEP(vpcID, br, vtepDev)
	if err != nil {
		return nil, err
	}

	return &Stack{VPCID: vpcID, Bridge: br, Vtep: vtep}, nil
}

func GetVPCStack(vpcID int32) (*Stack, error) {
	stack := &Stack{VPCID: vpcID}

	br, err := GetVPCBridge(vpcID)
	if err != nil {
		return nil, err
	}
	stack.Bridge = br.(*netlink.Bridge)

	vtep, err := GetVTEP(vpcID)
	if err != nil {
		return nil, err
	}
	stack.Vtep = vtep.(*netlink.Vxlan)

	return stack, nil
}

func DeleteVPCStack(stack *Stack) error {
	if err := DeleteVTEP(stack.VPCID); err != nil {
		return err
	}

	if err := DeleteVPCBridge(stack.VPCID); err != nil {
		return err
	}

	return nil
}
