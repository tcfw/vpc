package l2

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"golang.org/x/sys/unix"

	l2API "github.com/tcfw/vpc/pkg/api/v1/l2"
	"github.com/tcfw/vpc/pkg/l2/controller"
	sdnController "github.com/tcfw/vpc/pkg/l2/controller/bgp"
	"github.com/tcfw/vpc/pkg/l2/transport"

	transportTap "github.com/tcfw/vpc/pkg/l2/transport/tap"
	// transportVTEP "github.com/tcfw/vpc/pkg/l2/transport/vtep"

	"github.com/vishvananda/netlink"
	"google.golang.org/grpc"
)

const (
	mtu = 2000
)

//Serve start the GRPC server
func Serve(port uint) {
	lis, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()

	srv, err := NewServer()
	if err != nil {
		log.Fatalf("failed to start vtep: %s", err)
	}

	go srv.Gc()
	go srv.SDN()
	go func() {
		if err := srv.transport.Start(); err != nil {
			log.Printf("failed to start transport: %s\n", err)
		}
	}()

	l2API.RegisterL2ServiceServer(grpcServer, srv)
	log.Println("Starting gRPC server")
	go grpcServer.Serve(lis)

	killSignal := make(chan os.Signal, 1)
	signal.Notify(killSignal, syscall.SIGINT, syscall.SIGTERM)

	<-killSignal

	srv.sdn.Stop()
}

//Server l2 API server
type Server struct {
	m         sync.Mutex
	watches   []chan l2API.StackChange
	stacks    map[int32]*Stack
	sdn       controller.Controller
	transport transport.Transport
}

//NewServer creates a new server instance
func NewServer() (*Server, error) {
	transport, err := transportTap.NewListener()
	if err != nil {
		return nil, err
	}

	// pubLink, err := netlink.LinkByName(vtepDev())
	// if err != nil {
	// 	return nil, fmt.Errorf("faild to init VTEP transport: %s", err)
	// }

	// transport := transportVTEP.NewVTEPTransport(pubLink)

	srv := &Server{
		watches:   []chan l2API.StackChange{},
		stacks:    map[int32]*Stack{},
		transport: transport,
	}

	srv.transport.SetMTU(mtu)

	return srv, nil
}

//SDN starts the SDN controller to advertise type-2 and type-3 routes
func (s *Server) SDN() {
	if s.sdn != nil {
		return
	}

	vtepdev, err := netlink.LinkByName(vtepDev())
	if err != nil {
		log.Fatalf("Failed to find to vtep dev: %s", err)
	}

	ip4s, err := netlink.AddrList(vtepdev, unix.AF_INET)
	if err != nil {
		log.Fatalln(err)
	}
	ip6s, err := netlink.AddrList(vtepdev, unix.AF_INET6)
	if err != nil {
		log.Fatalln(err)
	}

	var rID net.IP

	addrs := append(ip6s, ip4s...)
	for _, addr := range addrs {
		if !addr.IP.IsLoopback() && addr.IP.IsGlobalUnicast() {
			rID = addr.IP
			break
		}
	}

	s.sdn, err = sdnController.NewEVPNController(rID, bgpPeers())
	if err != nil {
		log.Fatalf("Failed to init BGP: %s", err)
	}

	if err := s.sdn.Start(); err != nil {
		log.Fatalf("Failed to start BGP: %s", err)
	}

	if err := s.transport.SetSDN(s.sdn); err != nil {
		log.Printf("failed to set transport SDN ref: %s", err)
	}

}

//AddStack creates a new VPC stack
func (s *Server) AddStack(ctx context.Context, req *l2API.StackRequest) (*l2API.StackResponse, error) {
	if req.VpcId > 16777215 {
		return nil, fmt.Errorf("VPC ID out of range")
	}

	vpc := uint32(req.VpcId)

	s.m.Lock()
	defer func() {
		s.m.Unlock()
	}()

	stack, err := createStack(req.VpcId)
	if err != nil {
		return nil, fmt.Errorf("failed to create stack: %s", err)
	}

	if err := s.transport.AddEP(vpc, stack.Bridge); err != nil {
		return nil, fmt.Errorf("failed to start tap: %s", err)
	}

	missCh, _ := s.transport.ForwardingMiss(vpc)
	go s.HandleMisses(vpc, missCh)

	if err := s.sdn.RegisterEP(vpc); err != nil {
		defer deleteStack(stack)
		return nil, fmt.Errorf("failed to register endpoint: %s", err)
	}

	status := s.getStackStatus(stack)

	s.logChange(&l2API.StackChange{
		VpcId:  req.VpcId,
		Action: "created",
		Status: status,
	})

	resp := &l2API.StackResponse{
		Stack:  formatToAPIStack(stack),
		Status: status,
	}

	s.stacks[req.VpcId] = stack

	return resp, nil
}

//GetStack finds and returns an existing VPC stack
func (s *Server) GetStack(ctx context.Context, req *l2API.StackRequest) (*l2API.StackResponse, error) {
	if req.VpcId > 16777215 {
		return nil, fmt.Errorf("VPC ID out of range")
	}

	stack, err := s.fetchStack(req.VpcId)
	if err != nil {
		return nil, err
	}

	if stack.Bridge == nil {
		return nil, fmt.Errorf("Stack not created")
	}

	resp := &l2API.StackResponse{
		Stack:  formatToAPIStack(stack),
		Status: s.getStackStatus(stack),
	}

	return resp, nil
}

//StackStatus returns just the current status of devices in stack
func (s *Server) StackStatus(ctx context.Context, req *l2API.StackRequest) (*l2API.StackStatusResponse, error) {
	if req.VpcId > 16777215 {
		return nil, fmt.Errorf("VPC ID out of range")
	}

	stack, err := s.fetchStack(req.VpcId)
	if err != nil {
		return nil, err
	}

	return s.getStackStatus(stack), nil
}

//DeleteStack deletes all devices in the stack
func (s *Server) DeleteStack(ctx context.Context, req *l2API.StackRequest) (*l2API.Empty, error) {
	if req.VpcId > 16777215 {
		return nil, fmt.Errorf("VPC ID out of range")
	}

	stack, ok := s.stacks[req.VpcId]

	var err error

	if !ok {
		stack, err = getStack(req.VpcId)
		if err != nil {
			return nil, err
		}
	}

	s.m.Lock()
	defer s.m.Unlock()

	s.transport.DelEP(uint32(stack.VPCID))

	err = deleteStack(stack)

	s.logChange(&l2API.StackChange{
		VpcId:  req.VpcId,
		Action: "deleted",
	})

	delete(s.stacks, req.VpcId)

	return nil, err
}

//WatchStacks monitors for changes in the stacks
func (s *Server) WatchStacks(_ *l2API.Empty, stream l2API.L2Service_WatchStacksServer) error {
	ch := make(chan l2API.StackChange)

	s.m.Lock()
	s.watches = append(s.watches, ch)
	s.m.Unlock()

	defer func() {
		close(ch)
	}()

	for {
		change, ok := <-ch
		if ok == false {
			break
		}
		if err := stream.Send(&change); err != nil {
			return err
		}
	}

	return nil
}

//AddNIC Add a new NIC to a VPC linux bridge
func (s *Server) AddNIC(ctx context.Context, req *l2API.NicRequest) (*l2API.Nic, error) {
	if err := s.validateAddNicRequest(req); err != nil {
		return nil, err
	}

	stack, err := s.fetchStack(req.VpcId)
	if err != nil {
		return nil, err
	}

	ok, err := HasNIC(stack, req.Id)
	if err != nil {
		return nil, fmt.Errorf("Failed to create nic: %s", err)
	} else if ok && !req.ManuallyAdded {
		return nil, fmt.Errorf("NIC already exists")
	}

	s.m.Lock()
	defer s.m.Unlock()

	var link netlink.Link

	if req.ManuallyAdded {
		link, err = GetNIC(stack, req.Id)
		if err != nil {
			return nil, fmt.Errorf("failed to find nic: %s", err)
		}
		s.stacks[req.VpcId].Nics[req.Id] = &VNic{id: req.Id, vlan: uint16(req.SubnetVlanId), link: link, manual: true}
	} else {
		link, err = CreateNIC(stack, req.Id, uint16(req.SubnetVlanId))
		if err != nil {
			return nil, fmt.Errorf("failed to create nic: %s", err)
		}
	}

	for _, ip := range req.Ip {
		hwaddr := link.Attrs().HardwareAddr
		if req.ManuallyAdded {
			hwaddr, _ = net.ParseMAC(req.ManualHwaddr)
		}
		s.sdn.RegisterMacIP(uint32(req.VpcId), req.SubnetVlanId, hwaddr, net.ParseIP(ip))
	}

	// if err := s.transport.AddVLAN(uint32(stack.VPCID), uint16(req.SubnetVlanId)); err != nil {
	// 	log.Printf("failed to update vlan trunk: %s", err)
	// }

	s.logChange(&l2API.StackChange{
		VpcId:  req.VpcId,
		Action: "nic_added",
	})

	return formatToAPINic(stack, link, req.Id), nil
}

func (s *Server) validateAddNicRequest(req *l2API.NicRequest) error {
	if req.VpcId > 16777215 || req.VpcId == 0 {
		return fmt.Errorf("VPC ID out of range")
	}
	if req.ManuallyAdded && req.ManualHwaddr == "" {
		return fmt.Errorf("manually created nics must provide a hwaddr")
	}
	if !req.ManuallyAdded && req.ManualHwaddr != "" {
		return fmt.Errorf("automatically created nics cannot supply manual hwaddrs")
	}
	if req.SubnetVlanId == 0 {
		return fmt.Errorf("vlan must be set and non-zero")
	}

	return nil
}

//DeleteNIC Delete a NIC from a VPC linux bridge
func (s *Server) DeleteNIC(ctx context.Context, req *l2API.Nic) (*l2API.Empty, error) {
	if req.VpcId > 16777215 || req.VpcId == 0 {
		return nil, fmt.Errorf("VPC ID out of range")
	}

	if req.Vlan == 0 {
		return nil, fmt.Errorf("vlan must be set and non-zero")
	}

	stack, err := s.fetchStack(req.VpcId)
	if err != nil {
		return nil, err
	}

	ok, err := HasNIC(stack, req.Id)
	if err != nil {
		return nil, fmt.Errorf("Failed to create nic: %s", err)
	} else if !ok {
		return nil, fmt.Errorf("NIC does not exist: %s", req.Id)
	}

	link, _ := GetNIC(stack, req.Id)

	for _, ip := range req.Ip {
		if err := s.sdn.DeregisterMacIP(uint32(req.VpcId), req.Vlan, link.Attrs().HardwareAddr, net.ParseIP(ip)); err != nil {
			log.Printf("Failed to deregister NIC: %s", err)
		}
	}

	s.m.Lock()
	defer s.m.Unlock()

	nic, ok := stack.Nics[req.Id]

	// if err := s.transport.DelVLAN(uint32(stack.VPCID), nic.vlan); err != nil {
	// 	log.Printf("failed to update vlan trunk: %s", err)
	// }

	if ok && !nic.manual {
		err = DeleteNIC(stack, req.Id)
	} else if ok {
		delete(stack.Nics, req.Id)
	}

	s.logChange(&l2API.StackChange{
		VpcId:  req.VpcId,
		Action: "nic_deleted",
	})

	return &l2API.Empty{}, err
}

//NICStatus Get netlink status of NIC
func (s *Server) NICStatus(ctx context.Context, req *l2API.Nic) (*l2API.NicStatusResponse, error) {
	if req.VpcId > 16777215 {
		return nil, fmt.Errorf("VPC ID out of range")
	}

	stack, err := s.fetchStack(req.VpcId)
	if err != nil {
		return nil, err
	}

	nic, err := GetNIC(stack, req.Id)
	if err != nil {
		return nil, err
	}

	resp := &l2API.NicStatusResponse{}

	if nic == nil {
		resp.Status = l2API.LinkStatus_MISSING
	} else {
		if nic.Attrs().OperState != netlink.OperUp {
			resp.Status = l2API.LinkStatus_DOWN
		} else {
			resp.Status = l2API.LinkStatus_UP
		}
	}

	return resp, nil
}

//formatToAPIStack converts the local l2 stack to an API stack
func formatToAPIStack(stack *Stack) *l2API.Stack {
	APIStack := &l2API.Stack{
		VpcId: stack.VPCID,
	}

	if stack.Bridge != nil {
		APIStack.BridgeLinkIndex = int32(stack.Bridge.Attrs().Index)
		APIStack.BridgeLinkName = stack.Bridge.Name
	}

	return APIStack
}

//formatToAPINic converts the local NIC to an API NIC
func formatToAPINic(stack *Stack, link netlink.Link, id string) *l2API.Nic {
	vlan := getNicVlans(int32(link.Attrs().Index))

	return &l2API.Nic{
		VpcId:  stack.VPCID,
		Hwaddr: link.Attrs().HardwareAddr.String(),
		Name:   link.Attrs().Name,
		Index:  int32(link.Attrs().Index),
		Id:     id,
		Vlan:   uint32(vlan),
	}
}

//logChange broadcasts stack changes to all subscribers
func (s *Server) logChange(change *l2API.StackChange) {
	for _, ch := range s.watches {
		go func(ch chan l2API.StackChange) {
			ch <- *change
		}(ch)
	}
}

//fetchStack finds the stack for VPC from local or recreates it
func (s *Server) fetchStack(vpcID int32) (*Stack, error) {
	stack, ok := s.stacks[vpcID]

	if !ok {
		var err error
		stack, err = getStack(vpcID)
		if err != nil {
			return nil, err
		}
		s.stacks[vpcID] = stack
	}
	return stack, nil
}

//getStackStatus calculates the VPC stack status with interfaces and their 'up' status
func (s *Server) getStackStatus(stack *Stack) *l2API.StackStatusResponse {
	status := &l2API.StackStatusResponse{}

	ok, err := hasBridge(stack.VPCID)
	if err != nil || !ok {
		status.Bridge = l2API.LinkStatus_MISSING
	} else if stack.Bridge.OperState != netlink.OperUp {
		status.Bridge = l2API.LinkStatus_DOWN
	} else {
		status.Bridge = l2API.LinkStatus_UP
	}

	switch s.transport.Status(uint32(stack.VPCID)) {
	case transport.EPStatusUP:
		status.Transport = l2API.LinkStatus_UP
		break
	case transport.EPStatusDOWN:
		status.Transport = l2API.LinkStatus_DOWN
		break
	case transport.EPStatusMISSING:
		status.Transport = l2API.LinkStatus_MISSING
		break
	}

	return status
}
