package l3

import (
	"github.com/juliengk/go-netfilter/iptables"
)

const (
	iptTFILTER = "filter"

	iptCINPUT   = "INPUT"
	iptCFORWARD = "FORWARD"
	iptCOUTPUT  = "OUTPUT"
)

//SetDefaultFWRules sets the DROP policy on input and output, allows forwarding
func (r *Router) SetDefaultFWRules() error {
	ipt4, _ := iptables.New(4, false)
	ipt6, _ := iptables.New(6, false)

	ipts := []*iptables.IPTables{ipt4, ipt6}

	//apply same rules for both ipv4 and ipv6
	for _, ipt := range ipts {
		// ipt.Policy(iptTFILTER, iptCINPUT, "DROP")
		ipt.Policy(iptTFILTER, iptCFORWARD, "ACCEPT")
		// ipt.Policy(iptTFILTER, iptCOUTPUT, "DROP")

		//Allow any loopback
		ipt.Append(iptTFILTER, iptCINPUT, "-i", "lo", "-j", "ACCEPT")
		ipt.Append(iptTFILTER, iptCOUTPUT, "-o", "lo", "-j", "ACCEPT")

		//Allow BGP outbound
		ipt.Append(iptTFILTER, iptCOUTPUT, "-p", "tcp", "--dport", "179", "-m", "conntrack", "--ctstate", "NEW,ESTABLISHED", "-j", "ACCEPT")
		ipt.Append(iptTFILTER, iptCINPUT, "-p", "tcp", "--sport", "179", "-m", "conntrack", "--ctstate", "ESTABLISHED", "-j", "ACCEPT")

		//Allow DHCP packets in/out
		ipt.Append(iptTFILTER, iptCINPUT, "-p", "udp", "--dport", "67:68", "--sport", "67:68", "-j", "ACCEPT")
		ipt.Append(iptTFILTER, iptCOUTPUT, "-p", "udp", "--dport", "67:68", "--sport", "67:68", "-j", "ACCEPT")
		//Block DHCP from public iface
		ipt.Append(iptTFILTER, iptCINPUT, "-i", "eth0", "-p", "udp", "--dport", "67:68", "--sport", "67:68", "-j", "DROP")
	}
	return nil
}
