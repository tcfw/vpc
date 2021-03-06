// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package l2

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion6

// L2ServiceClient is the client API for L2Service service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type L2ServiceClient interface {
	AddStack(ctx context.Context, in *StackRequest, opts ...grpc.CallOption) (*StackResponse, error)
	GetStack(ctx context.Context, in *StackRequest, opts ...grpc.CallOption) (*StackResponse, error)
	StackStatus(ctx context.Context, in *StackRequest, opts ...grpc.CallOption) (*StackStatusResponse, error)
	DeleteStack(ctx context.Context, in *StackRequest, opts ...grpc.CallOption) (*Empty, error)
	WatchStacks(ctx context.Context, in *Empty, opts ...grpc.CallOption) (L2Service_WatchStacksClient, error)
	AddNIC(ctx context.Context, in *NicRequest, opts ...grpc.CallOption) (*Nic, error)
	DeleteNIC(ctx context.Context, in *Nic, opts ...grpc.CallOption) (*Empty, error)
	NICStatus(ctx context.Context, in *Nic, opts ...grpc.CallOption) (*NicStatusResponse, error)
}

type l2ServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewL2ServiceClient(cc grpc.ClientConnInterface) L2ServiceClient {
	return &l2ServiceClient{cc}
}

func (c *l2ServiceClient) AddStack(ctx context.Context, in *StackRequest, opts ...grpc.CallOption) (*StackResponse, error) {
	out := new(StackResponse)
	err := c.cc.Invoke(ctx, "/vpc.l2.L2Service/AddStack", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *l2ServiceClient) GetStack(ctx context.Context, in *StackRequest, opts ...grpc.CallOption) (*StackResponse, error) {
	out := new(StackResponse)
	err := c.cc.Invoke(ctx, "/vpc.l2.L2Service/GetStack", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *l2ServiceClient) StackStatus(ctx context.Context, in *StackRequest, opts ...grpc.CallOption) (*StackStatusResponse, error) {
	out := new(StackStatusResponse)
	err := c.cc.Invoke(ctx, "/vpc.l2.L2Service/StackStatus", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *l2ServiceClient) DeleteStack(ctx context.Context, in *StackRequest, opts ...grpc.CallOption) (*Empty, error) {
	out := new(Empty)
	err := c.cc.Invoke(ctx, "/vpc.l2.L2Service/DeleteStack", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *l2ServiceClient) WatchStacks(ctx context.Context, in *Empty, opts ...grpc.CallOption) (L2Service_WatchStacksClient, error) {
	stream, err := c.cc.NewStream(ctx, &_L2Service_serviceDesc.Streams[0], "/vpc.l2.L2Service/WatchStacks", opts...)
	if err != nil {
		return nil, err
	}
	x := &l2ServiceWatchStacksClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type L2Service_WatchStacksClient interface {
	Recv() (*StackChange, error)
	grpc.ClientStream
}

type l2ServiceWatchStacksClient struct {
	grpc.ClientStream
}

func (x *l2ServiceWatchStacksClient) Recv() (*StackChange, error) {
	m := new(StackChange)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *l2ServiceClient) AddNIC(ctx context.Context, in *NicRequest, opts ...grpc.CallOption) (*Nic, error) {
	out := new(Nic)
	err := c.cc.Invoke(ctx, "/vpc.l2.L2Service/AddNIC", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *l2ServiceClient) DeleteNIC(ctx context.Context, in *Nic, opts ...grpc.CallOption) (*Empty, error) {
	out := new(Empty)
	err := c.cc.Invoke(ctx, "/vpc.l2.L2Service/DeleteNIC", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *l2ServiceClient) NICStatus(ctx context.Context, in *Nic, opts ...grpc.CallOption) (*NicStatusResponse, error) {
	out := new(NicStatusResponse)
	err := c.cc.Invoke(ctx, "/vpc.l2.L2Service/NICStatus", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// L2ServiceServer is the server API for L2Service service.
// All implementations must embed UnimplementedL2ServiceServer
// for forward compatibility
type L2ServiceServer interface {
	AddStack(context.Context, *StackRequest) (*StackResponse, error)
	GetStack(context.Context, *StackRequest) (*StackResponse, error)
	StackStatus(context.Context, *StackRequest) (*StackStatusResponse, error)
	DeleteStack(context.Context, *StackRequest) (*Empty, error)
	WatchStacks(*Empty, L2Service_WatchStacksServer) error
	AddNIC(context.Context, *NicRequest) (*Nic, error)
	DeleteNIC(context.Context, *Nic) (*Empty, error)
	NICStatus(context.Context, *Nic) (*NicStatusResponse, error)
	mustEmbedUnimplementedL2ServiceServer()
}

// UnimplementedL2ServiceServer must be embedded to have forward compatible implementations.
type UnimplementedL2ServiceServer struct {
}

func (*UnimplementedL2ServiceServer) AddStack(context.Context, *StackRequest) (*StackResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AddStack not implemented")
}
func (*UnimplementedL2ServiceServer) GetStack(context.Context, *StackRequest) (*StackResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetStack not implemented")
}
func (*UnimplementedL2ServiceServer) StackStatus(context.Context, *StackRequest) (*StackStatusResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method StackStatus not implemented")
}
func (*UnimplementedL2ServiceServer) DeleteStack(context.Context, *StackRequest) (*Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteStack not implemented")
}
func (*UnimplementedL2ServiceServer) WatchStacks(*Empty, L2Service_WatchStacksServer) error {
	return status.Errorf(codes.Unimplemented, "method WatchStacks not implemented")
}
func (*UnimplementedL2ServiceServer) AddNIC(context.Context, *NicRequest) (*Nic, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AddNIC not implemented")
}
func (*UnimplementedL2ServiceServer) DeleteNIC(context.Context, *Nic) (*Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteNIC not implemented")
}
func (*UnimplementedL2ServiceServer) NICStatus(context.Context, *Nic) (*NicStatusResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method NICStatus not implemented")
}
func (*UnimplementedL2ServiceServer) mustEmbedUnimplementedL2ServiceServer() {}

func RegisterL2ServiceServer(s *grpc.Server, srv L2ServiceServer) {
	s.RegisterService(&_L2Service_serviceDesc, srv)
}

func _L2Service_AddStack_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(StackRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(L2ServiceServer).AddStack(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/vpc.l2.L2Service/AddStack",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(L2ServiceServer).AddStack(ctx, req.(*StackRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _L2Service_GetStack_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(StackRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(L2ServiceServer).GetStack(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/vpc.l2.L2Service/GetStack",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(L2ServiceServer).GetStack(ctx, req.(*StackRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _L2Service_StackStatus_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(StackRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(L2ServiceServer).StackStatus(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/vpc.l2.L2Service/StackStatus",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(L2ServiceServer).StackStatus(ctx, req.(*StackRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _L2Service_DeleteStack_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(StackRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(L2ServiceServer).DeleteStack(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/vpc.l2.L2Service/DeleteStack",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(L2ServiceServer).DeleteStack(ctx, req.(*StackRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _L2Service_WatchStacks_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(Empty)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(L2ServiceServer).WatchStacks(m, &l2ServiceWatchStacksServer{stream})
}

type L2Service_WatchStacksServer interface {
	Send(*StackChange) error
	grpc.ServerStream
}

type l2ServiceWatchStacksServer struct {
	grpc.ServerStream
}

func (x *l2ServiceWatchStacksServer) Send(m *StackChange) error {
	return x.ServerStream.SendMsg(m)
}

func _L2Service_AddNIC_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(NicRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(L2ServiceServer).AddNIC(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/vpc.l2.L2Service/AddNIC",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(L2ServiceServer).AddNIC(ctx, req.(*NicRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _L2Service_DeleteNIC_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Nic)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(L2ServiceServer).DeleteNIC(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/vpc.l2.L2Service/DeleteNIC",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(L2ServiceServer).DeleteNIC(ctx, req.(*Nic))
	}
	return interceptor(ctx, in, info, handler)
}

func _L2Service_NICStatus_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Nic)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(L2ServiceServer).NICStatus(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/vpc.l2.L2Service/NICStatus",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(L2ServiceServer).NICStatus(ctx, req.(*Nic))
	}
	return interceptor(ctx, in, info, handler)
}

var _L2Service_serviceDesc = grpc.ServiceDesc{
	ServiceName: "vpc.l2.L2Service",
	HandlerType: (*L2ServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "AddStack",
			Handler:    _L2Service_AddStack_Handler,
		},
		{
			MethodName: "GetStack",
			Handler:    _L2Service_GetStack_Handler,
		},
		{
			MethodName: "StackStatus",
			Handler:    _L2Service_StackStatus_Handler,
		},
		{
			MethodName: "DeleteStack",
			Handler:    _L2Service_DeleteStack_Handler,
		},
		{
			MethodName: "AddNIC",
			Handler:    _L2Service_AddNIC_Handler,
		},
		{
			MethodName: "DeleteNIC",
			Handler:    _L2Service_DeleteNIC_Handler,
		},
		{
			MethodName: "NICStatus",
			Handler:    _L2Service_NICStatus_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "WatchStacks",
			Handler:       _L2Service_WatchStacks_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "l2.proto",
}
