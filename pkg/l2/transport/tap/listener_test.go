package tap

import (
	"net"
	"testing"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/tcfw/vpc/pkg/l2/transport/tap/protocol"
)

func BenchmarkL2Fwd(b *testing.B) {
	var vnid uint32 = 1

	handler := &protocol.TestHandler{}
	fdb := NewFDB()
	tap := &testTap{
		vnid: vnid,
	}

	lis := &Listener{
		taps: map[uint32]Nic{},
		mtu:  1500,
		FDB:  fdb,
		conn: handler,
	}
	lis.taps[vnid] = tap

	lis.Start()
	tap.SetHandler(lis.Send)
	fdb.AddEntry(vnid, net.HardwareAddr{0xCC, 0x03, 0x04, 0xDC, 0x00, 0x10}, net.ParseIP("1.1.1.1"))

	frame, _ := constructPacket()
	frameLen := int64(len(frame))

	frames := []protocol.Packet{{VNID: vnid, Frame: frame[:]}}

	b.Run("Send", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			tap.Receive(frames)
			b.SetBytes(frameLen)
		}
	})
}

func constructPacket() ([]byte, error) {
	srcMAC := []byte{0xb2, 0x96, 0x81, 0x75, 0xb2, 0x11}
	dstMAC := []byte{0xe6, 0x15, 0x9c, 0x38, 0xf6, 0xe7}
	SrcIP := "127.0.0.1"
	DstIP := "8.8.8.8"
	SrcPort := 87654
	DstPort := 87654

	eth := &layers.Ethernet{
		SrcMAC:       net.HardwareAddr(srcMAC),
		DstMAC:       net.HardwareAddr(dstMAC),
		EthernetType: layers.EthernetTypeDot1Q,
	}
	vlan := &layers.Dot1Q{
		VLANIdentifier: 5,
		Type:           layers.EthernetTypeIPv4,
	}
	ip := &layers.IPv4{
		Version:  4,
		IHL:      5,
		TTL:      64,
		Id:       0,
		Protocol: layers.IPProtocolUDP,
		SrcIP:    net.ParseIP(SrcIP).To4(),
		DstIP:    net.ParseIP(DstIP).To4(),
	}
	udp := &layers.UDP{
		SrcPort: layers.UDPPort(SrcPort),
		DstPort: layers.UDPPort(DstPort),
	}
	udp.SetNetworkLayerForChecksum(ip)
	buffer := gopacket.NewSerializeBuffer()
	err := gopacket.SerializeLayers(buffer, gopacket.SerializeOptions{
		FixLengths:       true,
		ComputeChecksums: true,
	},
		eth, vlan, ip, udp,
		gopacket.Payload([]byte("123456789abcdefghi123456789abcdefghi")),
	)

	return buffer.Bytes(), err
}

func constructVPCPacket() ([]byte, error) {
	srcMAC := []byte{0xb2, 0x96, 0x81, 0x75, 0xb2, 0x11}
	dstMAC := []byte{0xe6, 0x15, 0x9c, 0x38, 0xf6, 0xe7}
	SrcIP := "127.0.0.1"
	DstIP := "8.8.8.8"
	SrcPort := 87654
	DstPort := 87654

	eth := &layers.Ethernet{
		SrcMAC:       net.HardwareAddr(srcMAC),
		DstMAC:       net.HardwareAddr(dstMAC),
		EthernetType: layers.EthernetTypeIPv4,
	}
	ip := &layers.IPv4{
		Version:  4,
		IHL:      5,
		TTL:      64,
		Id:       0,
		Protocol: 172,
		SrcIP:    net.ParseIP(SrcIP).To4(),
		DstIP:    net.ParseIP(DstIP).To4(),
	}
	udp := &layers.UDP{
		SrcPort: layers.UDPPort(SrcPort),
		DstPort: layers.UDPPort(DstPort),
	}
	udp.SetNetworkLayerForChecksum(ip)
	buffer := gopacket.NewSerializeBuffer()
	err := gopacket.SerializeLayers(buffer, gopacket.SerializeOptions{
		FixLengths:       true,
		ComputeChecksums: true,
	},
		eth, ip, udp,
		gopacket.Payload([]byte("123456789abcdefghi123456789")),
	)

	return buffer.Bytes(), err
}
