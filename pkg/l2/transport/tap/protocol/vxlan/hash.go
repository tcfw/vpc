package vxlan

import (
	"math"
	"math/big"
	"net"

	"github.com/tcfw/vpc/pkg/l2/transport/tap/protocol"
)

func hash(packet *protocol.Packet, rdst net.IP, mod int) int {
	i := big.NewInt(0)
	n := uint(math.Min(57, float64(len(packet.Frame))))

	i.SetBytes(packet.Frame[:n]) //Shorten down to the average length of a IPv6 header
	i.Mod(i, big.NewInt(int64(mod)))
	return int(i.Uint64())
}
