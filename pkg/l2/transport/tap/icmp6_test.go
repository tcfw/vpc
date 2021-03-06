package tap

import (
	"log"
	"net"
	"testing"

	"github.com/google/gopacket/layers"
	"github.com/stretchr/testify/assert"
)

func TestICMPv6NSDecode(t *testing.T) {
	listener, err := NewListener()
	if !assert.NoError(t, err) {
		return
	}

	frame := []byte{
		0x33, 0x33, 0xff, 0x00, 0x00, 0x01, 0xd6, 0xde, 0x5b, 0x62, 0x52, 0x6a,
		0x86, 0xdd, 0x60, 0x00, 0x00, 0x00, 0x00, 0x20, 0x3a, 0xff, 0xfe, 0x80, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0xd4, 0xde, 0x5b, 0xff, 0xfe, 0x62, 0x52, 0x6a, 0xff, 0x02, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0xff, 0x00, 0x00, 0x01, 0x87, 0x00, 0x76, 0x44, 0x00, 0x00,
		0x00, 0x00, 0xfe, 0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x01, 0x01, 0x01, 0xd6, 0xde, 0x5b, 0x62, 0x52, 0x6a,
	}

	layers, err := listener.decodeICMPv6Frame(frame)
	if !assert.NoError(t, err) {
		return
	}

	// assert.Equal(t, uint16(5), layers.vlan.VLANIdentifier)
	assert.Equal(t, net.ParseIP("fe80::1"), layers.icmp.TargetAddress)
}

func TestICMPv6NSEncode(t *testing.T) {
	listener, err := NewListener()
	if !assert.NoError(t, err) {
		return
	}

	layers := &icmpv6Layers{
		ethernet: &layers.Ethernet{
			SrcMAC: net.HardwareAddr{0x82, 0x11, 0x0, 0x0, 0x0, 0xff},
			DstMAC: net.HardwareAddr{0x33, 0x33, 0x0, 0x0, 0x0, 0x0},
		},
		vlan: &layers.Dot1Q{
			VLANIdentifier: 5,
		},
		ip: &layers.IPv6{
			SrcIP: net.ParseIP("fe80::1"),
			DstIP: net.ParseIP("ff02::1:ff00:1"),
		},
		icmp: &layers.ICMPv6NeighborSolicitation{
			TargetAddress: net.ParseIP("fe80::2"),
		},
	}

	raw, err := listener.buildICMPv6NAResponse(layers, net.HardwareAddr{0x82, 0x11, 0x22, 0x33, 0x44, 0x55})
	if !assert.NoError(t, err) {
		return
	}

	log.Printf("%x", raw)

	assert.NotEmpty(t, raw)
}
