package vpctp

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/asm"
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

	addrs, err := netlink.AddrList(link, netlink.FAMILY_V4)
	if err != nil || len(addrs) == 0 {
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

	//TODO(tcfw) support multiqueue devices
	xsk, err := xdp.NewTap(link, 0, func(xsksMap *ebpf.Map, qidMap *ebpf.Map) (*ebpf.Program, error) {
		return ebpf.NewProgram(&ebpf.ProgramSpec{
			Name:          "vpc_xsk_ebpf",
			Type:          ebpf.XDP,
			License:       "LGPL-2.1 or BSD-2-Clause",
			KernelVersion: 0,
			Instructions: asm.Instructions{
				//# int xdp_sock_prog(struct xdp_md *ctx) {
				{OpCode: 0xbf, Src: 0x1, Dst: 0x6}, //0: r6 = r1
				//# int *qidconf, index = ctx->rx_queue_index;
				{OpCode: 0x61, Src: 0x6, Dst: 0x1, Offset: 0x10}, //1: r1 = *(u32 *)(r6 + 16)
				{OpCode: 0x63, Src: 0x1, Dst: 0xa, Offset: -4},   //2: *(u32 *)(r10 - 4) = r1
				{OpCode: 0xbf, Src: 0xa, Dst: 0x2},               //3: r2 = r10
				{OpCode: 0x07, Dst: 0x2, Constant: -4},           //4: r2 += -4
				//# qidconf = bpf_map_lookup_elem(&qidconf_map, &index);
				{OpCode: 0x18, Src: 0x1, Dst: 0x1, Constant: int64(qidMap.FD())}, //5: r1 = 0 ll
				{OpCode: 0x85, Constant: 0x1},                                    //7: call 1
				{OpCode: 0xbf, Src: 0x0, Dst: 0x1},                               //8: r1 = r0
				{OpCode: 0xb7, Dst: 0x0, Constant: 0},                            //9: r0 = 0
				//# if (!qidconf)
				{OpCode: 0x15, Dst: 0x1, Offset: 0x1f},   //10: if r1 == 0 goto +31
				{OpCode: 0xb7, Dst: 0x0, Constant: 0x02}, //11: r0 = 2
				//# if (!*qidconf)
				{OpCode: 0x61, Src: 0x1, Dst: 0x1, Constant: 0}, //12: r1 = *(u32 *)(r1 + 0)
				{OpCode: 0x15, Dst: 0x1, Offset: 0x1c},          //13: if r1 == 0 goto +28
				//# void *data_end = (void *)(long)ctx->data_end;
				{OpCode: 0x61, Src: 0x6, Dst: 0x2, Offset: 0x04}, //14: r2 = *(u32 *)(r6 + 4)
				//# void *data = (void *)(long)ctx->data;
				{OpCode: 0x61, Src: 0x6, Dst: 0x1, Constant: 0}, //15: r1 = *(u32 *)(r6 + 0)
				//# if (data + nh_off > data_end)
				{OpCode: 0xbf, Src: 0x1, Dst: 0x3},               //16: r3 = r1
				{OpCode: 0x07, Dst: 0x3, Constant: 0x0e},         //17: r3 += 14
				{OpCode: 0x2d, Src: 0x2, Dst: 0x3, Offset: 0x17}, //18: if r3 > r2 goto +23
				//# h_proto = eth->h_proto;
				{OpCode: 0x71, Src: 0x1, Dst: 0x4, Offset: 0x0c}, //19: r4 = *(u8 *)(r1 + 12)
				{OpCode: 0x71, Src: 0x1, Dst: 0x3, Offset: 0x0d}, //20: r3 = *(u8 *)(r1 + 13)
				{OpCode: 0x67, Dst: 0x3, Constant: 0x08},         //21: r3 <<= 8
				{OpCode: 0x4f, Src: 0x4, Dst: 0x3},               //22: r3 |= r4
				//# if (h_proto == __bpf_constant_htons(ETH_P_IP))
				{OpCode: 0x15, Dst: 0x3, Offset: 0x06, Constant: 0x86dd}, //23: if r3 == 56710 goto +6
				{OpCode: 0x55, Dst: 0x3, Offset: 0x11, Constant: 0x08},   //24: if r3 != 8 goto +17
				{OpCode: 0xb7, Dst: 0x3, Constant: 0x17},                 //25: r3 = 23
				//# if (iph + 1 > data_end)
				{OpCode: 0xbf, Src: 0x1, Dst: 0x4},               //26: r4 = r1
				{OpCode: 0x07, Dst: 0x4, Constant: 0x22},         //27: r4 += 34
				{OpCode: 0x2d, Src: 0x2, Dst: 0x4, Offset: 0x0d}, //28: if r4 > r2 goto +13
				{OpCode: 0x05, Offset: 0x04},                     //29: goto +4
				{OpCode: 0xb7, Dst: 0x3, Constant: 0x14},         //30: r3 = 20
				//# if (ip6h + 1 > data_end)
				{OpCode: 0xbf, Src: 0x1, Dst: 0x4},               //31: r4 = r1
				{OpCode: 0x07, Dst: 0x4, Constant: 0x36},         //32: r4 += 54
				{OpCode: 0x2d, Src: 0x2, Dst: 0x4, Offset: 0x08}, //33: if r4 > r2 goto +8
				{OpCode: 0x0f, Src: 0x3, Dst: 0x1},               //34: r1 += r3
				{OpCode: 0x71, Src: 0x1, Dst: 0x1, Constant: 0},  //35: r1 = *(u8 *)(r1 + 0)
				//# if (ipproto != VPC_PROTO)
				{OpCode: 0x55, Dst: 0x1, Offset: 0x05, Constant: IPPROTO}, //36: if r1 != 172 goto +5
				//# return bpf_redirect_map(&xsks_map, index, 0);
				{OpCode: 0x61, Src: 0xa, Dst: 0x2, Offset: -4},                    //37: r2 = *(u32 *)(r10 - 4)
				{OpCode: 0x18, Src: 0x1, Dst: 0x1, Constant: int64(xsksMap.FD())}, //38: r1 = 0 ll
				{OpCode: 0xb7, Dst: 0x3, Constant: 0},                             //40: r3 = 0
				{OpCode: 0x85, Constant: 0x33},                                    //41: call 51
				//# }
				{OpCode: 0x95}, //42: exit
			},
		})
	})
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

func (vtp *VPCTP) listen() {
	pks := make([]protocol.Packet, 64)
	var i int
	for {
		_, n, err := vtp.xsk.BatchRead(vtp.rxDesc)
		if err != nil {
			log.Printf("failed to read new packets: %s", err)
			continue
		}

		for i = 0; i < n; i++ {
			pp, err := vtp.toPacket(vtp.rxDesc[i].Data[:vtp.rxDesc[i].Len])
			if err == nil {
				pks[i] = pp
			}
		}

		vtp.recv(pks[:n])
	}
}

func (vtp *VPCTP) toBytes(p *protocol.Packet, addr net.IP) ([]byte, error) {
	dstMac, err := vtp.dstMac(addr)
	if err != nil {
		return nil, err
	}

	vtp.buf.Reset()

	if vtp.ipType == layers.EthernetTypeIPv6 {
		vtp.prependEthIPV6(p, addr, dstMac)
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

func (vtp *VPCTP) prependEthIPV6(p *protocol.Packet, addr net.IP, dstMAC net.HardwareAddr) {
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

func (vtp *VPCTP) toPacket(b []byte) (protocol.Packet, error) {

	pckd := gopacket.NewPacket(b, layers.LayerTypeEthernet, gopacket.NoCopy)
	dNet := pckd.NetworkLayer()

	if dNet.LayerType() == layers.LayerTypeIPv4 && dNet.(*layers.IPv4).Protocol != IPPROTO {
		return protocol.Packet{}, fmt.Errorf("Invalid IPv4 protocol")
	} else if dNet.LayerType() == layers.LayerTypeIPv6 && dNet.(*layers.IPv6).NextHeader != IPPROTO {
		return protocol.Packet{}, fmt.Errorf("Invalid IPv6 protocol")
	}

	d := dNet.LayerPayload()
	pack := protocol.Packet{}

	//VNID
	pack.VNID = binary.LittleEndian.Uint32(d[:4])

	//Source
	pack.Source = bytes.TrimRight(d[4:20], "\x00")

	//Frame
	pack.Frame = d[20:]

	return pack, nil
}

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
