package tap

import (
	"fmt"

	"github.com/vishvananda/netlink"
)

//AddVLAN adds a VLAN to the vtep on the bridge
func (s *Listener) AddVLAN(vnid uint32, vlan uint16) error {
	_, ok := s.vlans[vlan]
	if !ok {
		s.vlans[vlan] = 1
	} else {
		s.vlans[vlan]++
	}

	// br, err := netlink.LinkByName(fmt.Sprintf("b-%d", vnid))
	// if err != nil {
	// 	return err
	// }

	// return netlink.BridgeVlanAdd(s.taps[vnid].IFace(), vlan, false, false, false, false)
	// return netlink.BridgeVlanAdd(br, vlan, false, false, false, false)
	return nil
}

//DelVLAN removes the VLAN to the vtep on the bridge
func (s *Listener) DelVLAN(vnid uint32, vlan uint16) error {
	_, ok := s.vlans[vlan]
	if !ok {
		return fmt.Errorf("vlan already removed")
	}

	s.vlans[vlan]--

	br, err := netlink.LinkByName(fmt.Sprintf("b-%d", vnid))
	if err != nil {
		return err
	}

	if s.vlans[vlan] <= 0 {
		// return netlink.BridgeVlanDel(s.taps[vnid].IFace(), vlan, false, false, false, false)
		return netlink.BridgeVlanDel(br, vlan, false, false, false, false)
	}

	return nil
}
