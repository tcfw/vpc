package l3

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/coreos/go-iptables/iptables"
	"github.com/lorenzosaino/go-sysctl"
	l2api "github.com/tcfw/vpc/pkg/api/v1/l2"
	"github.com/tcfw/vpc/pkg/l2"
	"github.com/tcfw/vpc/pkg/l3/ns"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
)

//Router provides a iso layer 3 router using networknamespaces
type Router struct {
	ID    string
	VPCID int32
	NetNS *ns.NetNS
	Veths map[string]netlink.Link
	ExtBr *netlink.Bridge

	l2       l2api.L2ServiceClient
	stack    *l2.Stack
	iptChain int32
	subnets  map[string]*Subnet

	bgp *RouterBGP
}

//Subnet contains info on a subnet and related interfaces
type Subnet struct {
	id       string
	vlan     uint16
	iface    netlink.Link
	ip       net.IP
	network  *net.IPNet
	vethPeer string
	innerMac net.HardwareAddr
}

//CreateRouter inits a router given a VPC stack
func CreateRouter(l2 l2api.L2ServiceClient, stack *l2.Stack, id string) (*Router, error) {
	router := &Router{
		VPCID:    stack.VPCID,
		ID:       id,
		Veths:    map[string]netlink.Link{},
		l2:       l2,
		stack:    stack,
		iptChain: stack.VPCID,
		subnets:  map[string]*Subnet{},
	}

	ns, err := ns.CreateNetNS(fmt.Sprintf("r-%s", id))
	if err != nil {
		return nil, fmt.Errorf("failed to create ns: %s", err)
	}

	router.NetNS = ns

	extBr, _ := netlink.LinkByName(configPublicBridge())
	router.ExtBr = extBr.(*netlink.Bridge)

	err = router.init()

	return router, err
}

func (r *Router) init() error {
	r.Ifup("lo")
	r.EnableForwarding()

	r.Exec(func() error {
		return r.SetDefaultFWRules()
	})

	pubIP, _ := netlink.ParseIPNet("192.168.122.254/24")

	exID := fmt.Sprintf("rx-%s", r.ID)
	r.CreateVeth(r.ExtBr, exID, "eth0", "any")

	r.Exec(func() error {
		//TODO(tcfw) maybe use DHCP?
		//External
		eth0, _ := netlink.LinkByName("eth0")
		netlink.AddrAdd(eth0, &netlink.Addr{IPNet: pubIP})
		netlink.LinkSetUp(eth0)

		//Use L4 hashing for ECMP
		sysctl.Set("net.ipv4.fib_multipath_hash_policy", "1")

		//Default GW(s)
		dst, _ := netlink.ParseIPNet("0.0.0.0/0")
		netlink.RouteReplace(&netlink.Route{Dst: dst, MultiPath: []*netlink.NexthopInfo{
			{
				LinkIndex: eth0.Attrs().Index,
				Gw:        net.IPv4(192, 168, 122, 1),
			},
		}})
		return nil

	})

	if configNAT() {
		r.EnableNATOn("eth0")
	}

	go func() {
		if err := r.Exec(func() error {
			ns, _ := netns.Get()
			peers := configBGPPeers()
			r.bgp, _ = NewRouterBGP(pubIP.IP, uint32(r.stack.VPCID), peers)
			if err := r.bgp.Start(ns); err != nil {
				log.Printf("Failed to start BGP: %s", err)
			}

			select {}
		}); err != nil {
			r.Delete()
			log.Fatalf("BGP: %s", err)
		}
	}()

	for _, subnetConfig := range configSubnets() {
		parts := strings.Split(subnetConfig, ":")
		if len(parts) != 2 {
			r.Delete()
			log.Fatalf("Invalid subnet config: %s. See --help for info", subnetConfig)
		}
		ip, cidr, err := net.ParseCIDR(parts[0])
		if err != nil {
			r.Delete()
			log.Fatalf("Invalid subnet (%s): %s", parts[0], err)
		}
		vlan, err := strconv.Atoi(parts[1])
		if err != nil {
			r.Delete()
			log.Fatalf("Invalid vlan: %s", err)
		}

		if err := r.AddSubnet(ip, cidr, uint16(vlan), configDHCP()); err != nil {
			r.Delete()
			return err
		}
	}

	log.Printf("Router (r-%s) is up!\n", r.ID)
	return nil
}

//AddSubnet attaches a new interface listening to a cidr and optionally enables DHCP
func (r *Router) AddSubnet(ip net.IP, cidr *net.IPNet, vlan uint16, dhcp bool) error {
	subnet, err := r.AddSubnetIFace(ip, cidr, vlan)
	if err != nil {
		return fmt.Errorf("failed to add subnet iface: %s", err)
	}

	if dhcp {
		go func() {
			err := r.Exec(func() error {
				dns := configDHCPDNS()
				dhcp, err := NewDHCPv4Server(subnet.vethPeer, cidr, dns)
				if err != nil {
					return err
				}
				return dhcp.DHCPV4OnSubnet()
			})
			if err != nil {
				log.Println(err)
			}
		}()
	}
	return nil
}

//EnableForwarding turns on ip forwarding via sysctl for packet routing
func (r *Router) EnableForwarding() error {
	return r.Exec(func() error {
		if err := sysctl.Set("net.ipv4.ip_forward", "1"); err != nil {
			return err
		}

		return sysctl.Set("net.ipv6.ip_forward", "1")
	})
}

//EnableNATOn adds iptable rules to enable natting on a specific interface
func (r *Router) EnableNATOn(iface string) error {
	return r.Exec(func() error {
		ipt, _ := iptables.New()
		if err := ipt.Append("nat", "POSTROUTING", "-o", iface, "-j", "MASQUERADE"); err != nil {
			log.Printf("%s\n\n", err)
		}
		return nil
	})
}

//AddSubnetIFace creates a new veth pair and adds a specific subnet to it
//The veth will come up with a specific mac address based on the number
// of subnets already created - see subnetMacs()
func (r *Router) AddSubnetIFace(addr net.IP, ipnet *net.IPNet, innerVlan uint16) (*Subnet, error) {
	c := len(r.Veths)
	ethID := fmt.Sprintf("eth%d", c)
	innerID := fmt.Sprintf("%s-%d", r.ID, c)
	id := fmt.Sprintf("n-%s", innerID)

	veth, err := r.CreateVeth(r.stack.Bridge, id, ethID, "any")
	if err != nil {
		return nil, err
	}

	netlink.BridgeVlanDel(veth, 1, true, true, false, false)

	if err := netlink.BridgeVlanAdd(veth, innerVlan, true, true, false, false); err != nil {
		return nil, fmt.Errorf("Failed to add VLAN to veth: %s", err)
	}

	var hwaddr net.HardwareAddr

	addrs := []netlink.Addr{}

	err = r.Exec(func() error {
		eth, _ := netlink.LinkByName(ethID)
		hwaddr = eth.Attrs().HardwareAddr

		netlink.LinkSetUp(eth)

		existingAddrs, err := netlink.AddrList(eth, netlink.FAMILY_V6)
		if err != nil {
			log.Println("failed to fetch existing addrs: ", err)
		} else {
			addrs = existingAddrs
		}

		addrNet := &net.IPNet{IP: addr, Mask: ipnet.Mask}
		return netlink.AddrAdd(eth, &netlink.Addr{IPNet: addrNet})
	})
	addrs = append(addrs, netlink.Addr{IPNet: &net.IPNet{IP: addr, Mask: ipnet.Mask}})

	addrsStrings := []string{}
	for _, ethAddr := range addrs {
		addrsStrings = append(addrsStrings, ethAddr.IP.String())
	}

	log.Printf("MAC: %s", hwaddr)

	if _, err := r.l2.AddNIC(context.Background(), &l2api.NicRequest{
		Id:            innerID,
		VpcId:         r.VPCID,
		SubnetVlanId:  uint32(innerVlan),
		ManuallyAdded: true,
		ManualHwaddr:  hwaddr.String(),
		Ip:            addrsStrings,
	}); err != nil {
		log.Println(err)
	}

	routerEth := r.Veths[id].(*netlink.Veth).PeerName

	subn := &Subnet{id: id, iface: veth, vethPeer: routerEth, vlan: innerVlan, ip: addr, network: ipnet, innerMac: hwaddr}

	r.subnets[id] = subn

	return subn, err
}

//Ifup set the link into the 'up' state
func (r *Router) Ifup(iface string) error {
	return r.Exec(func() error {
		dev, err := netlink.LinkByName(iface)
		if err != nil {
			return err
		}

		return netlink.LinkSetUp(dev)
	})
}

//CreateVeth creates a new veth pair attaching one side to a bridge and the
//other into the network namespace
func (r *Router) CreateVeth(bridge *netlink.Bridge, name string, peerName string, hwaddr string) (netlink.Link, error) {
	la := netlink.NewLinkAttrs()
	la.Name = name
	la.MTU = 1300

	veth := &netlink.Veth{
		LinkAttrs: la,
		PeerName:  peerName,
	}

	if err := netlink.LinkAdd(veth); err != nil {
		return nil, err
	}

	//Moved veth endpoint to router ns
	routerEth, err := netlink.LinkByName(veth.PeerName)
	if err != nil {
		return nil, err
	}

	if hwaddr != "any" {
		hwaddr, _ := net.ParseMAC(hwaddr)
		if err := netlink.LinkSetHardwareAddr(routerEth, hwaddr); err != nil {
			log.Println("Failed to set mac", err)
		}
	}

	if err := netlink.LinkSetNsFd(routerEth, r.NetNS.Fd()); err != nil {
		return nil, err
	}

	if err := netlink.LinkSetMaster(veth, bridge); err != nil {
		return nil, err
	}

	r.Veths[name] = veth

	err = netlink.LinkSetUp(veth)

	return veth, err
}

//Delete deletes all attached veth pairs and unbinds+deletes the netns
func (r *Router) Delete() error {
	for _, subnet := range r.subnets {
		existingAddrs := []netlink.Addr{}
		r.Exec(func() error {
			eth, _ := netlink.LinkByName(subnet.vethPeer)
			addrs, _ := netlink.AddrList(eth, netlink.FAMILY_ALL)
			existingAddrs = addrs
			return nil
		})
		addrString := []string{}
		for _, ethAddr := range existingAddrs {
			addrString = append(addrString, ethAddr.IP.String())
		}

		if _, err := r.l2.DeleteNIC(context.Background(), &l2api.Nic{
			Id:    subnet.id[2:], //exclude "n-..." for l2 id refs
			Index: int32(subnet.iface.Attrs().Index),
			VpcId: r.VPCID,
			Vlan:  uint32(subnet.vlan),
			Ip:    addrString,
		}); err != nil {
			log.Println(err)
		}
	}

	for _, veth := range r.Veths {
		netlink.LinkDel(veth)
	}

	err := r.NetNS.Close()
	r.NetNS = nil
	return err
}

//Exec executes a given func inside the router network namespace
func (r *Router) Exec(fn func() error) error {
	return r.NetNS.Exec(fn)
}

//NewID generates a unique id which can be assigned to routers
func NewID() string {
	length := 6

	charset := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	var seededRand *rand.Rand = rand.New(
		rand.NewSource(time.Now().UnixNano()))

	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}
