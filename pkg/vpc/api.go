package vpc

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"sync"

	vpcAPI "github.com/tcfw/vpc/pkg/api/v1/vpc"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
)

//Serve starts the GRPC api server
func Serve(port uint) {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	srv, err := NewServer()
	if err != nil {
		log.Fatalln(err)
	}

	vpcAPI.RegisterVPCServiceServer(grpcServer, srv)
	log.Println("Starting gRPC server")
	grpcServer.Serve(lis)
}

//NewServer creates a new server instance
func NewServer() (*Server, error) {
	db, err := DBConn()
	if err != nil {
		return nil, err
	}

	return &Server{db: db}, nil
}

//Server vpc API server
type Server struct {
	m  sync.Mutex
	db *sql.DB
}

//VPCs provides a list of VPCs for an account
func (s *Server) VPCs(ctx context.Context, req *vpcAPI.VPCsRequest) (*vpcAPI.VPCsResponse, error) {
	//TODO(tcfw) valudate access to VPC info if not machine
	q, err := s.db.QueryContext(ctx, `SELECT * FROM vpcs LIMIT 500`)
	if err != nil {
		return nil, err
	}
	defer func() {
		q.Close()
	}()

	vpcs := []*vpcAPI.VPC{}

	for q.Next() {
		vpc := &vpcAPI.VPC{}
		if err := q.Scan(vpc); err != nil {
			return nil, err
		}
		vpcs = append(vpcs, vpc)
	}

	return &vpcAPI.VPCsResponse{VPCs: vpcs}, nil
}

//VPCInfo supplies VPC info to backend services
func (s *Server) VPCInfo(ctx context.Context, req *vpcAPI.VPCInfoRequest) (*vpcAPI.VPCInfoResponse, error) {
	g, ctx := errgroup.WithContext(ctx)

	vpcInfo := &vpcAPI.VPCInfoResponse{}
	var m sync.Mutex

	g.Go(func() error {
		vpc, err := s.getVPC(ctx, req.VpcId)
		if err != nil {
			return err
		}
		m.Lock()
		defer func() { m.Unlock() }()

		vpcInfo.Vpc = vpc
		return nil
	})

	g.Go(func() error {
		subnets, err := s.getVPCSubnets(ctx, req.VpcId)
		if err != nil {
			return err
		}
		m.Lock()
		defer func() { m.Unlock() }()

		vpcInfo.Subnets = subnets
		return nil
	})

	g.Go(func() error {
		igw, err := s.getVPCIGW(ctx, req.VpcId)
		if err != nil {
			return err
		}
		m.Lock()
		defer func() { m.Unlock() }()

		vpcInfo.InternetGW = igw
		return nil
	})

	if err := g.Wait(); err != nil {
		return nil, err
	}

	return vpcInfo, nil
}

//Subnets TODO
func (s *Server) Subnets(ctx context.Context, req *vpcAPI.VPCInfoRequest) (*vpcAPI.SubnetsResponse, error) {
	//TODO(tcfw) valudate access to VPC info if not machine

	subnets, err := s.getVPCSubnets(ctx, req.VpcId)
	if err != nil {
		return nil, err
	}

	return &vpcAPI.SubnetsResponse{Subnets: subnets}, nil
}

//InternetGWs TODO
func (s *Server) InternetGWs(ctx context.Context, req *vpcAPI.VPCInfoRequest) (*vpcAPI.InternetGWsRespones, error) {
	//TODO(tcfw) valudate access to VPC info if not machine

	// igw, err := s.getVPCIGW(ctx, req.VpcId)
	// if err != nil {
	// 	return nil, err
	// }

	// return &vpcAPI.InternetGWsRespones{InternetGws: igw}, nil
	return &vpcAPI.InternetGWsRespones{InternetGws: []*vpcAPI.InternetGW{}}, nil
}

//getVPC fetches a VPC from the DB
func (s *Server) getVPC(ctx context.Context, vpcID int32) (*vpcAPI.VPC, error) {
	q := s.db.QueryRowContext(ctx, `SELECT * FROM vpcs WHERE id = ?`, vpcID)

	vpc := &vpcAPI.VPC{}
	if err := q.Scan(vpc); err != nil {
		return nil, err
	}

	return vpc, nil
}

//getVPCIGW fetches an IGW for a VPC from the DB
func (s *Server) getVPCIGW(ctx context.Context, vpcID int32) (*vpcAPI.InternetGW, error) {
	q := s.db.QueryRowContext(ctx, `SELECT * FROM internet_gateways WHERE vpc_id = ?`, vpcID)

	igw := &vpcAPI.InternetGW{}
	if err := q.Scan(igw); err != nil {
		return nil, err
	}

	return igw, nil
}

//getVPCSubnets fetches all subnets related to a VPC from the DB
func (s *Server) getVPCSubnets(ctx context.Context, vpcID int32) ([]*vpcAPI.Subnet, error) {
	q, err := s.db.QueryContext(ctx, `SELECT * FROM subnets WHERE vpc_id = ? LIMIT 4096`, vpcID)
	if err != nil {
		return nil, err
	}
	defer func() {
		q.Close()
	}()

	subnets := []*vpcAPI.Subnet{}

	for q.Next() {
		subnet := &vpcAPI.Subnet{}
		if err := q.Scan(subnet); err != nil {
			return nil, err
		}
		subnets = append(subnets, subnet)
	}

	return subnets, nil
}
