package transport

import (
	"net"

	"github.com/vishvananda/netlink"
)

//Transport are drivers for handling lower level packet actions
type Transport interface {
	Start() error
	SetMTU(mtu int32) error

	AddEP(vnid uint32, br *netlink.Bridge) error
	DelEP(vnid uint32) error
	Status(vnid uint32) EPStatus

	AddVLAN(vnid uint32, vlan uint16) error
	DelVLAN(vnid uint32, vlan uint16) error

	AddForwardEntry(vnid uint32, mac net.HardwareAddr, ip net.IP) error
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
