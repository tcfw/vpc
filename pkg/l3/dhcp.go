package l3

import (
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/apparentlymart/go-cidr/cidr"
	"github.com/insomniacslk/dhcp/dhcpv4"
	"github.com/insomniacslk/dhcp/dhcpv4/server4"
)

//RouterDHCP provides DHCP capabilities to routers
type RouterDHCP struct {
	mu sync.Mutex

	leaseTime  uint32
	ip         net.IP
	router     net.IP
	subnet     *net.IPNet
	dnsServers []net.IP
	iface      string
}

//NewDHCPv4Server consturcts a new DHCPv4 server
func NewDHCPv4Server(iface string, subnet *net.IPNet, dns []net.IP) (*RouterDHCP, error) {
	srv := &RouterDHCP{
		subnet:     subnet,
		dnsServers: dns,
		leaseTime:  uint32(2 * 7 * 24 * time.Hour / time.Second),
		iface:      iface,
	}

	DHCPIP, err := cidr.Host(subnet, 2)
	if err != nil {
		return nil, err
	}
	srv.ip = DHCPIP
	RouterIP, err := cidr.Host(subnet, 1)
	if err != nil {
		return nil, err
	}
	srv.router = RouterIP

	return srv, nil
}

//DHCPV4OnSubnet starts a DHCPv4 Server on a particular subnet
func (rd *RouterDHCP) DHCPV4OnSubnet() error {
	fmt.Println("Starting DHCPv4")

	laddr := net.UDPAddr{
		Port: 67,
	}

	server, err := server4.NewServer(rd.iface, &laddr, rd.HandleV4)
	if err != nil {
		return err
	}

	return server.Serve()
}

//HandleV4 handles DHCP v4
func (rd *RouterDHCP) HandleV4(conn net.PacketConn, peer net.Addr, m *dhcpv4.DHCPv4) {
	fmt.Println(m.Summary())
	switch m.MessageType() {
	case dhcpv4.MessageTypeDiscover:
		rd.handleDiscoveryRequest(conn, peer, m, dhcpv4.MessageTypeOffer)
		break
	case dhcpv4.MessageTypeRequest:
		rd.handleDiscoveryRequest(conn, peer, m, dhcpv4.MessageTypeAck)
		break
	}
}

func (rd *RouterDHCP) handleDiscoveryRequest(conn net.PacketConn, peer net.Addr, m *dhcpv4.DHCPv4, replyType dhcpv4.MessageType) {
	vmIP, err := rd.macToIP(m.ClientHWAddr)
	if err != nil {
		rd.sendDecline(conn, peer, m)
		return
	}

	if m.MessageType() == dhcpv4.MessageTypeRequest && !m.RequestedIPAddress().Equal(vmIP) {
		rd.sendDecline(conn, peer, m)
		log.Println("DHCPv4 sent decline due to req ip mismatch")
		return
	}

	mods := []dhcpv4.Modifier{
		dhcpv4.WithLeaseTime(rd.leaseTime),
		dhcpv4.WithServerIP(rd.ip),
		dhcpv4.WithYourIP(vmIP),
		dhcpv4.WithMessageType(replyType),
		dhcpv4.WithRouter(rd.router),
		dhcpv4.WithNetmask(rd.subnet.Mask),
	}

	mods = append(mods, rd.dhcpOptions()...)

	resp, err := dhcpv4.NewReplyFromRequest(m, mods...)
	if err != nil {
		fmt.Println("DHCPv4 Failed to construct DHCP offer: ", err)
		return
	}

	_, err = conn.WriteTo(resp.ToBytes(), peer)
	if err != nil {
		fmt.Println("DHCPv4 Failed to write DHCP offer: ", err)
	}
}

func (rd *RouterDHCP) sendDecline(conn net.PacketConn, peer net.Addr, m *dhcpv4.DHCPv4) {
	mods := []dhcpv4.Modifier{
		dhcpv4.WithMessageType(dhcpv4.MessageTypeDecline),
	}

	resp, err := dhcpv4.NewReplyFromRequest(m, mods...)
	if err != nil {
		log.Println("DHCPv4 Failed to construct decline: ", err)
		return
	}
	conn.WriteTo(resp.ToBytes(), peer)
}

func (rd *RouterDHCP) dhcpOptions() []dhcpv4.Modifier {
	return []dhcpv4.Modifier{
		dhcpv4.WithDNS(rd.dnsServers...),
	}
}

func (rd *RouterDHCP) macToIP(hwaddr net.HardwareAddr) (net.IP, error) {
	//TODO(tcfw) link to hyper service
	IP, err := cidr.Host(rd.subnet, 4)
	if err != nil {
		return nil, err
	}
	return IP, nil
}
