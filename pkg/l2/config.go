package l2

import "github.com/spf13/viper"

//vtepDev provides the iface for the vtep dev attachment from viper config
func vtepDev() string {
	return viper.GetString("vtepdev")
}

//bgpPeers list of BGP peers from viper config
func bgpPeers() []string {
	return viper.GetStringSlice("bgp_peers")
}
