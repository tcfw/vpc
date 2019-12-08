package controller

import (
	"net"
)

//Controller s allows storing and publishing of endpoints and IPs
type Controller interface {
	Start() error
	Stop()

	RegisterMacIP(uint32, uint32, net.HardwareAddr, net.IP) error
	DeregisterMacIP(uint32, uint32, net.HardwareAddr, net.IP) error
	LookupIP(uint32, uint16, net.IP) (net.HardwareAddr, net.IP, error)
	LookupMac(uint32, net.HardwareAddr) (net.IP, error)

	BroadcastEndpoints(uint32) ([]net.IP, error)
	RegisterEP(uint32) error
	DeregisterEP(uint32) error
}
