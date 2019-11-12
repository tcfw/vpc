package l3

import (
	"fmt"

	l2API "github.com/tcfw/vpc/pkg/api/v1/l2"
	l2 "github.com/tcfw/vpc/pkg/l2"
	"github.com/vishvananda/netlink"
)

//L2APIToStack converts the l2-agent API stack into the more useful stack
func L2APIToStack(l2Stack *l2API.Stack) (*l2.Stack, error) {
	stack := &l2.Stack{
		VPCID: l2Stack.VpcID,
	}

	br, err := netlink.LinkByIndex(int(l2Stack.BridgeLinkIndex))
	if err != nil {
		return nil, fmt.Errorf("Failed to find bridge: %s", err)
	}

	stack.Bridge = br.(*netlink.Bridge)

	vtep, err := netlink.LinkByIndex(int(l2Stack.VtepLinkIndex))
	if err != nil {
		return nil, fmt.Errorf("Failed to find vtep: %s", err)
	}

	stack.Vtep = vtep.(*netlink.Vxlan)

	return stack, nil
}
