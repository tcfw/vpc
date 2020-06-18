package vpctp

import (
	"net"
	"testing"
	"time"

	"github.com/google/gopacket/layers"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/tcfw/vpc/pkg/l2/transport/tap/protocol"
	"github.com/vishvananda/netlink"
)

func TestMarshalIPv4(t *testing.T) {
	v := NewHandler()
	v.listenAddr = net.IP{127, 0, 0, 1}
	v.hwaddr = net.HardwareAddr{6, 5, 4, 3, 2, 1}
	v.ipType = layers.EthernetTypeIPv4
	v.fibCache[string(net.IP{127, 0, 0, 1}.To16())] = netlink.Neigh{
		IP:           net.IP{127, 0, 0, 1},
		HardwareAddr: net.HardwareAddr{1, 2, 3, 4, 5, 6},
	}

	p := &protocol.Packet{
		VNID:   5,
		Source: []byte("abcdefg"),
		Frame:  []byte{0x1, 0x2},
	}

	b, err := v.toBytes(p, net.IP{127, 0, 0, 1})
	if assert.NoError(t, err) {
		t.Logf("% X", b)
		pfb, _ := v.toPacket(b)

		assert.Equal(t, p.VNID, pfb.VNID)
		assert.Equal(t, p.Source, pfb.Source)
		assert.Equal(t, p.Frame, pfb.Frame)
	}
}

func TestMarshalIPv6(t *testing.T) {
	v := NewHandler()
	dstIP := net.ParseIP("2001:DB8::2")
	v.listenAddr = net.ParseIP("2001:DB8::1")
	v.hwaddr = net.HardwareAddr{6, 5, 4, 3, 2, 1}
	v.ipType = layers.EthernetTypeIPv6
	v.fibCache[string(dstIP.To16())] = netlink.Neigh{
		IP:           dstIP,
		HardwareAddr: net.HardwareAddr{1, 2, 3, 4, 5, 6},
	}

	p := &protocol.Packet{
		VNID:   5,
		Source: []byte("abcdefgihjkl"),
		Frame:  []byte{0x1, 0x2},
	}

	b, err := v.toBytes(p, dstIP)
	if assert.NoError(t, err) {
		t.Logf("% X", b)
		pfb, _ := v.toPacket(b)

		assert.Equal(t, p.VNID, pfb.VNID)
		assert.Equal(t, p.Source, pfb.Source)
		assert.Equal(t, p.Frame, pfb.Frame)
	}
}

func TestSend(t *testing.T) {
	//Needs root
	v := NewHandler()
	p := protocol.Packet{
		VNID:   5,
		Source: []byte("abcdef"),
		Frame:  []byte{0x1, 0x2},
	}

	received := 0
	handledSomething := false

	v.SetHandler(func(ps []protocol.Packet) {
		received += len(ps)
		handledSomething = true
	})

	err := v.Start()
	if assert.NoError(t, err) {
		defer v.Stop()

		v.Send([]protocol.Packet{p}, net.ParseIP("::1"))
	}

	for {
		if handledSomething {
			break
		}
		time.Sleep(1 * time.Millisecond)
	}

	assert.Equal(t, 1, received)
}

func BenchmarkIPv4ToBytes(b *testing.B) {
	v := NewHandler()
	v.listenAddr = net.IP{127, 0, 0, 1}
	v.ipType = layers.EthernetTypeIPv4
	v.hwaddr = net.HardwareAddr{6, 5, 4, 3, 2, 1}
	v.fibCache[string(net.IP{127, 0, 0, 1}.To16())] = netlink.Neigh{
		IP:           net.IP{127, 0, 0, 1},
		HardwareAddr: net.HardwareAddr{1, 2, 3, 4, 5, 6},
	}

	p := &protocol.Packet{
		VNID:   5,
		Source: []byte("abcdefghijklmnop"),
		Frame: []byte{
			0xCC, 0x03, 0x04, 0xDC, 0x00, 0x10, 0xFF, 0xAA, 0xFA, 0xAA, 0xFF, 0xAA, 0x81, 0x00,
			0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x7F, 0x00, 0x00, 0x01, 0x08, 0x08, 0x08, 0x08, 0x00, 0x08, 0x00, 0x00,
			0x00, 0x01, 0x00, 0x01, 0x30, 0x31, 0x32, 0x33, 0x34, 0x35, 0x36, 0x37, 0x38, 0x39,
			0x61, 0x62, 0x63, 0x64, 0x65, 0x66,
		},
	}

	ip := net.IP{127, 0, 0, 1}

	b.Run("toBytes", func(b *testing.B) {
		b.SetBytes(int64(4 + 2 + len(p.Source) + len(p.Frame)))
		for i := 0; i < b.N; i++ {
			v.toBytes(p, ip)
		}
	})
}

func BenchmarkIPv6ToBytes(b *testing.B) {
	v := NewHandler()
	v.listenAddr = net.ParseIP("::1")
	v.ipType = layers.EthernetTypeIPv6
	v.hwaddr = net.HardwareAddr{6, 5, 4, 3, 2, 1}
	v.fibCache[string(net.ParseIP("::1").To16())] = netlink.Neigh{
		IP:           net.ParseIP("::1"),
		HardwareAddr: net.HardwareAddr{1, 2, 3, 4, 5, 6},
	}

	p := &protocol.Packet{
		VNID:   5,
		Source: []byte("abcdefghijklmnop"),
		Frame: []byte{
			0xCC, 0x03, 0x04, 0xDC, 0x00, 0x10, 0xFF, 0xAA, 0xFA, 0xAA, 0xFF, 0xAA, 0x81, 0x00,
			0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x7F, 0x00, 0x00, 0x01, 0x08, 0x08, 0x08, 0x08, 0x00, 0x08, 0x00, 0x00,
			0x00, 0x01, 0x00, 0x01, 0x30, 0x31, 0x32, 0x33, 0x34, 0x35, 0x36, 0x37, 0x38, 0x39,
			0x61, 0x62, 0x63, 0x64, 0x65, 0x66,
		},
	}

	ip := net.ParseIP("::1")

	b.Run("toBytes", func(b *testing.B) {
		b.SetBytes(int64(4 + 2 + len(p.Source) + len(p.Frame)))
		for i := 0; i < b.N; i++ {
			v.toBytes(p, ip)
		}
	})
}

func BenchmarkSend(b *testing.B) {
	//Needs root
	v := NewHandler()
	p := []protocol.Packet{}
	n := 10
	for i := 0; i < n; i++ {
		p = append(p, protocol.Packet{
			VNID:   5,
			Source: []byte("abcdef"),
			Frame: []byte{
				0xCC, 0x03, 0x04, 0xDC, 0x00, 0x10, 0xFF, 0xAA, 0xFA, 0xAA, 0xFF, 0xAA, 0x81, 0x00,
				0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x7F, 0x00, 0x00, 0x01, 0x08, 0x08, 0x08, 0x08, 0x00, 0x08, 0x00, 0x00,
				0x00, 0x01, 0x00, 0x01, 0x30, 0x31, 0x32, 0x33, 0x34, 0x35, 0x36, 0x37, 0x38, 0x39,
				0x61, 0x62, 0x63, 0x64, 0x65, 0x66,
			}})
	}

	received := 0

	v.SetHandler(func(ps []protocol.Packet) {
		received += len(ps)
	})

	if err := v.Start(); err != nil {
		panic(err)
	}

	v.hwaddr = net.HardwareAddr{1, 2, 3, 4, 5, 6}

	defer v.Stop()
	dst := net.IP{127, 0, 0, 1}

	b.Run("Send", func(b *testing.B) {
		b.SetBytes(int64((4 + 2 + len(p[0].Source) + len(p[0].Frame)) * n))

		for i := 0; i < b.N; i++ {
			v.Send(p, dst)
		}
	})
}

func TestSendRecv(t *testing.T) {
	//Needs root

	received := []protocol.Packet{}

	vSender := NewHandler()
	vReceiver := NewHandler()
	var br *netlink.Bridge
	var vethSender *netlink.Veth
	var vethSenderPeer netlink.Link
	var vethReceiver *netlink.Veth
	var vethReceiverPeer netlink.Link

	defer func() {
		netlink.LinkDel(vethReceiver)
		netlink.LinkDel(vethSender)
		netlink.LinkDel(br)
	}()

	var n int
	var err error
	var tout time.Time

	p := []protocol.Packet{
		{
			VNID:   5,
			Source: []byte("abc"),
			Frame: []byte{
				0xCC, 0x03, 0x04, 0xDC, 0x00, 0x10, 0xFF, 0xAA, 0xFA, 0xAA, 0xFF, 0xAA, 0x81, 0x00,
				0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x7F, 0x00, 0x00, 0x01, 0x08, 0x08, 0x08, 0x08, 0x00, 0x08, 0x00, 0x00,
				0x00, 0x01, 0x00, 0x01, 0x30, 0x31, 0x32, 0x33, 0x34, 0x35, 0x36, 0x37, 0x38, 0x39,
				0x61, 0x62, 0x63, 0x64, 0x65, 0x66,
			},
		},
	}

	senderAddr, _ := netlink.ParseAddr("10.254.0.253/24")
	recvAddr, _ := netlink.ParseAddr("10.254.0.254/24")

	//bridge setup
	br = &netlink.Bridge{
		LinkAttrs: netlink.LinkAttrs{
			Name: "b-vpctp-test-sr",
		},
	}

	err = netlink.LinkAdd(br)
	if err != nil {
		t.Error(err)
		goto delBr
	}
	netlink.LinkSetUp(br)

	//test Interfaces
	vethSender = &netlink.Veth{
		LinkAttrs: netlink.LinkAttrs{
			MasterIndex: br.Index,
			Name:        "vpctp-send",
		},
		PeerName: "vpctp-send-p",
	}

	err = netlink.LinkAdd(vethSender)
	if err != nil {
		t.Error(err)
		goto delVethSend
	}
	vethSenderPeer, _ = netlink.LinkByName("vpctp-send-p")

	vethReceiver = &netlink.Veth{
		LinkAttrs: netlink.LinkAttrs{
			MasterIndex: br.Index,
			Name:        "vpctp-recv",
		},
		PeerName: "vpctp-recv-p",
	}

	err = netlink.LinkAdd(vethReceiver)
	if err != nil {
		t.Error(err)
		goto delVethRecv
	}
	vethReceiverPeer, _ = netlink.LinkByName("vpctp-recv-p")

	netlink.AddrAdd(vethSenderPeer, senderAddr)
	netlink.AddrAdd(vethReceiverPeer, recvAddr)

	netlink.LinkSetUp(vethSenderPeer)
	netlink.LinkSetUp(vethSender)
	netlink.LinkSetUp(vethReceiverPeer)
	netlink.LinkSetUp(vethReceiver)

	vReceiver.SetHandler(func(ps []protocol.Packet) {
		for _, p := range ps {
			if len(p.Frame) > 0 {
				received = append(received, p)
			}
		}
	})

	viper.Set("vtepdev", "vpctp-send-p")
	vSender.Start()
	viper.Set("vtepdev", "vpctp-recv-p")
	vReceiver.Start()

	vSender.fibCache[string(recvAddr.IP.To16())] = netlink.Neigh{
		State:        netlink.NUD_PERMANENT,
		IP:           recvAddr.IP,
		HardwareAddr: vethReceiverPeer.Attrs().HardwareAddr,
	}

	vReceiver.fibCache[string(senderAddr.IP.To16())] = netlink.Neigh{
		State:        netlink.NUD_PERMANENT,
		IP:           senderAddr.IP,
		HardwareAddr: vethSenderPeer.Attrs().HardwareAddr,
	}

	//Delay introduced as setting up the XDP socket can sometimes miss the first few packets
	//coming off the bridge
	time.Sleep(10 * time.Millisecond)

	n, err = vSender.Send(p, recvAddr.IP)
	if err != nil {
		t.Error(err)
		goto delVethRecv
	}
	if n == 0 {
		t.Error("n was zero, but no error")
		goto delVethRecv
	}

	tout = time.Now().Add(5 * time.Second)

	for {
		if time.Now().After(tout) || len(received) != 0 {
			break
		}

		time.Sleep(1 * time.Millisecond)
	}

	if len(received) == 0 {
		t.Error("received nothing after waiting :'(")
	} else {
		assert.Equal(t, received[0].VNID, p[0].VNID)
		assert.Equal(t, received[0].Source[:len(p[0].Source)], p[0].Source)
		assert.Equal(t, received[0].Frame, p[0].Frame)
	}

delVethRecv:
	netlink.LinkDel(vethReceiver)
delVethSend:
	netlink.LinkDel(vethSender)
delBr:
	netlink.LinkDel(br)
}

func TestArpRequest(t *testing.T) {
	v := NewHandler()
	v.hwaddr = net.HardwareAddr{6, 5, 4, 3, 2, 1}
	v.listenAddr = net.ParseIP("2.2.2.2")
	b, err := v.arpRequest(net.ParseIP("1.1.1.1"))
	assert.NoError(t, err)
	t.Logf("% X", b)
}

func TestIPv6NeighSoli(t *testing.T) {
	v := NewHandler()
	v.hwaddr = net.HardwareAddr{6, 5, 4, 3, 2, 1}
	v.listenAddr = net.ParseIP("2001:DB8::1")
	b, err := v.neighborSolicitation(net.ParseIP("2001:DB8::2"))
	assert.NoError(t, err)
	t.Fatalf("% X", b)
}
