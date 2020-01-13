package vxlan

import (
	"math/big"
	"net"

	"github.com/tcfw/vpc/pkg/l2/transport/tap/protocol"
)

func hash(packet *protocol.Packet, rdst net.IP, mod int) int {
	i := big.NewInt(0)
	i.SetBytes(packet.Frame[:57]) //Shorten down to the average length of a IPv6 header
	i.Mod(i, big.NewInt(int64(mod)))
	return int(i.Uint64())
}
