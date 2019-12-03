package l2

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

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

//RegisterEP adds a Type-3 EVPN route to the VTEP
func (rbgp *BGPSpeak) RegisterEP(vni uint32) error {
	rd, _ := ptypes.MarshalAny(&api.RouteDistinguisherIPAddress{Admin: rbgp.rID.String(), Assigned: vni})

	nlri, _ := ptypes.MarshalAny(&api.EVPNEthernetAutoDiscoveryRoute{
		Rd:          rd,
		EthernetTag: 0,
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

//DeregisterEP removes the Type-3 EVPN route to the VTEP
func (rbgp *BGPSpeak) DeregisterEP(vni uint32) error {
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

//LookupIP resolve l3 requests from bridge
func (rbgp *BGPSpeak) LookupIP(vni uint32, vlan uint16, ip net.IP) (net.HardwareAddr, net.IP, error) {
	log.Printf("LOOKUP IP: %+v %d, %+v\n", vni, vlan, ip)

	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()

	var hwaddr net.HardwareAddr
	var gw net.IP

	found := make(chan struct{})

	if err := rbgp.s.ListPath(ctx, &api.ListPathRequest{
		TableType: api.TableType_GLOBAL,
		Family: &api.Family{
			Afi:  api.Family_AFI_L2VPN,
			Safi: api.Family_SAFI_EVPN,
		},
	}, func(dst *api.Destination) {
		for _, path := range dst.Paths {
			if path.GetSourceId() != rbgp.rID.String() && path.SourceId != "<nil>" {
				switch true {
				case ptypes.Is(path.Nlri, &api.EVPNMACIPAdvertisementRoute{}):
					macIP := &api.EVPNMACIPAdvertisementRoute{}
					ptypes.UnmarshalAny(path.Nlri, macIP)

					if macIP.Labels[0] != vni || macIP.EthernetTag != uint32(vlan) {
						continue
					}

					rd := &api.RouteDistinguisherIPAddress{}
					ptypes.UnmarshalAny(macIP.Rd, rd)

					rdIP := net.ParseIP(rd.Admin)

					if macIP.IpAddress == ip.String() {
						hwaddr, _ = net.ParseMAC(macIP.MacAddress)
						gw = rdIP

						found <- struct{}{}
						return
					}
				}
			}
		}
	}); err != nil {
		return nil, nil, err
	}

	select {
	case <-found:
		break
	case <-ctx.Done():
		return nil, nil, fmt.Errorf("timed out")
	}

	return hwaddr, gw, nil
}

//LookupMac resolves l2 requets for VTEP forwarding
func (rbgp *BGPSpeak) LookupMac(vni uint32, mac net.HardwareAddr) (net.IP, error) {
	log.Printf("LOOKUP MAC: %d %s\n", vni, mac)

	timeout := 4 * time.Second

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	var gw net.IP

	found := make(chan struct{})

	go func() {
		rbgp.s.ListPath(ctx, &api.ListPathRequest{
			TableType: api.TableType_GLOBAL,
			Family: &api.Family{
				Afi:  api.Family_AFI_L2VPN,
				Safi: api.Family_SAFI_EVPN,
			},
		}, func(dst *api.Destination) {
			for _, path := range dst.Paths {
				if path.GetSourceId() != rbgp.rID.String() && path.SourceId != "<nil>" {
					switch true {
					case ptypes.Is(path.Nlri, &api.EVPNMACIPAdvertisementRoute{}):
						macIP := &api.EVPNMACIPAdvertisementRoute{}
						ptypes.UnmarshalAny(path.Nlri, macIP)

						if macIP.Labels[0] != vni { //|| macIP.EthernetTag != uint32(vpcID)
							continue
						}

						rd := &api.RouteDistinguisherIPAddress{}
						ptypes.UnmarshalAny(macIP.Rd, rd)

						rdIP := net.ParseIP(rd.Admin)

						if macIP.MacAddress == mac.String() {
							gw = rdIP

							found <- struct{}{}
							return
						}
					}
				}
			}
		})
	}()

	select {
	case <-found:
		break
	case <-time.After(timeout):
		return nil, fmt.Errorf("timed out")
	}

	return gw, nil
}

//BroadcastEndpoints adds all known vteps for a given VNI to the VTEP FDB for broadcast traffic
func (rbgp *BGPSpeak) BroadcastEndpoints(vni uint32) ([]net.IP, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()

	endpoints := []net.IP{}
	var mu sync.Mutex

	if err := rbgp.s.ListPath(ctx, &api.ListPathRequest{
		TableType: api.TableType_GLOBAL,
		Family: &api.Family{
			Afi:  api.Family_AFI_L2VPN,
			Safi: api.Family_SAFI_EVPN,
		},
	}, func(dst *api.Destination) {
		for _, path := range dst.Paths {
			//Ingore our own paths
			if path.GetSourceId() != rbgp.rID.String() && path.SourceId != "<nil>" {
				switch true {
				case ptypes.Is(path.Nlri, &api.EVPNEthernetAutoDiscoveryRoute{}):
					eadRoute := &api.EVPNEthernetAutoDiscoveryRoute{}
					ptypes.UnmarshalAny(path.Nlri, eadRoute)

					if eadRoute.Label != vni {
						continue
					}

					rd := &api.RouteDistinguisherIPAddress{}
					ptypes.UnmarshalAny(eadRoute.Rd, rd)

					rdIP := net.ParseIP(rd.Admin)

					mu.Lock()
					defer mu.Unlock()

					endpoints = append(endpoints, rdIP)
				}
			}
		}

	}); err != nil {
		return nil, err
	}

	<-ctx.Done()

	return endpoints, nil
}
