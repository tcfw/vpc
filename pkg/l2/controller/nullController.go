package controller

import "net"

type NullController struct{}

func (n *NullController) Start() error {
	return nil
}

func (n *NullController) Stop() {
	return
}

func (n *NullController) RegisterMacIP(vnid uint32, vlan uint32, mac net.HardwareAddr, ip net.IP) error {
	return nil
}

func (n *NullController) DeregisterMacIP(vnid uint32, vlan uint32, mac net.HardwareAddr, ip net.IP) error {
	return nil
}

func (n *NullController) LookupIP(vnid uint32, vlan uint16, ip net.IP) (net.HardwareAddr, net.IP, error) {
	//return a test mac address
	return net.HardwareAddr{0x82, 0x2f, 0x00, 0x00, 0x00, 0xff}, ip, nil
}

func (n *NullController) LookupMac(vnid uint32, mac net.HardwareAddr) (net.IP, error) {
	return net.ParseIP("10.4.0.5"), nil
}

func (n *NullController) BroadcastEndpoints(vnid uint32) ([]net.IP, error) {
	return []net.IP{net.ParseIP("192.168.1.2")}, nil
}

func (n *NullController) RegisterEP(vnid uint32) error {
	return nil
}

func (n *NullController) DeregisterEP(vnid uint32) error {
	return nil
}
