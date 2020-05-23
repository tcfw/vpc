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

	//Construct the packet to send
	frame := constructPacket()
	frameLen := int64(len(frame))

	frames := []*protocol.Packet{&protocol.Packet{VNID: vnid, Frame: frame}}

	var n int64 = 0

	b.ResetTimer()

	b.Run("Send", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			tap.Receive(frames)
			n += frameLen
			b.SetBytes(frameLen)
		}
	})
}

func constructPacket() []byte {
	ethernetLayer := &layers.Ethernet{
		SrcMAC:       net.HardwareAddr{0xFF, 0xAA, 0xFA, 0xAA, 0xFF, 0xAA},
		DstMAC:       net.HardwareAddr{0xCC, 0x03, 0x04, 0xDC, 0x00, 0x10},
		EthernetType: layers.EthernetTypeDot1Q,
	}
	vlanLayer := &layers.Dot1Q{
		VLANIdentifier: 1,
	}
	ipLayer := &layers.IPv4{
		SrcIP: net.IP{127, 0, 0, 1},
		DstIP: net.IP{8, 8, 8, 8},
	}

	icmpLayer := &layers.ICMPv4{
		TypeCode: layers.ICMPv4TypeEchoRequest,
		Id:       1,
		Seq:      1,
	}

	buffer := gopacket.NewSerializeBuffer()
	gopacket.SerializeLayers(buffer, gopacket.SerializeOptions{},
		ethernetLayer,
		vlanLayer,
		ipLayer,
		icmpLayer,
		gopacket.Payload([]byte("123456789abcdefghijklmnopqrstuvwxyz")),
	)

	return buffer.Bytes()
}
