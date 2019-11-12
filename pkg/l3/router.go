package l3

import (
	"fmt"
	"math/rand"
	"net"
	"time"

	"github.com/coreos/go-iptables/iptables"
	"github.com/lorenzosaino/go-sysctl"
	"github.com/tcfw/vpc/pkg/l2"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
)

//Router provides a iso layer 3 router using networknamespaces
type Router struct {
	ID    string
	VPCID int32
	NetNS netns.NsHandle
	Veths map[string]netlink.Link
	ExtBr *netlink.Bridge

	stack    *l2.Stack
	iptChain int32
}

//CreateRouter inits a router given a VPC stack
func CreateRouter(stack *l2.Stack, id string) (*Router, error) {
	router := &Router{
		VPCID:    stack.VPCID,
		ID:       id,
		Veths:    map[string]netlink.Link{},
		stack:    stack,
		iptChain: stack.VPCID,
	}

	ns, err := createNetNS(fmt.Sprintf("r-%s", id))
	if err != nil {
		return nil, err
	}

	router.NetNS = ns

	extBr, _ := netlink.LinkByName("virbr0")
	router.ExtBr = extBr.(*netlink.Bridge)

	err = router.init()

	return router, err
}

func (r *Router) init() error {
	r.Ifup("lo")
	r.EnableForwarding()

	exID := fmt.Sprintf("rx-%s", r.ID)
	r.CreateVeth(r.ExtBr, exID, "eth0", "any")

	r.Exec(func() error {
		//External
		eth0, _ := netlink.LinkByName("eth0")
		extNetwork, _ := netlink.ParseIPNet("192.168.122.254/24")
		netlink.AddrAdd(eth0, &netlink.Addr{IPNet: extNetwork})
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

	r.EnableNATOn("eth0")

	intIP, _ := netlink.ParseIPNet("10.4.0.1/24")
	r.AddSubnet(intIP)

	return r.Exec(func() error {
		id := fmt.Sprintf("r-%s", r.ID)
		fmt.Printf("Router (%s) is up!\n", id)
		return nil
	})
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
			fmt.Printf("%s\n\n", err)
		}
		return nil
	})
}

//AddSubnet creates a new veth pair and adds a specific subnet to it
//The veth will come up with a specific mac address based on the number
// of subnets already created - see subnetMacs()
func (r *Router) AddSubnet(subnet *net.IPNet) (netlink.Link, error) {
	c := len(r.Veths)
	ethID := fmt.Sprintf("eth%d", c)
	id := fmt.Sprintf("r-%s-%d", r.ID, c)

	// macs := subnetMacs()

	veth, err := r.CreateVeth(r.stack.Bridge, id, ethID, "any")
	if err != nil {
		return nil, err
	}

	err = r.Exec(func() error {
		eth, _ := netlink.LinkByName(ethID)
		netlink.AddrAdd(eth, &netlink.Addr{IPNet: subnet})
		return netlink.LinkSetUp(eth)
	})

	return veth, err
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
			fmt.Println("Failed to set mac", err)
		}
	}

	if err := netlink.LinkSetNsFd(routerEth, int(r.NetNS)); err != nil {
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
	for _, veth := range r.Veths {
		netlink.LinkDel(veth)
	}

	if err := unbindNSName(fmt.Sprintf("r-%s", r.ID)); err != nil {
		fmt.Println(err)
	}

	err := r.NetNS.Close()
	r.NetNS = 0
	return err
}

//Exec executes a given func inside the router network namespace
func (r *Router) Exec(fn func() error) error {
	return execInNetNs(r.NetNS, fn)
}

//NewID generates a unique id which can be assigned to routers
func NewID() string {
	length := 5

	charset := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	var seededRand *rand.Rand = rand.New(
		rand.NewSource(time.Now().UnixNano()))

	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}
