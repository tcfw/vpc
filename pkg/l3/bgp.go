package l3

import (
	"context"
	"fmt"
	"log"
	"net"
	"os/exec"

	"github.com/vishvananda/netns"

	"google.golang.org/grpc"

	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/any"
	api "github.com/osrg/gobgp/api"
)

const (
	asPrefix = uint32(4200000000)
)

//RouterBGP contains a small gobgp server to advertise communities and routes between subnets
type RouterBGP struct {
	s       api.GobgpApiClient
	sConn   *grpc.ClientConn
	pubIP   *net.IP
	subnets []*net.IPNet
	vni     uint32
	peers   []string

	proc *exec.Cmd
}

//NewRouterBGP inits a new BGP server
func NewRouterBGP(pubIP net.IP, vni uint32, peers []string) (*RouterBGP, error) {
	srv := &RouterBGP{
		pubIP:   &pubIP,
		subnets: []*net.IPNet{},
		vni:     vni,
		peers:   peers,
	}

	return srv, nil
}

func (rbgp *RouterBGP) startProc() error {
	cmd := exec.Command("/home/tom/go/bin/gobgpd")

	if err := cmd.Start(); err != nil {
		log.Printf("Failed to exec gobgpd: %s", err)
		return err
	}

	rbgp.proc = cmd

	return nil
}

//Start begins the bgp session and start advertising
func (rbgp *RouterBGP) Start(ns netns.NsHandle) error {
	rbgp.startProc()

	conn, err := grpc.Dial("127.0.0.1:50051", grpc.WithInsecure(), grpc.WithBlock(), grpc.WithContextDialer(func(ctx context.Context, addr string) (net.Conn, error) {
		netns.Set(ns)
		var d net.Dialer
		return d.DialContext(ctx, "tcp", addr)
	}))
	if err != nil {
		return err
	}
	rbgp.sConn = conn
	rbgp.s = api.NewGobgpApiClient(rbgp.sConn)

	ctx := context.Background()

	_, err = rbgp.s.StartBgp(ctx, &api.StartBgpRequest{
		Global: &api.Global{
			ListenAddresses: []string{rbgp.pubIP.String()},
			As:              rbgp.ASN(),
			RouterId:        rbgp.pubIP.String(),
			ListenPort:      -1, // gobgp won't listen on tcp:179
		},
	})
	if err != nil {
		return err
	}

	// if _, err := rbgp.s.MonitorPeer(ctx, &api.MonitorPeerRequest{}, func(p *api.Peer) { rbgp.peerChange(p) }); err != nil {
	// 	return err
	// }

	rbgp.s.AddDefinedSet(ctx, &api.AddDefinedSetRequest{
		DefinedSet: &api.DefinedSet{
			Name:     "ASN",
			List:     []string{fmt.Sprintf("%d", rbgp.ASN())},
			Prefixes: []*api.Prefix{&api.Prefix{}},
		},
	})

	for _, peer := range rbgp.peers {
		n := &api.Peer{
			ApplyPolicy: &api.ApplyPolicy{
				InPolicy: &api.PolicyAssignment{
					// default import only vni
					Policies: []*api.Policy{
						&api.Policy{
							// allow only with matching community
							Statements: []*api.Statement{
								&api.Statement{
									Conditions: &api.Conditions{
										AsPathSet: &api.MatchSet{MatchType: api.MatchType_ANY, Name: "ASN"},
									},
									Actions: &api.Actions{RouteAction: api.RouteAction_ACCEPT},
								},
							},
						},
					},
					DefaultAction: api.RouteAction_REJECT,
				},
				ExportPolicy: &api.PolicyAssignment{
					//default export all accept
					DefaultAction: api.RouteAction_ACCEPT,
				},
			},
			Conf: &api.PeerConf{
				NeighborAddress: peer,
				PeerAs:          65000,
			},
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

		if _, err := rbgp.s.AddPeer(ctx, &api.AddPeerRequest{
			Peer: n,
		}); err != nil {
			return err
		}
	}

	return nil
}

//AdvertSubnet adds a subnet to the bgp server within the vpc community
func (rbgp *RouterBGP) AdvertSubnet(subnet *net.IPNet) error {
	rd, _ := ptypes.MarshalAny(&api.RouteDistinguisherIPAddress{Admin: rbgp.pubIP.String(), Assigned: rbgp.vni})

	maskSize, _ := subnet.Mask.Size()
	nlri, _ := ptypes.MarshalAny(&api.EVPNIPPrefixRoute{
		Rd:          rd,
		EthernetTag: 0, //TODO(tcfw) use subnet vlan
		Esi:         &api.EthernetSegmentIdentifier{Type: 0, Value: []byte("single-homed")},
		IpPrefix:    subnet.IP.String(),
		IpPrefixLen: uint32(maskSize),
		Label:       rbgp.vni,
		GwAddress:   rbgp.pubIP.String(),
	})

	//L2VPN EVPN
	family := &api.Family{
		Afi:  api.Family_AFI_L2VPN,
		Safi: api.Family_SAFI_EVPN,
	}

	attrOrigin, _ := ptypes.MarshalAny(&api.OriginAttribute{
		Origin: 0,
	})
	attrNextHop, _ := ptypes.MarshalAny(&api.NextHopAttribute{
		NextHop: rbgp.pubIP.String(),
	})

	attrAsPath, _ := ptypes.MarshalAny(&api.AsPathAttribute{
		Segments: []*api.AsSegment{
			&api.AsSegment{
				Type:    0,
				Numbers: []uint32{rbgp.ASN()},
			},
		},
	})

	_, err := rbgp.s.AddPath(context.Background(), &api.AddPathRequest{
		Path: &api.Path{
			Family: family,
			Nlri:   nlri,
			Pattrs: []*any.Any{attrOrigin, attrNextHop, attrAsPath},
		},
	})

	return err
}

//ASN provides the privte ASN of the vpc
func (rbgp *RouterBGP) ASN() uint32 {
	return asPrefix + rbgp.vni
}

func (rbgp *RouterBGP) peerChange(p *api.Peer) {
	// log.Printf("%v\n", p)
}
