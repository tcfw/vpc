// Code generated by protoc-gen-go. DO NOT EDIT.
// source: l2.proto

package vpc_l2

import (
	context "context"
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

type LinkStatus int32

const (
	LinkStatus_DOWN    LinkStatus = 0
	LinkStatus_UP      LinkStatus = 1
	LinkStatus_MISSING LinkStatus = 2
)

var LinkStatus_name = map[int32]string{
	0: "DOWN",
	1: "UP",
	2: "MISSING",
}

var LinkStatus_value = map[string]int32{
	"DOWN":    0,
	"UP":      1,
	"MISSING": 2,
}

func (x LinkStatus) String() string {
	return proto.EnumName(LinkStatus_name, int32(x))
}

func (LinkStatus) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_f692411d37028fef, []int{0}
}

type NicRequest_NicType int32

const (
	NicRequest_TAP     NicRequest_NicType = 0
	NicRequest_MACVTAP NicRequest_NicType = 1
)

var NicRequest_NicType_name = map[int32]string{
	0: "TAP",
	1: "MACVTAP",
}

var NicRequest_NicType_value = map[string]int32{
	"TAP":     0,
	"MACVTAP": 1,
}

func (x NicRequest_NicType) String() string {
	return proto.EnumName(NicRequest_NicType_name, int32(x))
}

func (NicRequest_NicType) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_f692411d37028fef, []int{5, 0}
}

type StackRequest struct {
	VpcId                int32    `protobuf:"varint,1,opt,name=vpc_id,json=vpcId,proto3" json:"vpc_id,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *StackRequest) Reset()         { *m = StackRequest{} }
func (m *StackRequest) String() string { return proto.CompactTextString(m) }
func (*StackRequest) ProtoMessage()    {}
func (*StackRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_f692411d37028fef, []int{0}
}

func (m *StackRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_StackRequest.Unmarshal(m, b)
}
func (m *StackRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_StackRequest.Marshal(b, m, deterministic)
}
func (m *StackRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_StackRequest.Merge(m, src)
}
func (m *StackRequest) XXX_Size() int {
	return xxx_messageInfo_StackRequest.Size(m)
}
func (m *StackRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_StackRequest.DiscardUnknown(m)
}

var xxx_messageInfo_StackRequest proto.InternalMessageInfo

func (m *StackRequest) GetVpcId() int32 {
	if m != nil {
		return m.VpcId
	}
	return 0
}

type Stack struct {
	VpcID                int32    `protobuf:"varint,1,opt,name=vpcID,proto3" json:"vpcID,omitempty"`
	BridgeLinkName       string   `protobuf:"bytes,2,opt,name=bridge_link_name,json=bridgeLinkName,proto3" json:"bridge_link_name,omitempty"`
	BridgeLinkIndex      int32    `protobuf:"varint,3,opt,name=bridge_link_index,json=bridgeLinkIndex,proto3" json:"bridge_link_index,omitempty"`
	VtepLinkName         string   `protobuf:"bytes,4,opt,name=vtep_link_name,json=vtepLinkName,proto3" json:"vtep_link_name,omitempty"`
	VtepLinkIndex        int32    `protobuf:"varint,5,opt,name=vtep_link_index,json=vtepLinkIndex,proto3" json:"vtep_link_index,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Stack) Reset()         { *m = Stack{} }
func (m *Stack) String() string { return proto.CompactTextString(m) }
func (*Stack) ProtoMessage()    {}
func (*Stack) Descriptor() ([]byte, []int) {
	return fileDescriptor_f692411d37028fef, []int{1}
}

func (m *Stack) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Stack.Unmarshal(m, b)
}
func (m *Stack) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Stack.Marshal(b, m, deterministic)
}
func (m *Stack) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Stack.Merge(m, src)
}
func (m *Stack) XXX_Size() int {
	return xxx_messageInfo_Stack.Size(m)
}
func (m *Stack) XXX_DiscardUnknown() {
	xxx_messageInfo_Stack.DiscardUnknown(m)
}

var xxx_messageInfo_Stack proto.InternalMessageInfo

func (m *Stack) GetVpcID() int32 {
	if m != nil {
		return m.VpcID
	}
	return 0
}

func (m *Stack) GetBridgeLinkName() string {
	if m != nil {
		return m.BridgeLinkName
	}
	return ""
}

func (m *Stack) GetBridgeLinkIndex() int32 {
	if m != nil {
		return m.BridgeLinkIndex
	}
	return 0
}

func (m *Stack) GetVtepLinkName() string {
	if m != nil {
		return m.VtepLinkName
	}
	return ""
}

func (m *Stack) GetVtepLinkIndex() int32 {
	if m != nil {
		return m.VtepLinkIndex
	}
	return 0
}

type StackResponse struct {
	Stack                *Stack               `protobuf:"bytes,1,opt,name=stack,proto3" json:"stack,omitempty"`
	Status               *StackStatusResponse `protobuf:"bytes,2,opt,name=status,proto3" json:"status,omitempty"`
	XXX_NoUnkeyedLiteral struct{}             `json:"-"`
	XXX_unrecognized     []byte               `json:"-"`
	XXX_sizecache        int32                `json:"-"`
}

func (m *StackResponse) Reset()         { *m = StackResponse{} }
func (m *StackResponse) String() string { return proto.CompactTextString(m) }
func (*StackResponse) ProtoMessage()    {}
func (*StackResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_f692411d37028fef, []int{2}
}

func (m *StackResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_StackResponse.Unmarshal(m, b)
}
func (m *StackResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_StackResponse.Marshal(b, m, deterministic)
}
func (m *StackResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_StackResponse.Merge(m, src)
}
func (m *StackResponse) XXX_Size() int {
	return xxx_messageInfo_StackResponse.Size(m)
}
func (m *StackResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_StackResponse.DiscardUnknown(m)
}

var xxx_messageInfo_StackResponse proto.InternalMessageInfo

func (m *StackResponse) GetStack() *Stack {
	if m != nil {
		return m.Stack
	}
	return nil
}

func (m *StackResponse) GetStatus() *StackStatusResponse {
	if m != nil {
		return m.Status
	}
	return nil
}

type StackStatusResponse struct {
	Bridge               LinkStatus `protobuf:"varint,1,opt,name=bridge,proto3,enum=vpc.l2.LinkStatus" json:"bridge,omitempty"`
	Vtep                 LinkStatus `protobuf:"varint,2,opt,name=vtep,proto3,enum=vpc.l2.LinkStatus" json:"vtep,omitempty"`
	XXX_NoUnkeyedLiteral struct{}   `json:"-"`
	XXX_unrecognized     []byte     `json:"-"`
	XXX_sizecache        int32      `json:"-"`
}

func (m *StackStatusResponse) Reset()         { *m = StackStatusResponse{} }
func (m *StackStatusResponse) String() string { return proto.CompactTextString(m) }
func (*StackStatusResponse) ProtoMessage()    {}
func (*StackStatusResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_f692411d37028fef, []int{3}
}

func (m *StackStatusResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_StackStatusResponse.Unmarshal(m, b)
}
func (m *StackStatusResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_StackStatusResponse.Marshal(b, m, deterministic)
}
func (m *StackStatusResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_StackStatusResponse.Merge(m, src)
}
func (m *StackStatusResponse) XXX_Size() int {
	return xxx_messageInfo_StackStatusResponse.Size(m)
}
func (m *StackStatusResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_StackStatusResponse.DiscardUnknown(m)
}

var xxx_messageInfo_StackStatusResponse proto.InternalMessageInfo

func (m *StackStatusResponse) GetBridge() LinkStatus {
	if m != nil {
		return m.Bridge
	}
	return LinkStatus_DOWN
}

func (m *StackStatusResponse) GetVtep() LinkStatus {
	if m != nil {
		return m.Vtep
	}
	return LinkStatus_DOWN
}

type StackChange struct {
	VpcId                int32                `protobuf:"varint,1,opt,name=vpc_id,json=vpcId,proto3" json:"vpc_id,omitempty"`
	Action               string               `protobuf:"bytes,2,opt,name=action,proto3" json:"action,omitempty"`
	Status               *StackStatusResponse `protobuf:"bytes,3,opt,name=status,proto3" json:"status,omitempty"`
	XXX_NoUnkeyedLiteral struct{}             `json:"-"`
	XXX_unrecognized     []byte               `json:"-"`
	XXX_sizecache        int32                `json:"-"`
}

func (m *StackChange) Reset()         { *m = StackChange{} }
func (m *StackChange) String() string { return proto.CompactTextString(m) }
func (*StackChange) ProtoMessage()    {}
func (*StackChange) Descriptor() ([]byte, []int) {
	return fileDescriptor_f692411d37028fef, []int{4}
}

func (m *StackChange) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_StackChange.Unmarshal(m, b)
}
func (m *StackChange) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_StackChange.Marshal(b, m, deterministic)
}
func (m *StackChange) XXX_Merge(src proto.Message) {
	xxx_messageInfo_StackChange.Merge(m, src)
}
func (m *StackChange) XXX_Size() int {
	return xxx_messageInfo_StackChange.Size(m)
}
func (m *StackChange) XXX_DiscardUnknown() {
	xxx_messageInfo_StackChange.DiscardUnknown(m)
}

var xxx_messageInfo_StackChange proto.InternalMessageInfo

func (m *StackChange) GetVpcId() int32 {
	if m != nil {
		return m.VpcId
	}
	return 0
}

func (m *StackChange) GetAction() string {
	if m != nil {
		return m.Action
	}
	return ""
}

func (m *StackChange) GetStatus() *StackStatusResponse {
	if m != nil {
		return m.Status
	}
	return nil
}

type NicRequest struct {
	VpcId                int32              `protobuf:"varint,1,opt,name=vpc_id,json=vpcId,proto3" json:"vpc_id,omitempty"`
	Id                   string             `protobuf:"bytes,2,opt,name=id,proto3" json:"id,omitempty"`
	Type                 NicRequest_NicType `protobuf:"varint,3,opt,name=type,proto3,enum=vpc.l2.NicRequest_NicType" json:"type,omitempty"`
	XXX_NoUnkeyedLiteral struct{}           `json:"-"`
	XXX_unrecognized     []byte             `json:"-"`
	XXX_sizecache        int32              `json:"-"`
}

func (m *NicRequest) Reset()         { *m = NicRequest{} }
func (m *NicRequest) String() string { return proto.CompactTextString(m) }
func (*NicRequest) ProtoMessage()    {}
func (*NicRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_f692411d37028fef, []int{5}
}

func (m *NicRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_NicRequest.Unmarshal(m, b)
}
func (m *NicRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_NicRequest.Marshal(b, m, deterministic)
}
func (m *NicRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_NicRequest.Merge(m, src)
}
func (m *NicRequest) XXX_Size() int {
	return xxx_messageInfo_NicRequest.Size(m)
}
func (m *NicRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_NicRequest.DiscardUnknown(m)
}

var xxx_messageInfo_NicRequest proto.InternalMessageInfo

func (m *NicRequest) GetVpcId() int32 {
	if m != nil {
		return m.VpcId
	}
	return 0
}

func (m *NicRequest) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

func (m *NicRequest) GetType() NicRequest_NicType {
	if m != nil {
		return m.Type
	}
	return NicRequest_TAP
}

type Nic struct {
	VpcId                int32    `protobuf:"varint,1,opt,name=vpc_id,json=vpcId,proto3" json:"vpc_id,omitempty"`
	Hwaddr               string   `protobuf:"bytes,2,opt,name=hwaddr,proto3" json:"hwaddr,omitempty"`
	Id                   string   `protobuf:"bytes,3,opt,name=id,proto3" json:"id,omitempty"`
	Name                 string   `protobuf:"bytes,4,opt,name=name,proto3" json:"name,omitempty"`
	Index                int32    `protobuf:"varint,5,opt,name=index,proto3" json:"index,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Nic) Reset()         { *m = Nic{} }
func (m *Nic) String() string { return proto.CompactTextString(m) }
func (*Nic) ProtoMessage()    {}
func (*Nic) Descriptor() ([]byte, []int) {
	return fileDescriptor_f692411d37028fef, []int{6}
}

func (m *Nic) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Nic.Unmarshal(m, b)
}
func (m *Nic) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Nic.Marshal(b, m, deterministic)
}
func (m *Nic) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Nic.Merge(m, src)
}
func (m *Nic) XXX_Size() int {
	return xxx_messageInfo_Nic.Size(m)
}
func (m *Nic) XXX_DiscardUnknown() {
	xxx_messageInfo_Nic.DiscardUnknown(m)
}

var xxx_messageInfo_Nic proto.InternalMessageInfo

func (m *Nic) GetVpcId() int32 {
	if m != nil {
		return m.VpcId
	}
	return 0
}

func (m *Nic) GetHwaddr() string {
	if m != nil {
		return m.Hwaddr
	}
	return ""
}

func (m *Nic) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

func (m *Nic) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *Nic) GetIndex() int32 {
	if m != nil {
		return m.Index
	}
	return 0
}

type NicStatusResponse struct {
	Status               LinkStatus `protobuf:"varint,1,opt,name=status,proto3,enum=vpc.l2.LinkStatus" json:"status,omitempty"`
	XXX_NoUnkeyedLiteral struct{}   `json:"-"`
	XXX_unrecognized     []byte     `json:"-"`
	XXX_sizecache        int32      `json:"-"`
}

func (m *NicStatusResponse) Reset()         { *m = NicStatusResponse{} }
func (m *NicStatusResponse) String() string { return proto.CompactTextString(m) }
func (*NicStatusResponse) ProtoMessage()    {}
func (*NicStatusResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_f692411d37028fef, []int{7}
}

func (m *NicStatusResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_NicStatusResponse.Unmarshal(m, b)
}
func (m *NicStatusResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_NicStatusResponse.Marshal(b, m, deterministic)
}
func (m *NicStatusResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_NicStatusResponse.Merge(m, src)
}
func (m *NicStatusResponse) XXX_Size() int {
	return xxx_messageInfo_NicStatusResponse.Size(m)
}
func (m *NicStatusResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_NicStatusResponse.DiscardUnknown(m)
}

var xxx_messageInfo_NicStatusResponse proto.InternalMessageInfo

func (m *NicStatusResponse) GetStatus() LinkStatus {
	if m != nil {
		return m.Status
	}
	return LinkStatus_DOWN
}

type Empty struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Empty) Reset()         { *m = Empty{} }
func (m *Empty) String() string { return proto.CompactTextString(m) }
func (*Empty) ProtoMessage()    {}
func (*Empty) Descriptor() ([]byte, []int) {
	return fileDescriptor_f692411d37028fef, []int{8}
}

func (m *Empty) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Empty.Unmarshal(m, b)
}
func (m *Empty) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Empty.Marshal(b, m, deterministic)
}
func (m *Empty) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Empty.Merge(m, src)
}
func (m *Empty) XXX_Size() int {
	return xxx_messageInfo_Empty.Size(m)
}
func (m *Empty) XXX_DiscardUnknown() {
	xxx_messageInfo_Empty.DiscardUnknown(m)
}

var xxx_messageInfo_Empty proto.InternalMessageInfo

func init() {
	proto.RegisterEnum("vpc.l2.LinkStatus", LinkStatus_name, LinkStatus_value)
	proto.RegisterEnum("vpc.l2.NicRequest_NicType", NicRequest_NicType_name, NicRequest_NicType_value)
	proto.RegisterType((*StackRequest)(nil), "vpc.l2.StackRequest")
	proto.RegisterType((*Stack)(nil), "vpc.l2.Stack")
	proto.RegisterType((*StackResponse)(nil), "vpc.l2.StackResponse")
	proto.RegisterType((*StackStatusResponse)(nil), "vpc.l2.StackStatusResponse")
	proto.RegisterType((*StackChange)(nil), "vpc.l2.StackChange")
	proto.RegisterType((*NicRequest)(nil), "vpc.l2.NicRequest")
	proto.RegisterType((*Nic)(nil), "vpc.l2.Nic")
	proto.RegisterType((*NicStatusResponse)(nil), "vpc.l2.NicStatusResponse")
	proto.RegisterType((*Empty)(nil), "vpc.l2.Empty")
}

func init() { proto.RegisterFile("l2.proto", fileDescriptor_f692411d37028fef) }

var fileDescriptor_f692411d37028fef = []byte{
	// 592 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x9c, 0x54, 0x41, 0x6f, 0xda, 0x4c,
	0x10, 0xb5, 0x31, 0x36, 0x61, 0x9c, 0x10, 0x32, 0xf9, 0xf2, 0x29, 0xa5, 0x87, 0x46, 0xdb, 0x36,
	0x4a, 0x89, 0x84, 0x2a, 0xa7, 0x3d, 0xf4, 0x54, 0x51, 0xa8, 0x22, 0x4b, 0xa9, 0x1b, 0x41, 0xda,
	0x1c, 0x91, 0xe3, 0x5d, 0x85, 0x15, 0x60, 0x1c, 0xbc, 0xa1, 0xe5, 0xde, 0x3f, 0x96, 0x7f, 0x56,
	0x79, 0x77, 0x01, 0x93, 0x50, 0x54, 0xf5, 0xe6, 0x99, 0x79, 0xfb, 0xe6, 0xbd, 0xd9, 0x59, 0xc3,
	0xd6, 0xd0, 0x6b, 0x24, 0x93, 0xb1, 0x18, 0xa3, 0x33, 0x4d, 0xa2, 0xc6, 0xd0, 0x23, 0xaf, 0x61,
	0xbb, 0x2b, 0xc2, 0x68, 0xd0, 0x61, 0x77, 0xf7, 0x2c, 0x15, 0x78, 0x00, 0x59, 0xa5, 0xc7, 0xe9,
	0xa1, 0x79, 0x64, 0x9e, 0xd8, 0x1d, 0x7b, 0x9a, 0x44, 0x3e, 0x25, 0x0f, 0x26, 0xd8, 0x12, 0x87,
	0xff, 0x81, 0x4c, 0xb5, 0xf3, 0xf5, 0x36, 0x9e, 0x40, 0xf5, 0x66, 0xc2, 0xe9, 0x2d, 0xeb, 0x0d,
	0x79, 0x3c, 0xe8, 0xc5, 0xe1, 0x88, 0x1d, 0x16, 0x8e, 0xcc, 0x93, 0x72, 0xa7, 0xa2, 0xf2, 0x17,
	0x3c, 0x1e, 0x04, 0xe1, 0x88, 0x61, 0x1d, 0xf6, 0xf2, 0x48, 0x1e, 0x53, 0xf6, 0xf3, 0xd0, 0x92,
	0x5c, 0xbb, 0x4b, 0xa8, 0x9f, 0xa5, 0xf1, 0x15, 0x54, 0xa6, 0x82, 0x25, 0x39, 0xce, 0xa2, 0xe4,
	0xdc, 0xce, 0xb2, 0x0b, 0xc6, 0x63, 0xd8, 0x5d, 0xa2, 0x14, 0x9f, 0x2d, 0xf9, 0x76, 0xe6, 0x30,
	0xc9, 0x46, 0x38, 0xec, 0x68, 0xab, 0x69, 0x32, 0x8e, 0x53, 0x86, 0x2f, 0xc1, 0x4e, 0xb3, 0x84,
	0xb4, 0xe2, 0x7a, 0x3b, 0x0d, 0x35, 0x93, 0x86, 0x42, 0xa9, 0x1a, 0x9e, 0x81, 0x93, 0x8a, 0x50,
	0xdc, 0xa7, 0xd2, 0x8f, 0xeb, 0x3d, 0x5f, 0x41, 0x75, 0x65, 0x69, 0xce, 0xd8, 0xd1, 0x50, 0xc2,
	0x61, 0x7f, 0x4d, 0x19, 0xeb, 0xe0, 0x28, 0x8b, 0xb2, 0x63, 0xc5, 0xc3, 0x39, 0x57, 0x26, 0x52,
	0x63, 0x35, 0x02, 0x8f, 0xa1, 0x98, 0xc9, 0x97, 0x5d, 0xd7, 0x23, 0x65, 0x9d, 0xdc, 0x81, 0x2b,
	0x5b, 0xb5, 0xfa, 0x61, 0x7c, 0xcb, 0xfe, 0x70, 0x7f, 0xf8, 0x3f, 0x38, 0x61, 0x24, 0xf8, 0x38,
	0xd6, 0xb7, 0xa2, 0xa3, 0x9c, 0x3b, 0xeb, 0xef, 0xdd, 0xfd, 0x32, 0x01, 0x02, 0x1e, 0x6d, 0x5e,
	0x19, 0xac, 0x40, 0x81, 0x53, 0xdd, 0xae, 0xc0, 0x29, 0x36, 0xa0, 0x28, 0x66, 0x09, 0x93, 0x8d,
	0x2a, 0x5e, 0x6d, 0xde, 0x68, 0x49, 0x94, 0x7d, 0x5e, 0xcd, 0x12, 0xd6, 0x91, 0x38, 0xf2, 0x02,
	0x4a, 0x3a, 0x81, 0x25, 0xb0, 0xae, 0x9a, 0x97, 0x55, 0x03, 0x5d, 0x28, 0x7d, 0x69, 0xb6, 0xbe,
	0x67, 0x81, 0x49, 0x62, 0xb0, 0x02, 0x1e, 0x6d, 0x70, 0xdc, 0xff, 0x11, 0x52, 0x3a, 0x99, 0x3b,
	0x56, 0x91, 0x96, 0x65, 0x2d, 0x64, 0x21, 0x14, 0x73, 0x9b, 0x25, 0xbf, 0xb3, 0x1d, 0xcf, 0xef,
	0x91, 0x0a, 0xc8, 0x47, 0xd8, 0x0b, 0x78, 0xf4, 0xf4, 0x4a, 0xf5, 0x00, 0x37, 0x5c, 0xa9, 0x9e,
	0x5b, 0x09, 0xec, 0xcf, 0xa3, 0x44, 0xcc, 0xea, 0xa7, 0x00, 0xcb, 0x32, 0x6e, 0x41, 0xb1, 0xfd,
	0xf5, 0x3a, 0xa8, 0x1a, 0xe8, 0x40, 0xe1, 0xdb, 0x65, 0xd5, 0x94, 0x36, 0xfd, 0x6e, 0xd7, 0x0f,
	0xce, 0xab, 0x05, 0xef, 0xc1, 0x82, 0xf2, 0x85, 0xd7, 0x65, 0x93, 0x29, 0x8f, 0x18, 0x7e, 0x80,
	0xad, 0x26, 0xa5, 0xfa, 0x29, 0xae, 0x2e, 0xac, 0x9a, 0x62, 0xed, 0xe0, 0x51, 0x56, 0x09, 0x25,
	0x46, 0x76, 0xf4, 0x9c, 0x89, 0x7f, 0x3a, 0xfa, 0x49, 0x2f, 0x99, 0x56, 0xbc, 0xfe, 0xf4, 0xa6,
	0xdd, 0x21, 0x06, 0xbe, 0x03, 0xb7, 0xcd, 0x86, 0x4c, 0xb0, 0x4d, 0x0a, 0x16, 0x6f, 0x50, 0x0e,
	0x8a, 0x18, 0xf8, 0x1e, 0xdc, 0xeb, 0x50, 0x44, 0x7d, 0x89, 0x4a, 0x71, 0xb5, 0x5e, 0xdb, 0x5f,
	0x21, 0x51, 0x4f, 0x80, 0x18, 0x6f, 0x4d, 0x3c, 0x05, 0xa7, 0x49, 0x69, 0xe0, 0xb7, 0x10, 0x9f,
	0x2e, 0x5a, 0xcd, 0xcd, 0xe5, 0x88, 0x81, 0x6f, 0xa0, 0xac, 0x94, 0x65, 0xf8, 0x7c, 0x6d, 0x9d,
	0x9c, 0x72, 0xe0, 0xb7, 0xf4, 0x18, 0x56, 0xa0, 0xcf, 0x72, 0xc1, 0x63, 0xef, 0x37, 0x8e, 0xfc,
	0xe9, 0x9e, 0xfd, 0x0e, 0x00, 0x00, 0xff, 0xff, 0xbf, 0x25, 0x63, 0xca, 0x80, 0x05, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// L2ServiceClient is the client API for L2Service service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
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
	cc *grpc.ClientConn
}

func NewL2ServiceClient(cc *grpc.ClientConn) L2ServiceClient {
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
type L2ServiceServer interface {
	AddStack(context.Context, *StackRequest) (*StackResponse, error)
	GetStack(context.Context, *StackRequest) (*StackResponse, error)
	StackStatus(context.Context, *StackRequest) (*StackStatusResponse, error)
	DeleteStack(context.Context, *StackRequest) (*Empty, error)
	WatchStacks(*Empty, L2Service_WatchStacksServer) error
	AddNIC(context.Context, *NicRequest) (*Nic, error)
	DeleteNIC(context.Context, *Nic) (*Empty, error)
	NICStatus(context.Context, *Nic) (*NicStatusResponse, error)
}

// UnimplementedL2ServiceServer can be embedded to have forward compatible implementations.
type UnimplementedL2ServiceServer struct {
}

func (*UnimplementedL2ServiceServer) AddStack(ctx context.Context, req *StackRequest) (*StackResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AddStack not implemented")
}
func (*UnimplementedL2ServiceServer) GetStack(ctx context.Context, req *StackRequest) (*StackResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetStack not implemented")
}
func (*UnimplementedL2ServiceServer) StackStatus(ctx context.Context, req *StackRequest) (*StackStatusResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method StackStatus not implemented")
}
func (*UnimplementedL2ServiceServer) DeleteStack(ctx context.Context, req *StackRequest) (*Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteStack not implemented")
}
func (*UnimplementedL2ServiceServer) WatchStacks(req *Empty, srv L2Service_WatchStacksServer) error {
	return status.Errorf(codes.Unimplemented, "method WatchStacks not implemented")
}
func (*UnimplementedL2ServiceServer) AddNIC(ctx context.Context, req *NicRequest) (*Nic, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AddNIC not implemented")
}
func (*UnimplementedL2ServiceServer) DeleteNIC(ctx context.Context, req *Nic) (*Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteNIC not implemented")
}
func (*UnimplementedL2ServiceServer) NICStatus(ctx context.Context, req *Nic) (*NicStatusResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method NICStatus not implemented")
}

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
