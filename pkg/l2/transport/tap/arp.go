package tap

import (
	"fmt"
	"log"
	"net"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/tcfw/vpc/pkg/l2/transport/tap/protocol"
)

//arpReduce proxies any ARP request to the SDN and provides injects a reply if available
func (s *Listener) arpReduce(packet *protocol.Packet) error {
	defer func() {
		//@TODO fix the panics caused by misreading ARP requests
		if r := recover(); r != nil {
			log.Println("paniced in arp reduce", r)
		}
	}()

	if s.sdn == nil {
		return fmt.Errorf("no SDN attached")
	}

	_, arpRequest, err := s.decodeARPFrame(packet.Frame)
	if err != nil {
		return fmt.Errorf("failed to decode ARP request: %s", err)
	}

	mac, _, err := s.sdn.LookupIP(packet.VNID, 1, net.IP(arpRequest.DstProtAddress))
	if err != nil {
		return fmt.Errorf("failed to find IP for %s: %s", net.IP(arpRequest.DstProtAddress).String(), err)
	} else if len(mac) == 0 || mac == nil {
		return fmt.Errorf("invalid mac returned from SDN")
	}

	resp, err := s.buildARPResponse(1, arpRequest, mac)
	if err != nil {
		return fmt.Errorf("failed to build arp response: %s", err)
	}

	log.Printf("ARP REPLY: %s - % X", mac, resp)
	_, err = s.taps[packet.VNID].Write(resp)

	return err
}

//decodeARPFrame decodes dot1q and arp layers and validates accordingly
func (s *Listener) decodeARPFrame(frame []byte) (*layers.Dot1Q, *layers.ARP, error) {
	packetData := gopacket.NewPacket(frame, layers.LayerTypeEthernet, gopacket.Default)

	// vlan := &layers.Dot1Q{}
	// dot1qLayer := packetData.Layer(layers.LayerTypeDot1Q)
	// if err := vlan.DecodeFromBytes(dot1qLayer.LayerContents(), gopacket.NilDecodeFeedback); err != nil {
	// 	return nil, nil, fmt.Errorf("failed to decode dot1q frame: %s", err)
	// }

	arp := &layers.ARP{}
	arpLayer := packetData.Layer(layers.LayerTypeARP)
	if err := arp.DecodeFromBytes(arpLayer.LayerContents(), gopacket.NilDecodeFeedback); err != nil {
		return nil, nil, fmt.Errorf("failed to decode arp frame: %s", err)
	}

	if arp.Protocol != layers.EthernetTypeIPv4 {
		return nil, nil, fmt.Errorf("unsupported addr type: %s", arp.AddrType)
	}

	if arp.Operation != layers.ARPRequest {
		return nil, nil, fmt.Errorf("unsupported ARP type: %d", arp.Operation)
	}

	if arp.HwAddressSize != 6 || arp.ProtAddressSize != 4 {
		return nil, nil, fmt.Errorf("unsupported address length: %d, %d", arp.HwAddressSize, arp.ProtAddressSize)
	}

	return nil, arp, nil
}

//buildARPResponse uses the original arp request to create an arp response
func (s *Listener) buildARPResponse(vlanID uint16, arpRequest *layers.ARP, mac net.HardwareAddr) ([]byte, error) {
	eth := &layers.Ethernet{
		SrcMAC:       mac,
		DstMAC:       arpRequest.SourceHwAddress,
		EthernetType: layers.EthernetTypeARP,
	}
	// vlan := &layers.Dot1Q{
	// 	VLANIdentifier: vlanID,
	// 	Type:           layers.EthernetTypeARP,
	// }
	arp := &layers.ARP{
		AddrType:          layers.LinkTypeEthernet,
		Protocol:          layers.EthernetTypeIPv4,
		Operation:         layers.ARPReply,
		HwAddressSize:     6, //48-bit MAC
		ProtAddressSize:   4, //IPv4 (4 bytes)
		SourceHwAddress:   mac,
		SourceProtAddress: arpRequest.DstProtAddress,
		DstHwAddress:      arpRequest.SourceHwAddress,
		DstProtAddress:    arpRequest.SourceProtAddress,
	}

	buf := gopacket.NewSerializeBuffer()
	err := gopacket.SerializeLayers(buf, gopacket.SerializeOptions{FixLengths: true, ComputeChecksums: true},
		eth,
		// vlan,
		arp)

	return buf.Bytes(), err
}
