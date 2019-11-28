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
	ipt, _ := iptables.New(4, false)

	ipt.Policy(iptTFILTER, iptCINPUT, "DROP")
	ipt.Policy(iptTFILTER, iptCOUTPUT, "DROP")

	return nil
}
