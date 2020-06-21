package vpctp

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/spf13/viper"
	"github.com/vishvananda/netlink"
	"golang.org/x/sys/unix"

	"github.com/tcfw/vpc/pkg/l2/transport/tap/protocol"
	"github.com/tcfw/vpc/pkg/l2/xdp"
)

const (
	//IPPROTO IP protocol number (unassigned)
	IPPROTO = 172

	//IPHopLimit set packet TTL
	IPHopLimit = 64
)

var (
	nullHandler = func(p []protocol.Packet) {}
)

//NewHandler creates a new VxLAN handler
func NewHandler() *VPCTP {
	vpctp := &VPCTP{
		txDesc:   make([]xdp.BatchDesc, 64),
		rxDesc:   make([]xdp.BatchDesc, 64),
		tbuf:     gopacket.NewSerializeBufferExpectedSize(42, 0),
		ribCache: map[string]netlink.Route{},
		fibCache: map[string]netlink.Neigh{},
		recv:     nullHandler,
	}

	for i := 0; i < 64; i++ {
		vpctp.txDesc[i] = xdp.BatchDesc{
			Data: make([]byte, 2048),
		}
		vpctp.rxDesc[i] = xdp.BatchDesc{
			Data: make([]byte, 2048),
		}
	}

	//Add v4/v6 loopbacks
	vpctp.fibCache[string(net.IPv4(127, 0, 0, 1).To16())] = netlink.Neigh{
		LinkIndex:    1,
		Family:       unix.AF_INET,
		State:        netlink.NUD_PERMANENT,
		IP:           net.IPv4(127, 0, 0, 1),
		HardwareAddr: net.HardwareAddr{0, 0, 0, 0, 0, 0},
	}

	vpctp.fibCache[string(net.ParseIP("::1").To16())] = netlink.Neigh{
		LinkIndex:    1,
		Family:       unix.AF_INET6,
		State:        netlink.NUD_PERMANENT,
		IP:           net.ParseIP("::1"),
		HardwareAddr: net.HardwareAddr{0, 0, 0, 0, 0, 0},
	}

	return vpctp
}

//VPCTP tunnel protocol raw socket handler
type VPCTP struct {
	sock net.PacketConn
	xsk  *xdp.Tap
	recv protocol.HandlerFunc

	buf    bytes.Buffer
	tbuf   gopacket.SerializeBuffer
	ipType layers.EthernetType

	txDesc []xdp.BatchDesc
	rxDesc []xdp.BatchDesc

	iface      netlink.Link
	hwaddr     net.HardwareAddr
	listenAddr net.IP

	ribCache map[string]netlink.Route
	fibCache map[string]netlink.Neigh
	ribCMu   sync.RWMutex
	fibCMu   sync.RWMutex
}

//Start starts listening for remote packets coming in
func (vtp *VPCTP) Start() error {
	dev := viper.GetString("vtepdev")
	if len(dev) == 0 {
		dev = "lo"
	}

	link, err := netlink.LinkByName(dev)
	if err != nil {
		return err
	}
	vtp.iface = link
	vtp.hwaddr = link.Attrs().HardwareAddr

	//Check IPv6 first
	// ipv6Addrs, err := netlink.AddrList(link, netlink.FAMILY_V6)
	// if err != nil {
	// 	return fmt.Errorf("failed to quer for ipv6 addrs: %s", err)
	// }
	ipv4Addrs, err := netlink.AddrList(link, netlink.FAMILY_V4)
	if err != nil {
		return fmt.Errorf("failed to quer for ipv6 addrs: %s", err)
	}

	// addrs := append(ipv6Addrs, ipv4Addrs...)
	addrs := ipv4Addrs
	if len(addrs) == 0 {
		return fmt.Errorf("failed to get addr from dev; is it set up?")
	}

	for _, addr := range addrs {
		if addr.IP.IsLinkLocalUnicast() || addr.IP.IsGlobalUnicast() || addr.IP.IsLoopback() {
			vtp.listenAddr = addr.IP
			break
		}
	}
	if vtp.listenAddr == nil {
		log.Printf("!! No suitable listening addr found for interface %s", link.Attrs().Name)
	} else {
		if vtp.listenAddr.To4() != nil {
			vtp.ipType = layers.EthernetTypeIPv4
		} else {
			vtp.ipType = layers.EthernetTypeIPv6
		}
	}

	//TODO(tcfw) support multiqueue devices against ethtool
	xsk, err := xdp.NewTap(link, 0, vpctpEbpfProg)
	if err != nil {
		return err
	}
	vtp.xsk = xsk

	go vtp.fibRib()
	go vtp.listen()

	return nil
}

func (vtp *VPCTP) fibRib() {
	go vtp.monitorFIB()
	go vtp.monitorRIB()

	select {}
}

func (vtp *VPCTP) monitorFIB() {
	ifaceI := vtp.iface.Attrs().Index
	updates := make(chan netlink.NeighUpdate, 100)
	done := make(chan struct{})

	if err := netlink.NeighSubscribeWithOptions(updates, done, netlink.NeighSubscribeOptions{ListExisting: true}); err != nil {
		log.Printf("failed to start netlink FIB monitor: %s", err)
		return
	}

	for update := range updates {
		if update.Neigh.LinkIndex != ifaceI {
			continue
		}
		vtp.fibCMu.Lock()
		cacheK := string(update.Neigh.IP.To16())
		if update.Type == unix.RTM_NEWNEIGH && (update.Neigh.IP.IsLinkLocalUnicast() || update.Neigh.IP.IsGlobalUnicast()) {
			// log.Printf("FIB-U: %+v", update.Neigh)
			vtp.fibCache[cacheK] = update.Neigh
		} else if _, ok := vtp.fibCache[cacheK]; update.Type == unix.RTM_DELNEIGH && ok {
			delete(vtp.fibCache, cacheK)
		}
		vtp.fibCMu.Unlock()
	}
}

func (vtp *VPCTP) monitorRIB() {
	ifaceI := vtp.iface.Attrs().Index
	updates := make(chan netlink.RouteUpdate, 100)
	done := make(chan struct{})

	if err := netlink.RouteSubscribeWithOptions(updates, done, netlink.RouteSubscribeOptions{ListExisting: true}); err != nil {
		log.Printf("failed to start netlink RIB monitor: %s", err)
		return
	}

	for update := range updates {
		if update.Route.LinkIndex != ifaceI || update.Route.Dst == nil {
			continue
		}
		vtp.ribCMu.Lock()
		cacheK := update.Route.Dst.String()
		if update.Type == unix.RTM_NEWROUTE {
			vtp.ribCache[cacheK] = update.Route
			// log.Printf("RIB-U: %+v", update.Route)
		} else if _, ok := vtp.ribCache[cacheK]; update.Type == unix.RTM_DELROUTE && ok {
			delete(vtp.ribCache, cacheK)
		}
		vtp.ribCMu.Unlock()
	}

}

//SetHandler sets the receiving handler callback func
func (vtp *VPCTP) SetHandler(h protocol.HandlerFunc) {
	vtp.recv = h
}

//Stop stop the listening socket
func (vtp *VPCTP) Stop() error {
	return vtp.xsk.Close()
}

//Send sends a group of packets to the specified destination
func (vtp *VPCTP) Send(ps []protocol.Packet, addr net.IP) (int, error) {
	for i, p := range ps {
		pkd, err := vtp.toBytes(&p, addr)
		if err != nil {
			log.Printf("failed to wrap packet: %s", err)
			continue
		}

		vtp.txDesc[i].Data = pkd
		vtp.txDesc[i].Len = len(pkd)
	}

	return vtp.xsk.BatchWrite(vtp.txDesc[:len(ps)])
}

//listen reads from the XDP UMEM reading packets off the network interface
//Also starts up the backup protocol handler
func (vtp *VPCTP) listen() {
	go vtp.listenPC()

	pks := make([]protocol.Packet, 64)
	var i int
	for {
		_, n, err := vtp.xsk.BatchRead(vtp.rxDesc)
		if err != nil {
			log.Printf("failed to read new packets: %s", err)
			continue
		}

		for i = 0; i < n; i++ {
			pp, err := vtp.toPacket(vtp.rxDesc[i].Data[:vtp.rxDesc[i].Len], true)
			if err == nil {
				pks[i] = pp
			}
		}

		vtp.recv(pks[:n])
	}
}

//listenPC starts a _backup_ protocol handler
//This is used when some other interface picks up our protocol outside of the XDP
//like a bridge
func (vtp *VPCTP) listenPC() {
	pconn, err := net.ListenPacket("ip:172", "")
	if err != nil {
		log.Printf("failed to start loopback protocol hander: %s", err)
		return
	}

	buf := make([]byte, 2048)
	for {
		_, _, err := pconn.ReadFrom(buf)
		if err != nil {
			log.Printf("vpctp-pconn: failed to read packet: %s", err)
			continue
		}

		p, err := vtp.toPacket(buf, false)
		if err != nil {
			log.Printf("vpctp-pconn: invalid packet received")
		}

		vtp.recv([]protocol.Packet{p})
	}
}

//toBytes converts the raw VPC packet to a lower-level network packet
func (vtp *VPCTP) toBytes(p *protocol.Packet, addr net.IP) ([]byte, error) {
	dstMac, err := vtp.dstMac(addr)
	if err != nil {
		return nil, err
	}

	vtp.buf.Reset()

	if vtp.ipType == layers.EthernetTypeIPv6 {
		vtp.ethIPV6Frames(p, addr, dstMac)
	}

	vtp.buf.WriteByte(byte(p.VNID))
	vtp.buf.WriteByte(byte(p.VNID >> 8))
	vtp.buf.WriteByte(byte(p.VNID >> 16))
	vtp.buf.WriteByte(byte(p.VNID >> 24))

	src := p.Source

	if len(p.Source) < 16 {
		//pad source to be at least 16 bytes
		pad := bytes.Repeat([]byte{byte(0)}, 16-len(p.Source))
		src = append(src, pad...)
	}

	vtp.buf.Write(src[:16])

	vtp.buf.Write(p.Frame)

	if vtp.ipType == layers.EthernetTypeIPv6 {
		return vtp.buf.Bytes(), nil
	}

	var ipHdr gopacket.SerializableLayer

	//TODO(@tcfw) performance gain here by avoiding allocs to gopacket layers
	//IPv4 will get ugly with checksums

	ipHdr = &layers.IPv4{
		Version:  4,
		TTL:      IPHopLimit,
		SrcIP:    vtp.listenAddr,
		DstIP:    addr.To4(),
		Protocol: IPPROTO,
	}

	err = gopacket.SerializeLayers(vtp.tbuf, gopacket.SerializeOptions{ComputeChecksums: true, FixLengths: true},
		&layers.Ethernet{
			DstMAC:       dstMac,
			SrcMAC:       vtp.hwaddr,
			EthernetType: vtp.ipType,
		},
		ipHdr,
		gopacket.Payload(vtp.buf.Bytes()),
	)

	return vtp.tbuf.Bytes(), err
}

//ethIPV6Frames prepends the ethernet & ipv6 frames to the main buffer
//the IPv6 flow label is set to the LSBs of the VNID
func (vtp *VPCTP) ethIPV6Frames(p *protocol.Packet, addr net.IP, dstMAC net.HardwareAddr) {
	//eth
	vtp.buf.Write(dstMAC)
	vtp.buf.Write(vtp.hwaddr)
	// PutUint16
	vtp.buf.WriteByte(byte(vtp.ipType >> 8))
	vtp.buf.WriteByte(byte(vtp.ipType))

	//ipv6
	var tc uint16
	pLen := len(p.Frame) + 64                                 //(VNID+Source+InnerFrame)
	vtp.buf.WriteByte(byte(6<<4 | tc>>4))                     //ver | tc
	vtp.buf.WriteByte(byte(uint8(tc)<<4 | uint8(p.VNID>>16))) //tc | fl([bits :4])
	// PutUint16
	vtp.buf.WriteByte(byte(p.VNID)) //fl[bits 4:]
	vtp.buf.WriteByte(byte(p.VNID >> 8))
	// PutUint16
	vtp.buf.WriteByte(byte(pLen))
	vtp.buf.WriteByte(byte(pLen >> 8))
	vtp.buf.WriteByte(IPPROTO)    //nh
	vtp.buf.WriteByte(IPHopLimit) //hl
	vtp.buf.Write(vtp.listenAddr) //src
	vtp.buf.Write(addr)           //dst
}

//toPacket converts the raw packet to a usable VPC packet
func (vtp *VPCTP) toPacket(b []byte, includesNet bool) (protocol.Packet, error) {
	var d []byte
	if includesNet {
		pckd := gopacket.NewPacket(b, layers.LayerTypeEthernet, gopacket.NoCopy)
		dNet := pckd.NetworkLayer()

		if dNet.LayerType() == layers.LayerTypeIPv4 && dNet.(*layers.IPv4).Protocol != IPPROTO {
			return protocol.Packet{}, fmt.Errorf("Invalid IPv4 protocol")
		} else if dNet.LayerType() == layers.LayerTypeIPv6 && dNet.(*layers.IPv6).NextHeader != IPPROTO {
			return protocol.Packet{}, fmt.Errorf("Invalid IPv6 protocol")
		}

		d = dNet.LayerPayload()
	} else {
		d = b
	}

	pack := protocol.Packet{}

	//VNID
	pack.VNID = binary.LittleEndian.Uint32(d[:4])

	//Source
	pack.Source = bytes.TrimRight(d[4:20], "\x00")

	//Frame
	pack.Frame = d[20:]

	return pack, nil
}

//dstMac calculates the destination MAC address of addr
//if the neighbor doesn't exist, it will attempt to discover
func (vtp *VPCTP) dstMac(addr net.IP) (net.HardwareAddr, error) {
	vtp.fibCMu.RLock()
	defer vtp.fibCMu.RUnlock()

	//has direct fib entry
	if neigh, ok := vtp.fibCache[string(addr.To16())]; ok {
		return neigh.HardwareAddr, nil
	}

	vtp.ribCMu.RLock()
	defer vtp.ribCMu.RUnlock()

	var gw net.IP
	for _, route := range vtp.ribCache {
		if route.Dst != nil && route.Gw == nil && route.Dst.Contains(addr) { //Directly connected network
			return vtp.arp(addr)
		} else if route.Dst != nil && route.Dst.Contains(addr) && route.Gw != nil { //GW connected network
			gw = route.Gw
		} else if route.Gw != nil && route.Dst == nil { //Default GW
			if gw != nil {
				gw = route.Gw
			}
		}
	}

	if gw == nil {
		return nil, fmt.Errorf("failed to find suitable neigh")
	}

	if neigh, ok := vtp.fibCache[string(gw.To16())]; ok {
		return neigh.HardwareAddr, nil
	}

	//Send ARP for GW
	gwMac, err := vtp.arp(gw)
	if err != nil {
		return nil, err
	}
	return gwMac, nil
}

//arp constructs a simple arp packet
func (vtp *VPCTP) arp(addr net.IP) (net.HardwareAddr, error) {
	log.Printf("Looking up: %s", addr)

	errCh := make(chan error)
	dstCh := make(chan net.HardwareAddr)
	doneCh := make(chan struct{})
	updatesCh := make(chan netlink.NeighUpdate, 100)

	//Stop the neighbor subscription
	defer func() {
		doneCh <- struct{}{}
	}()

	netlink.NeighSubscribe(updatesCh, doneCh)

	go func() {
		for update := range updatesCh {
			if update.Type == unix.RTM_NEWNEIGH {
				continue
			}
			if update.Neigh.IP.Equal(addr) {
				dstCh <- update.Neigh.HardwareAddr
				break
			}
		}
	}()

	var packet []byte
	var err error

	if vtp.ipType == layers.EthernetTypeIPv4 {
		packet, err = vtp.arpRequest(addr)
	} else {
		packet, err = vtp.neighborSolicitation(addr)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to construct arp request: %s", err)
	}

	if _, err = vtp.xsk.Write(packet); err != nil {
		return nil, fmt.Errorf("failed to send arp: %s", err)
	}

	select {
	case <-time.After(3 * time.Second):
		return nil, fmt.Errorf("arp timeout")
	case err := <-errCh:
		return nil, err
	case dst := <-dstCh:
		return dst, nil
	}
}

func (vtp *VPCTP) arpRequest(addr net.IP) ([]byte, error) {
	buf := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{
		FixLengths:       true,
		ComputeChecksums: true,
	}

	err := gopacket.SerializeLayers(buf, opts,
		&layers.Ethernet{
			DstMAC:       net.HardwareAddr{0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
			SrcMAC:       vtp.hwaddr,
			EthernetType: layers.EthernetTypeARP,
		},
		&layers.ARP{
			AddrType:          layers.LinkTypeEthernet,
			Protocol:          layers.EthernetTypeIPv4,
			HwAddressSize:     6,
			ProtAddressSize:   4,
			Operation:         layers.ARPRequest,
			SourceHwAddress:   []byte(vtp.hwaddr),
			SourceProtAddress: []byte(vtp.listenAddr.To4()),
			DstHwAddress:      make([]byte, len(vtp.hwaddr)),
			DstProtAddress:    []byte(addr.To4()),
		},
	)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (vtp *VPCTP) neighborSolicitation(addr net.IP) ([]byte, error) {
	ip16 := addr.To16()
	if ip16 == nil {
		return nil, fmt.Errorf("incompatible IP addr")
	}

	dstIP := net.ParseIP(fmt.Sprintf("ff02::1:ff:%x:%x:%x", ip16[13], ip16[14], ip16[15]))
	dstHwAddr := net.HardwareAddr{0x33, 0x33, 0xff, ip16[13], ip16[14], ip16[15]}

	buf := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{
		FixLengths:       true,
		ComputeChecksums: true,
	}
	ipv6 := &layers.IPv6{
		Version:    6,
		HopLimit:   IPHopLimit,
		SrcIP:      vtp.listenAddr,
		DstIP:      dstIP,
		NextHeader: layers.IPProtocolICMPv6,
	}
	icmpv6 := &layers.ICMPv6{
		TypeCode: layers.CreateICMPv6TypeCode(layers.ICMPv6TypeNeighborSolicitation, 0),
	}
	icmpv6.SetNetworkLayerForChecksum(ipv6)

	err := gopacket.SerializeLayers(buf, opts,
		&layers.Ethernet{
			DstMAC:       dstHwAddr,
			SrcMAC:       vtp.hwaddr,
			EthernetType: layers.EthernetTypeIPv6,
		},
		ipv6,
		icmpv6,
		&layers.ICMPv6NeighborSolicitation{
			TargetAddress: addr,
		},
	)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
