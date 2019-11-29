package l2

import (
	"context"
	"fmt"
	"log"
	"net"

	"golang.org/x/sys/unix"

	"github.com/vishvananda/netlink"

	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/any"
	api "github.com/osrg/gobgp/api"
	gobgp "github.com/osrg/gobgp/pkg/server"
)

//BGPSpeak contains a small gobgp server to advertise communities and routes between subnets
type BGPSpeak struct {
	s     *gobgp.BgpServer
	vnis  map[uint32]uint32
	peers []string
	asn   uint32
	rID   net.IP
}

//NewBGPSpeak inits a new BGP server. rID must match VTEP pub ip
func NewBGPSpeak(rID net.IP, peers []string) (*BGPSpeak, error) {
	srv := &BGPSpeak{
		s:     gobgp.NewBgpServer(),
		vnis:  map[uint32]uint32{},
		peers: peers,
		asn:   65000,
		rID:   rID,
	}
	return srv, nil
}

//Start begins the bgp session and start advertising
func (rbgp *BGPSpeak) Start() error {
	go rbgp.s.Serve()

	ctx := context.Background()

	if err := rbgp.s.StartBgp(ctx, &api.StartBgpRequest{
		Global: &api.Global{
			As:         rbgp.asn,
			RouterId:   rbgp.rID.String(),
			ListenPort: -1,
		},
	}); err != nil {
		return err
	}

	for _, peer := range rbgp.peers {
		if err := rbgp.AddPeer(peer); err != nil {
			return err
		}
	}

	return nil
}

//AddPeer adds a bgp neighbor with l2vpn-evpn capabilities
func (rbgp *BGPSpeak) AddPeer(addr string) error {
	ctx := context.Background()

	n := &api.Peer{
		// ApplyPolicy: &api.ApplyPolicy{
		// 	InPolicy: &api.PolicyAssignment{
		// 		// default import only vni
		// 		Policies: []*api.Policy{
		// 			&api.Policy{
		// 				// allow only with matching community
		// 				Statements: []*api.Statement{
		// 					&api.Statement{
		// 						Conditions: &api.Conditions{
		// 							AsPathSet: &api.MatchSet{MatchType: api.MatchType_ANY, Name: "ASN"},
		// 						},
		// 						Actions: &api.Actions{RouteAction: api.RouteAction_ACCEPT},
		// 					},
		// 				},
		// 			},
		// 		},
		// 		DefaultAction: api.RouteAction_REJECT,
		// 	},
		// 	ExportPolicy: &api.PolicyAssignment{
		// 		//default export all accept
		// 		DefaultAction: api.RouteAction_ACCEPT,
		// 	},
		// },
		Conf: &api.PeerConf{
			NeighborAddress: addr,
			PeerAs:          65000,
		},
		// RouteReflector: &api.RouteReflector{
		// 	RouteReflectorClient:    true,
		// 	RouteReflectorClusterId: rbgp.rID.String(),
		// },
		AfiSafis: []*api.AfiSafi{
			{
				Config: &api.AfiSafiConfig{
					Family: &api.Family{
						Afi:  api.Family_AFI_L2VPN,
						Safi: api.Family_SAFI_EVPN,
					},
					Enabled: true,
				},
			},
		},
		Timers: &api.Timers{
			Config: &api.TimersConfig{
				ConnectRetry:      5,
				HoldTime:          15,
				KeepaliveInterval: 3,
			},
		},
	}

	return rbgp.s.AddPeer(ctx, &api.AddPeerRequest{
		Peer: n,
	})
}

//RegisterVTEP adds a Type-3 EVPN route to the VTEP
func (rbgp *BGPSpeak) RegisterVTEP(vni uint32) error {
	rd, _ := ptypes.MarshalAny(&api.RouteDistinguisherIPAddress{Admin: rbgp.rID.String(), Assigned: vni})

	nlri, _ := ptypes.MarshalAny(&api.EVPNEthernetAutoDiscoveryRoute{
		Rd:          rd,
		EthernetTag: 0, //TODO(tcfw) use subnet vlan
		Esi:         &api.EthernetSegmentIdentifier{Type: 0, Value: []byte("single-homed")},
		Label:       vni,
	})

	if _, ok := rbgp.vnis[vni]; !ok {
		rbgp.vnis[vni] = vni
	}

	rbgp.updateASNPolicy()

	return rbgp.addL2VPNEVPNNLRIPath(nlri)
}

//addL2VPNEVPNNLRIPath executes the add path with given NLRI
func (rbgp *BGPSpeak) addL2VPNEVPNNLRIPath(nlri *any.Any) error {
	family := &api.Family{
		Afi:  api.Family_AFI_L2VPN,
		Safi: api.Family_SAFI_EVPN,
	}

	_, err := rbgp.s.AddPath(context.Background(), &api.AddPathRequest{
		Path: &api.Path{
			Family: family,
			Nlri:   nlri,
			Pattrs: rbgp.pathPAttrs(),
		},
	})

	return err
}

//pathPAttrs default path attrs used in BGP paths
func (rbgp *BGPSpeak) pathPAttrs() []*any.Any {
	attrOrigin, _ := ptypes.MarshalAny(&api.OriginAttribute{
		Origin: 0,
	})

	attrNextHop, _ := ptypes.MarshalAny(&api.NextHopAttribute{
		NextHop: rbgp.rID.String(),
	})

	return []*any.Any{attrOrigin, attrNextHop}
}

//DeregisterVTEP removes the Type-3 EVPN route to the VTEP
func (rbgp *BGPSpeak) DeregisterVTEP(vni uint32) error {
	rd, _ := ptypes.MarshalAny(&api.RouteDistinguisherIPAddress{Admin: rbgp.rID.String(), Assigned: vni})

	nlri, _ := ptypes.MarshalAny(&api.EVPNEthernetAutoDiscoveryRoute{
		Rd:          rd,
		EthernetTag: 0,
		Esi:         &api.EthernetSegmentIdentifier{Type: 0, Value: []byte("single-homed")},
		Label:       vni,
	})

	if _, ok := rbgp.vnis[vni]; !ok {
		delete(rbgp.vnis, vni)
	}

	rbgp.updateASNPolicy()

	return rbgp.delL2VPNEVPNNLRIPath(nlri)
}

//delL2VPNEVPNNLRIPath executes the delete path of given NLRI
func (rbgp *BGPSpeak) delL2VPNEVPNNLRIPath(nlri *any.Any) error {
	family := &api.Family{
		Afi:  api.Family_AFI_L2VPN,
		Safi: api.Family_SAFI_EVPN,
	}

	return rbgp.s.DeletePath(context.Background(), &api.DeletePathRequest{
		Path: &api.Path{
			Family: family,
			Nlri:   nlri,
			Pattrs: rbgp.pathPAttrs(),
		},
	})
}

//updateASNPolicy updates the defined ASN policy with all known VNIs
func (rbgp *BGPSpeak) updateASNPolicy() {

	list := []string{}

	for _, vni := range rbgp.vnis {
		list = append(list, fmt.Sprintf("%d", 4200000000+vni))
	}

	rbgp.s.AddDefinedSet(context.Background(), &api.AddDefinedSetRequest{
		DefinedSet: &api.DefinedSet{
			Name:     "ASN",
			List:     list,
			Prefixes: []*api.Prefix{{}},
		},
	})
}

//RegisterMacIP adds a Type-2 EVPN route to a local mac/ip
func (rbgp *BGPSpeak) RegisterMacIP(vni uint32, vlan uint32, mac net.HardwareAddr, ip net.IP) error {
	log.Printf("REGISTER MAC: %+v %+v %+v %+v\n", vni, vlan, mac, ip)

	rd, _ := ptypes.MarshalAny(&api.RouteDistinguisherIPAddress{Admin: rbgp.rID.String(), Assigned: vni})

	nlri, _ := ptypes.MarshalAny(&api.EVPNMACIPAdvertisementRoute{
		Rd:          rd,
		EthernetTag: vlan,
		Esi:         &api.EthernetSegmentIdentifier{Type: 0, Value: []byte("single-homed")},
		Labels:      []uint32{vni},
		MacAddress:  mac.String(),
		IpAddress:   ip.String(),
	})

	return rbgp.addL2VPNEVPNNLRIPath(nlri)
}

//DeregisterMacIP removes the Type-2 EVPN route to a local mac/ip
func (rbgp *BGPSpeak) DeregisterMacIP(vni uint32, vlan uint32, mac net.HardwareAddr, ip net.IP) error {
	log.Printf("DEREG MAC: %+v %+v %+v %+v\n", vni, vlan, mac, ip)

	rd, _ := ptypes.MarshalAny(&api.RouteDistinguisherIPAddress{Admin: rbgp.rID.String(), Assigned: vni})

	nlri, _ := ptypes.MarshalAny(&api.EVPNMACIPAdvertisementRoute{
		Rd:          rd,
		EthernetTag: vlan,
		Esi:         &api.EthernetSegmentIdentifier{Type: 0, Value: []byte("single-homed")},
		Labels:      []uint32{vni},
		MacAddress:  mac.String(),
		IpAddress:   ip.String(),
	})

	return rbgp.delL2VPNEVPNNLRIPath(nlri)
}

//LookupMac resolve l2 requests from bridge to forward to the correct VTEP
func (rbgp *BGPSpeak) LookupMac(vni uint32, mac net.HardwareAddr) (*net.IP, error) {
	log.Printf("LOOKUP MAC: %+v %+v\n", vni, mac)

	if err := rbgp.s.ListPath(context.Background(), &api.ListPathRequest{
		TableType: api.TableType_ADJ_IN,
		Family: &api.Family{
			Afi:  api.Family_AFI_L2VPN,
			Safi: api.Family_SAFI_EVPN,
		},
	}, func(dst *api.Destination) {
		log.Printf("%+v", dst)
	}); err != nil {
		return nil, err
	}

	return nil, nil
}

//ApplyVTEPAD adds all known vteps for a given VNI to the VTEP FDB
func (rbgp *BGPSpeak) ApplyVTEPAD(vni uint32) error {
	if err := rbgp.s.ListPath(context.Background(), &api.ListPathRequest{
		TableType: api.TableType_GLOBAL,
		Family: &api.Family{
			Afi:  api.Family_AFI_L2VPN,
			Safi: api.Family_SAFI_EVPN,
		},
	}, func(dst *api.Destination) {
		for _, path := range dst.Paths {
			//Ingore our own paths
			if path.GetSourceId() != rbgp.rID.String() && path.SourceId != "<nil>" {
				// log.Printf("Got path: %+v", path.Nlri)

				switch true {
				case ptypes.Is(path.Nlri, &api.EVPNEthernetAutoDiscoveryRoute{}):
					eadRoute := &api.EVPNEthernetAutoDiscoveryRoute{}
					ptypes.UnmarshalAny(path.Nlri, eadRoute)

					rd := &api.RouteDistinguisherIPAddress{}
					ptypes.UnmarshalAny(eadRoute.Rd, rd)

					rdIP := net.ParseIP(rd.Admin)

					link, err := netlink.LinkByName(fmt.Sprintf("t-%d", vni))
					if err != nil {
						log.Println("failed to find VTEP iface:", err)
						continue
					}

					hwaddr, _ := net.ParseMAC("00:00:00:00:00:00")

					neigh := &netlink.Neigh{
						LinkIndex:    link.Attrs().Index,
						HardwareAddr: hwaddr,
						Family:       unix.AF_BRIDGE,
						IP:           rdIP,
						State:        netlink.NUD_REACHABLE,
						Flags:        netlink.NTF_SELF | netlink.NTF_PROXY,
					}

					log.Printf("Appending neigh: %+v", neigh)

					if err := netlink.NeighAppend(neigh); err != nil {
						log.Println("failed to add VTEP to FDB:", err)
					}
				}
			}
		}

	}); err != nil {
		return err
	}

	return nil
}
