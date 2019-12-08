package l2

import (
	"fmt"
	"net"

	"github.com/tcfw/vpc/pkg/l2/transport"
	"github.com/vishvananda/netlink"
	"golang.org/x/sys/unix"
)

const (
	vtepPattern = "t-%d"
)

//VTEPTransport adds support for VTEP devices via linux kernel
type VTEPTransport struct {
	mtu      int32
	vlans    map[uint16]int
	ifaces   map[uint32]netlink.Link
	pubIface netlink.Link
}

//NewVTEPTransport inits the transport
func NewVTEPTransport(pubIface netlink.Link) *VTEPTransport {
	return &VTEPTransport{
		mtu:      1500,
		vlans:    map[uint16]int{},
		ifaces:   map[uint32]netlink.Link{},
		pubIface: pubIface,
	}
}

//Start begins the listener
func (v *VTEPTransport) Start() error {
	return nil
}

//SetMTU sets the default MTU of added vtep devices
func (v *VTEPTransport) SetMTU(mtu int32) error {
	v.mtu = mtu

	for _, iface := range v.ifaces {
		netlink.LinkSetMTU(iface, int(v.mtu))
	}

	return nil
}

//AddEP adds a new VTEP to a given bridge
func (v *VTEPTransport) AddEP(vnid uint32, br *netlink.Bridge) error {
	vtep := &netlink.Vxlan{
		LinkAttrs: netlink.LinkAttrs{
			Name:        fmt.Sprintf(vtepPattern, vnid),
			MTU:         int(v.mtu),
			MasterIndex: br.Index,
		},
		VxlanId:      int(vnid),
		VtepDevIndex: v.pubIface.Attrs().Index,
		L2miss:       true,
		L3miss:       true,
		Learning:     false,
		Proxy:        true,
		Port:         4789,
	}

	if err := netlink.LinkAdd(vtep); err != nil {
		return fmt.Errorf("failed to add link to netlink: %s", err)
	}

	if err := netlink.LinkSetUp(vtep); err != nil {
		return fmt.Errorf("failed to bring up vtep: %s", err)
	}

	v.ifaces[vnid] = vtep

	return nil
}

//DelEP removes the VTEP device
func (v *VTEPTransport) DelEP(vnid uint32) error {
	if _, ok := v.ifaces[vnid]; !ok {
		return fmt.Errorf("vtep not active")
	}

	return netlink.LinkDel(v.ifaces[vnid])
}

//Status gives the device status of VTEP
func (v *VTEPTransport) Status(vnid uint32) transport.EPStatus {
	if _, ok := v.ifaces[vnid]; !ok {
		return transport.EPStatusMISSING
	}

	var status transport.EPStatus

	switch v.ifaces[vnid].Attrs().OperState {
	case netlink.OperNotPresent:
		status = transport.EPStatusMISSING
		break
	case netlink.OperDown:
		status = transport.EPStatusDOWN
		break
	case netlink.OperUp:
		status = transport.EPStatusUP
		break
	}

	return status
}

//AddVLAN adds a vlan to the VTEP for trunking
func (v *VTEPTransport) AddVLAN(vnid uint32, vlan uint16) error {
	_, ok := v.vlans[vlan]
	if !ok {
		v.vlans[vlan] = 1
	} else {
		v.vlans[vlan]++
	}

	return netlink.BridgeVlanAdd(v.ifaces[vnid], vlan, false, false, false, false)
}

//DelVLAN removes a vlan from the VTEP for trunking
func (v *VTEPTransport) DelVLAN(vnid uint32, vlan uint16) error {
	_, ok := v.vlans[vlan]
	if !ok {
		return fmt.Errorf("vlan already removed")
	}

	v.vlans[vlan]--

	if v.vlans[vlan] <= 0 {
		return netlink.BridgeVlanDel(v.ifaces[vnid], vlan, false, false, false, false)
	}

	return nil
}

//AddForwardEntry adds a forwarding entry to the VTEPs FDB
func (v *VTEPTransport) AddForwardEntry(vnid uint32, mac net.HardwareAddr, ip net.IP) error {
	iface, ok := v.ifaces[vnid]
	if !ok {
		return fmt.Errorf("vtep missing")
	}

	neigh := &netlink.Neigh{
		LinkIndex:    iface.Attrs().Index,
		HardwareAddr: mac,
		IP:           ip,
		Family:       unix.AF_BRIDGE,
		State:        netlink.NUD_REACHABLE,
		Flags:        netlink.NTF_SELF | netlink.NTF_PROXY,
	}

	return netlink.NeighAppend(neigh)
}

//ForwardingMiss provides a readonly channel to react to missed FDB entries
func (v *VTEPTransport) ForwardingMiss(vnid uint32) (<-chan transport.ForwardingMiss, error) {
	updatesCh := make(chan netlink.NeighUpdate)
	doneCh := make(chan struct{})

	forwardMiss := make(chan transport.ForwardingMiss)

	if err := netlink.NeighSubscribe(updatesCh, doneCh); err != nil {
		return nil, err
	}

	go func() {
		defer close(doneCh)

		for {
			change, ok := <-updatesCh
			if !ok {
				goto close
			}

			iface, ok := v.ifaces[vnid]
			if !ok {
				goto close
			}

			//Ignore updates from other interfaces
			if change.LinkIndex != iface.Attrs().Index {
				continue
			}

			if miss := v.filterNeighUpdate(change); miss != nil {
				forwardMiss <- *miss
			}
		}
	close:
		close(forwardMiss)
		return
	}()

	return forwardMiss, nil
}

//filterNeighUpdate filters the neighbor updates or FDB misses from the VTEP if they should be processed up stream via the SDN
func (v *VTEPTransport) filterNeighUpdate(change netlink.NeighUpdate) *transport.ForwardingMiss {
	if change.State != netlink.NUD_STALE {
		return nil
	}

	miss := &transport.ForwardingMiss{}

	if change.IP == nil {
		//l2 miss
		miss.Type = transport.MissTypeEP
		miss.HwAddr = change.HardwareAddr
	} else if change.HardwareAddr == nil {
		//l3 miss
		miss.Type = transport.MissTypeARP
		miss.IP = change.IP

	}

	return miss
}
