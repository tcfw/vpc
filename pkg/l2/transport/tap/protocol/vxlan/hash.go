package vxlan

import (
	"hash/crc32"
	"net"

	"github.com/tcfw/vpc/pkg/l2/transport/tap/protocol"
)

func hash(packet *protocol.Packet, rdst net.IP, mod int) int {
	n := len(packet.Frame)
	if n > 57 {
		n = 57
	}

	i := crc32.ChecksumIEEE(packet.Frame[:n])

	return int(i % uint32(mod))
}
