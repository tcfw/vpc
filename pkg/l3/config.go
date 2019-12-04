package l3

import (
	"net"

	"github.com/spf13/viper"
)

func configPublicBridge() string {
	return viper.GetString("br")
}

func configDHCP() bool {
	return viper.GetBool("dhcp")
}

func configNAT() bool {
	return viper.GetBool("nat")
}

func configDHCPDNS() []net.IP {
	ips := []net.IP{}

	sDNS := viper.GetStringSlice("dns")

	for _, ip := range sDNS {
		ips = append(ips, net.ParseIP(ip))
	}
	return ips
}

func configBGPPeers() []string {
	return viper.GetStringSlice("bgp")
}

func configSubnets() []string {
	return viper.GetStringSlice("ips")
}
