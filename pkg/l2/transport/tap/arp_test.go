package tap

import (
	"log"
	"net"
	"testing"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/stretchr/testify/assert"

	"github.com/tcfw/vpc/pkg/l2/controller"
	"github.com/tcfw/vpc/pkg/l2/transport/tap/protocol"
	"github.com/vishvananda/netlink"
)

func TestValidARPRequest(t *testing.T) {
	controller := &controller.NullController{}
	listener, err := NewListener()
	if !assert.NoError(t, err) {
		return
	}
	listener.SetSDN(controller)

	var vnid uint32 = 5

	listener.taps[vnid] = &Tap{
		iface: &netlink.Tuntap{LinkAttrs: netlink.LinkAttrs{
			HardwareAddr: net.HardwareAddr{0x86, 0x22, 0x00, 0x00, 0x00, 0xff},
		}},
	}

	eth := &layers.Ethernet{
		SrcMAC:       []byte{0x82, 0x11, 0x00, 0x00, 0x00, 0xff},
		DstMAC:       []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
		EthernetType: layers.EthernetTypeDot1Q,
	}
	vlan := &layers.Dot1Q{
		VLANIdentifier: 5,
		Type:           layers.EthernetTypeARP,
	}
	arp := &layers.ARP{
		AddrType:          layers.LinkTypeEthernet,
		Protocol:          layers.EthernetTypeIPv4,
		Operation:         layers.ARPRequest,
		HwAddressSize:     6,
		ProtAddressSize:   4,
		SourceHwAddress:   []byte{0x82, 0x11, 0x00, 0x00, 0x00, 0xff},
		SourceProtAddress: []byte{10, 4, 0, 2},
		DstHwAddress:      []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		DstProtAddress:    []byte{10, 4, 0, 1},
	}

	buf := gopacket.NewSerializeBuffer()
	gopacket.SerializeLayers(buf, gopacket.SerializeOptions{},
		eth,
		vlan,
		arp)

	frame := buf.Bytes()
	log.Printf("%x", frame)

	packet := &protocol.Packet{VNID: 5, Frame: frame}

	err = listener.arpReduce(packet)
	assert.NoError(t, err)

	// response := <-listener.taps[vnid].in
	// assert.NotEmpty(t, response.Frame)

}

func TestARPResponse(t *testing.T) {
	controller := &controller.NullController{}
	listener, err := NewListener()
	if !assert.NoError(t, err) {
		return
	}
	listener.SetSDN(controller)

	var vnid uint32 = 5

	listener.taps[vnid] = &Tap{
		iface: &netlink.Tuntap{LinkAttrs: netlink.LinkAttrs{
			HardwareAddr: net.HardwareAddr{0x86, 0x22, 0x00, 0x00, 0x00, 0xff},
		}},
	}

	//Req: who has 10.4.0.1 tell 10.4.0.2 (82:11:00:00:00:ff)
	arpRequest := []byte{
		0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0x82, 0x11, 0x00, 0x00, 0x00, 0xFF, 0x81, 0x00, 0x00, 0x05,
		0x08, 0x06, 0x00, 0x01, 0x08, 0x00, 0x06, 0x04, 0x00, 0x01, 0x82, 0x11, 0x00, 0x00, 0x00, 0xFF,
		0x0A, 0x04, 0x00, 0x02, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x0A, 0x04, 0x00, 0x01, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	}

	err = listener.arpReduce(&protocol.Packet{VNID: 5, Frame: arpRequest})
	assert.NoError(t, err)

	//Resp: 82:11:00:00:00:ff, 10.4.0.1 is at 82:2f:00:00:00:ff
	// response := <-listener.taps[vnid].in
	// if !assert.NotEmpty(t, response.Frame) {
	// 	return
	// }

	// log.Printf("%x", response.Frame)

	// packetData := gopacket.NewPacket(response.Frame, layers.LayerTypeEthernet, gopacket.Default)

	// eth := &layers.Ethernet{}
	// ethLayer := packetData.Layer(layers.LayerTypeEthernet)
	// if err := eth.DecodeFromBytes(ethLayer.LayerContents(), gopacket.NilDecodeFeedback); err != nil {
	// 	t.Fatalf("failed to decode ethernet frame: %s", err)
	// }

	// vlan := &layers.Dot1Q{}
	// dot1qLayer := packetData.Layer(layers.LayerTypeDot1Q)
	// if err := vlan.DecodeFromBytes(dot1qLayer.LayerContents(), gopacket.NilDecodeFeedback); err != nil {
	// 	t.Fatalf("failed to decode dot1q frame: %s", err)
	// }

	// arp := &layers.ARP{}
	// arpLayer := packetData.Layer(layers.LayerTypeARP)
	// if err := arp.DecodeFromBytes(arpLayer.LayerContents(), gopacket.NilDecodeFeedback); err != nil {
	// 	t.Fatalf("failed to decode arp packet: %s", err)
	// }

	// dstHwAddr := net.HardwareAddr{0x82, 0x11, 0x00, 0x00, 0x00, 0xff}

	// //From NullController static IP/MAC
	// remoteMac := net.HardwareAddr{0x82, 0x2f, 0x00, 0x00, 0x00, 0xff}

	// assert.Equal(t, dstHwAddr, eth.DstMAC)
	// assert.Equal(t, remoteMac, eth.SrcMAC)

	// assert.Equal(t, uint8(6), arp.HwAddressSize)
	// assert.Equal(t, uint8(4), arp.ProtAddressSize)
	// assert.Equal(t, uint16(2), arp.Operation) //Reply type

	// assert.Equal(t, remoteMac, net.HardwareAddr(arp.SourceHwAddress))
	// assert.Equal(t, dstHwAddr, net.HardwareAddr(arp.DstHwAddress))

	// assert.Equal(t, []byte{10, 4, 0, 1}, arp.SourceProtAddress)
	// assert.Equal(t, []byte{10, 4, 0, 2}, arp.DstProtAddress)
}
