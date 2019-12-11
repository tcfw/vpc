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

	return netlink.BridgeVlanAdd(s.taps[vnid].iface, vlan, false, false, false, false)
}

//DelVLAN removes the VLAN to the vtep on the bridge
func (s *Listener) DelVLAN(vnid uint32, vlan uint16) error {
	_, ok := s.vlans[vlan]
	if !ok {
		return fmt.Errorf("vlan already removed")
	}

	s.vlans[vlan]--

	if s.vlans[vlan] <= 0 {
		return netlink.BridgeVlanDel(s.taps[vnid].iface, vlan, false, false, false, false)
	}

	return nil
}
