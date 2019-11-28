package l2

import "github.com/spf13/viper"

func vtepDev() string {
	return viper.GetString("vtepdev")
}

func bgpPeers() []string {
	return viper.GetStringSlice("bgp_peers")
}
