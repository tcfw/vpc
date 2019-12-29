package controller

import (
	"net"
)

//Controller s allows storing and publishing of endpoints and IPs
type Controller interface {
	Start() error
	Stop()

	RegisterMacIP(vnid uint32, vlan uint32, mac net.HardwareAddr, ip net.IP) error
	DeregisterMacIP(vnid uint32, vlan uint32, mac net.HardwareAddr, ip net.IP) error
	LookupIP(vnid uint32, vlan uint16, ip net.IP) (net.HardwareAddr, net.IP, error)
	LookupMac(vnid uint32, mac net.HardwareAddr) (net.IP, error)

	BroadcastEndpoints(vnid uint32) ([]net.IP, error)
	RegisterEP(vnid uint32) error
	DeregisterEP(vnid uint32) error
}
