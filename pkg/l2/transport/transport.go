package transport

import (
	"net"

	"github.com/vishvananda/netlink"
)

//Transport are drivers for handling lower level packet actions
type Transport interface {
	//Start begins the listener
	Start() error

	//SetMTU sets the default MTU of added interfaces
	SetMTU(mtu int32) error

	//AddEP adds a new VNID endpoint device to a given bridge
	AddEP(vnid uint32, br *netlink.Bridge) error

	//DelEP removes a  VNID endpoint device
	DelEP(vnid uint32) error

	//Status gives the device status of the VNID endpoint device
	Status(vnid uint32) EPStatus

	//AddVLAN adds a vlan to the trunking device
	AddVLAN(vnid uint32, vlan uint16) error

	//DelVLAN removes a vlan from the trunking device
	DelVLAN(vnid uint32, vlan uint16) error

	//AddForwardEntry adds a forwarding entry for the trunking devices to relay packets to other endpoints
	AddForwardEntry(vnid uint32, mac net.HardwareAddr, ip net.IP) error

	//ForwardingMiss provides a readonly channel to react to missing forwarding entries
	ForwardingMiss(vnid uint32) (<-chan ForwardingMiss, error)
}

//ForwardingMiss contains info for missed forwarding info
type ForwardingMiss struct {
	Type   ForwardingMissType
	HwAddr net.HardwareAddr
	IP     net.IP
}

//ForwardingMissType type of forwarding miss
type ForwardingMissType int

const (
	//MissTypeEP L2 Miss or FDB Miss
	MissTypeEP ForwardingMissType = iota

	//MissTypeARP L3 Miss or ARP proxy
	MissTypeARP
)

func (m ForwardingMissType) String() string {
	return [...]string{"End Point", "ARP"}[m]
}

//EPStatus provides finite state of an endpoint
type EPStatus int

const (
	//EPStatusUP is available and live
	EPStatusUP EPStatus = iota

	//EPStatusDOWN is available but administratively down
	EPStatusDOWN

	//EPStatusMISSING is missing
	EPStatusMISSING
)
