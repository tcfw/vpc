package controller

import "net"

//NullController an empty controller
type NullController struct{}

//Start should start the controller
func (n *NullController) Start() error {
	return nil
}

//Stop should stop the controller
func (n *NullController) Stop() {
	return
}

//RegisterMacIP usually would register the mac ip address combo
func (n *NullController) RegisterMacIP(vnid uint32, vlan uint32, mac net.HardwareAddr, ip net.IP) error {
	return nil
}

//DeregisterMacIP usually would remove a mac/ip combo
func (n *NullController) DeregisterMacIP(vnid uint32, vlan uint32, mac net.HardwareAddr, ip net.IP) error {
	return nil
}

//LookupIP usually looks up a mac address for a given IP.
//Hard coded to return 82:2f:00:00:00:ff
func (n *NullController) LookupIP(vnid uint32, vlan uint16, ip net.IP) (net.HardwareAddr, net.IP, error) {
	//return a test mac address
	return net.HardwareAddr{0x82, 0x2f, 0x00, 0x00, 0x00, 0xff}, ip, nil
}

//LookupMac usually would look up an IP for a given mac
//Hard coded to return net.IP{10.4.0.5}
func (n *NullController) LookupMac(vnid uint32, mac net.HardwareAddr) (net.IP, error) {
	return net.ParseIP("10.4.0.5"), nil
}

//BroadcastEndpoints usually would return a list of VTEP endpoints for a given VNID
//Hard coded to return []net.IP{192.168.1.2}
func (n *NullController) BroadcastEndpoints(vnid uint32) ([]net.IP, error) {
	return []net.IP{net.ParseIP("192.168.1.2")}, nil
}

//RegisterEP usually register a VTEP
func (n *NullController) RegisterEP(vnid uint32) error {
	return nil
}

//DeregisterEP usually remove a given VTEP
func (n *NullController) DeregisterEP(vnid uint32) error {
	return nil
}
