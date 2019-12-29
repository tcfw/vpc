package tap

import (
	"fmt"
	"log"
	"net"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"

	"github.com/tcfw/vpc/pkg/l2/transport/tap/protocol"
)

const (
	icmpv6NeighborAdvertisementFlagsSolicated = 1 << 6
	icmpv6NeighborAdvertisementFlagsOverride  = 1 << 5
)

type icmpv6Layers struct {
	ethernet *layers.Ethernet
	vlan     *layers.Dot1Q
	ip       *layers.IPv6
	icmp     *layers.ICMPv6NeighborSolicitation
}

//icmp6NDPReduce proxies any ICMPv6 Neighbor Solicitation requests to the SDN and provides injects a neighbor advertisement if available
//NOTE: Router solicification is handled by the layer-3 agent router still
func (s *Listener) icmp6NDPReduce(packet *protocol.Packet) error {
	log.Printf("ICMPv6 NDP")
	log.Printf("%x", packet.Frame)

	if packet.Frame[58] != layers.ICMPv6TypeNeighborSolicitation {
		return fmt.Errorf("NDP type %d is not supported", packet.Frame[58])
	}

	reqLayers, err := s.decodeICMPv6Frame(packet.Frame)
	if err != nil {
		return fmt.Errorf("failed to decode frame: %s", err)
	}

	mac, _, err := s.sdn.LookupIP(packet.VNID, reqLayers.vlan.VLANIdentifier, reqLayers.icmp.TargetAddress)
	if err != nil {
		return fmt.Errorf("failed to find IP: %s", err)
	}

	resp, err := s.buildICMPv6NAResponse(reqLayers, mac)
	if err != nil {
		return fmt.Errorf("failed to build arp response: %s", err)
	}

	log.Printf("Sending ICMPv6 respone %x", resp)

	s.taps[packet.VNID].in <- &protocol.Packet{VNID: packet.VNID, Frame: resp}

	return nil
}

//decodeICMPv6Frame decodes dot1q, ipv6 and ICMPv6 layers and validates accordingly
func (s *Listener) decodeICMPv6Frame(frame []byte) (*icmpv6Layers, error) {
	eth := &layers.Ethernet{}
	vlan := &layers.Dot1Q{}
	ipv6 := &layers.IPv6{}
	icmp := &layers.ICMPv6{}
	icmpv6 := &layers.ICMPv6NeighborSolicitation{}

	parser := gopacket.NewDecodingLayerParser(layers.LayerTypeEthernet, eth, vlan, ipv6, icmp, icmpv6)

	respLayers := make([]gopacket.LayerType, 0)
	if err := parser.DecodeLayers(frame, &respLayers); err != nil {
		return nil, err
	}

	//RFC4861 hop limit should still be 255
	if ipv6.HopLimit != 255 {
		goto invalidFrame
	}

	//RFC4861 neighbor solicitations cannot be to multicast addrs
	if icmpv6.TargetAddress.IsMulticast() {
		goto invalidFrame
	}

	return &icmpv6Layers{ethernet: eth, vlan: vlan, ip: ipv6, icmp: icmpv6}, nil

invalidFrame:
	return nil, fmt.Errorf("Invalid ICMPv6")

}

func (s *Listener) buildICMPv6NAResponse(reqLayers *icmpv6Layers, mac net.HardwareAddr) ([]byte, error) {
	eth := &layers.Ethernet{
		SrcMAC:       mac,
		DstMAC:       reqLayers.ethernet.SrcMAC,
		EthernetType: layers.EthernetTypeDot1Q,
	}
	vlan := &layers.Dot1Q{
		VLANIdentifier: reqLayers.vlan.VLANIdentifier,
		Type:           layers.EthernetTypeIPv6,
	}
	ipv6 := &layers.IPv6{
		Version:    6,
		DstIP:      reqLayers.ip.SrcIP,
		SrcIP:      reqLayers.icmp.TargetAddress,
		HopLimit:   255,
		NextHeader: layers.IPProtocolICMPv6,
	}
	icmpv6 := &layers.ICMPv6{
		TypeCode: layers.CreateICMPv6TypeCode(layers.ICMPv6TypeNeighborAdvertisement, 0),
	}
	icmpv6.SetNetworkLayerForChecksum(ipv6)
	icmpv6NA := &layers.ICMPv6NeighborAdvertisement{
		Flags:         icmpv6NeighborAdvertisementFlagsSolicated | icmpv6NeighborAdvertisementFlagsOverride,
		TargetAddress: reqLayers.icmp.TargetAddress,
		Options: layers.ICMPv6Options{
			layers.ICMPv6Option{
				Type: layers.ICMPv6OptTargetAddress,
				Data: mac,
			},
		},
	}

	buf := gopacket.NewSerializeBuffer()
	err := gopacket.SerializeLayers(buf, gopacket.SerializeOptions{FixLengths: true, ComputeChecksums: true},
		eth,
		vlan,
		ipv6,
		icmpv6,
		icmpv6NA)

	return buf.Bytes(), err
}
