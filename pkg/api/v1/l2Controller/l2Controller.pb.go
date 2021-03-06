// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0-devel
// 	protoc        v3.6.1
// source: l2Controller.proto

package l2Controller

import (
	proto "github.com/golang/protobuf/proto"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// This is a compile-time assertion that a sufficiently up-to-date version
// of the legacy proto package is being used.
const _ = proto.ProtoPackageIsVersion4

type LookupType int32

const (
	LookupType_MAC LookupType = 0
	LookupType_IP  LookupType = 1
)

// Enum value maps for LookupType.
var (
	LookupType_name = map[int32]string{
		0: "MAC",
		1: "IP",
	}
	LookupType_value = map[string]int32{
		"MAC": 0,
		"IP":  1,
	}
)

func (x LookupType) Enum() *LookupType {
	p := new(LookupType)
	*p = x
	return p
}

func (x LookupType) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (LookupType) Descriptor() protoreflect.EnumDescriptor {
	return file_l2Controller_proto_enumTypes[0].Descriptor()
}

func (LookupType) Type() protoreflect.EnumType {
	return &file_l2Controller_proto_enumTypes[0]
}

func (x LookupType) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use LookupType.Descriptor instead.
func (LookupType) EnumDescriptor() ([]byte, []int) {
	return file_l2Controller_proto_rawDescGZIP(), []int{0}
}

type MACIPRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	VNID uint32 `protobuf:"varint,1,opt,name=VNID,proto3" json:"VNID,omitempty"`
	VLAN uint32 `protobuf:"varint,2,opt,name=VLAN,proto3" json:"VLAN,omitempty"`
	MAC  string `protobuf:"bytes,3,opt,name=MAC,proto3" json:"MAC,omitempty"`
	IP   string `protobuf:"bytes,4,opt,name=IP,proto3" json:"IP,omitempty"`
}

func (x *MACIPRequest) Reset() {
	*x = MACIPRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_l2Controller_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MACIPRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MACIPRequest) ProtoMessage() {}

func (x *MACIPRequest) ProtoReflect() protoreflect.Message {
	mi := &file_l2Controller_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MACIPRequest.ProtoReflect.Descriptor instead.
func (*MACIPRequest) Descriptor() ([]byte, []int) {
	return file_l2Controller_proto_rawDescGZIP(), []int{0}
}

func (x *MACIPRequest) GetVNID() uint32 {
	if x != nil {
		return x.VNID
	}
	return 0
}

func (x *MACIPRequest) GetVLAN() uint32 {
	if x != nil {
		return x.VLAN
	}
	return 0
}

func (x *MACIPRequest) GetMAC() string {
	if x != nil {
		return x.MAC
	}
	return ""
}

func (x *MACIPRequest) GetIP() string {
	if x != nil {
		return x.IP
	}
	return ""
}

type MACIPResp struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *MACIPResp) Reset() {
	*x = MACIPResp{}
	if protoimpl.UnsafeEnabled {
		mi := &file_l2Controller_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MACIPResp) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MACIPResp) ProtoMessage() {}

func (x *MACIPResp) ProtoReflect() protoreflect.Message {
	mi := &file_l2Controller_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MACIPResp.ProtoReflect.Descriptor instead.
func (*MACIPResp) Descriptor() ([]byte, []int) {
	return file_l2Controller_proto_rawDescGZIP(), []int{1}
}

type LookupRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	LookupType LookupType `protobuf:"varint,1,opt,name=LookupType,proto3,enum=vpc.l2Controller.LookupType" json:"LookupType,omitempty"`
	VNID       uint32     `protobuf:"varint,2,opt,name=VNID,proto3" json:"VNID,omitempty"`
	VLAN       uint32     `protobuf:"varint,3,opt,name=VLAN,proto3" json:"VLAN,omitempty"`
	MAC        string     `protobuf:"bytes,4,opt,name=MAC,proto3" json:"MAC,omitempty"`
	IP         string     `protobuf:"bytes,5,opt,name=IP,proto3" json:"IP,omitempty"`
}

func (x *LookupRequest) Reset() {
	*x = LookupRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_l2Controller_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *LookupRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*LookupRequest) ProtoMessage() {}

func (x *LookupRequest) ProtoReflect() protoreflect.Message {
	mi := &file_l2Controller_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use LookupRequest.ProtoReflect.Descriptor instead.
func (*LookupRequest) Descriptor() ([]byte, []int) {
	return file_l2Controller_proto_rawDescGZIP(), []int{2}
}

func (x *LookupRequest) GetLookupType() LookupType {
	if x != nil {
		return x.LookupType
	}
	return LookupType_MAC
}

func (x *LookupRequest) GetVNID() uint32 {
	if x != nil {
		return x.VNID
	}
	return 0
}

func (x *LookupRequest) GetVLAN() uint32 {
	if x != nil {
		return x.VLAN
	}
	return 0
}

func (x *LookupRequest) GetMAC() string {
	if x != nil {
		return x.MAC
	}
	return ""
}

func (x *LookupRequest) GetIP() string {
	if x != nil {
		return x.IP
	}
	return ""
}

type LookupResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	MAC string `protobuf:"bytes,1,opt,name=MAC,proto3" json:"MAC,omitempty"`
	IP  string `protobuf:"bytes,2,opt,name=IP,proto3" json:"IP,omitempty"`
}

func (x *LookupResponse) Reset() {
	*x = LookupResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_l2Controller_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *LookupResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*LookupResponse) ProtoMessage() {}

func (x *LookupResponse) ProtoReflect() protoreflect.Message {
	mi := &file_l2Controller_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use LookupResponse.ProtoReflect.Descriptor instead.
func (*LookupResponse) Descriptor() ([]byte, []int) {
	return file_l2Controller_proto_rawDescGZIP(), []int{3}
}

func (x *LookupResponse) GetMAC() string {
	if x != nil {
		return x.MAC
	}
	return ""
}

func (x *LookupResponse) GetIP() string {
	if x != nil {
		return x.IP
	}
	return ""
}

type VNIDRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	VNID uint32 `protobuf:"varint,1,opt,name=VNID,proto3" json:"VNID,omitempty"`
}

func (x *VNIDRequest) Reset() {
	*x = VNIDRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_l2Controller_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *VNIDRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*VNIDRequest) ProtoMessage() {}

func (x *VNIDRequest) ProtoReflect() protoreflect.Message {
	mi := &file_l2Controller_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use VNIDRequest.ProtoReflect.Descriptor instead.
func (*VNIDRequest) Descriptor() ([]byte, []int) {
	return file_l2Controller_proto_rawDescGZIP(), []int{4}
}

func (x *VNIDRequest) GetVNID() uint32 {
	if x != nil {
		return x.VNID
	}
	return 0
}

type RegResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *RegResponse) Reset() {
	*x = RegResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_l2Controller_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RegResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RegResponse) ProtoMessage() {}

func (x *RegResponse) ProtoReflect() protoreflect.Message {
	mi := &file_l2Controller_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RegResponse.ProtoReflect.Descriptor instead.
func (*RegResponse) Descriptor() ([]byte, []int) {
	return file_l2Controller_proto_rawDescGZIP(), []int{5}
}

type BroadcastEndpointResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	IP []string `protobuf:"bytes,1,rep,name=IP,proto3" json:"IP,omitempty"`
}

func (x *BroadcastEndpointResponse) Reset() {
	*x = BroadcastEndpointResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_l2Controller_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *BroadcastEndpointResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BroadcastEndpointResponse) ProtoMessage() {}

func (x *BroadcastEndpointResponse) ProtoReflect() protoreflect.Message {
	mi := &file_l2Controller_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BroadcastEndpointResponse.ProtoReflect.Descriptor instead.
func (*BroadcastEndpointResponse) Descriptor() ([]byte, []int) {
	return file_l2Controller_proto_rawDescGZIP(), []int{6}
}

func (x *BroadcastEndpointResponse) GetIP() []string {
	if x != nil {
		return x.IP
	}
	return nil
}

var File_l2Controller_proto protoreflect.FileDescriptor

var file_l2Controller_proto_rawDesc = []byte{
	0x0a, 0x12, 0x6c, 0x32, 0x43, 0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c, 0x6c, 0x65, 0x72, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x12, 0x10, 0x76, 0x70, 0x63, 0x2e, 0x6c, 0x32, 0x43, 0x6f, 0x6e, 0x74,
	0x72, 0x6f, 0x6c, 0x6c, 0x65, 0x72, 0x22, 0x58, 0x0a, 0x0c, 0x4d, 0x41, 0x43, 0x49, 0x50, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x12, 0x0a, 0x04, 0x56, 0x4e, 0x49, 0x44, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x0d, 0x52, 0x04, 0x56, 0x4e, 0x49, 0x44, 0x12, 0x12, 0x0a, 0x04, 0x56, 0x4c,
	0x41, 0x4e, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x04, 0x56, 0x4c, 0x41, 0x4e, 0x12, 0x10,
	0x0a, 0x03, 0x4d, 0x41, 0x43, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x4d, 0x41, 0x43,
	0x12, 0x0e, 0x0a, 0x02, 0x49, 0x50, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x49, 0x50,
	0x22, 0x0b, 0x0a, 0x09, 0x4d, 0x41, 0x43, 0x49, 0x50, 0x52, 0x65, 0x73, 0x70, 0x22, 0x97, 0x01,
	0x0a, 0x0d, 0x4c, 0x6f, 0x6f, 0x6b, 0x75, 0x70, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12,
	0x3c, 0x0a, 0x0a, 0x4c, 0x6f, 0x6f, 0x6b, 0x75, 0x70, 0x54, 0x79, 0x70, 0x65, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x0e, 0x32, 0x1c, 0x2e, 0x76, 0x70, 0x63, 0x2e, 0x6c, 0x32, 0x43, 0x6f, 0x6e, 0x74,
	0x72, 0x6f, 0x6c, 0x6c, 0x65, 0x72, 0x2e, 0x4c, 0x6f, 0x6f, 0x6b, 0x75, 0x70, 0x54, 0x79, 0x70,
	0x65, 0x52, 0x0a, 0x4c, 0x6f, 0x6f, 0x6b, 0x75, 0x70, 0x54, 0x79, 0x70, 0x65, 0x12, 0x12, 0x0a,
	0x04, 0x56, 0x4e, 0x49, 0x44, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x04, 0x56, 0x4e, 0x49,
	0x44, 0x12, 0x12, 0x0a, 0x04, 0x56, 0x4c, 0x41, 0x4e, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0d, 0x52,
	0x04, 0x56, 0x4c, 0x41, 0x4e, 0x12, 0x10, 0x0a, 0x03, 0x4d, 0x41, 0x43, 0x18, 0x04, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x03, 0x4d, 0x41, 0x43, 0x12, 0x0e, 0x0a, 0x02, 0x49, 0x50, 0x18, 0x05, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x02, 0x49, 0x50, 0x22, 0x32, 0x0a, 0x0e, 0x4c, 0x6f, 0x6f, 0x6b, 0x75,
	0x70, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x10, 0x0a, 0x03, 0x4d, 0x41, 0x43,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x4d, 0x41, 0x43, 0x12, 0x0e, 0x0a, 0x02, 0x49,
	0x50, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x49, 0x50, 0x22, 0x21, 0x0a, 0x0b, 0x56,
	0x4e, 0x49, 0x44, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x12, 0x0a, 0x04, 0x56, 0x4e,
	0x49, 0x44, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x04, 0x56, 0x4e, 0x49, 0x44, 0x22, 0x0d,
	0x0a, 0x0b, 0x52, 0x65, 0x67, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x2b, 0x0a,
	0x19, 0x42, 0x72, 0x6f, 0x61, 0x64, 0x63, 0x61, 0x73, 0x74, 0x45, 0x6e, 0x64, 0x70, 0x6f, 0x69,
	0x6e, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x0e, 0x0a, 0x02, 0x49, 0x50,
	0x18, 0x01, 0x20, 0x03, 0x28, 0x09, 0x52, 0x02, 0x49, 0x50, 0x2a, 0x1d, 0x0a, 0x0a, 0x4c, 0x6f,
	0x6f, 0x6b, 0x75, 0x70, 0x54, 0x79, 0x70, 0x65, 0x12, 0x07, 0x0a, 0x03, 0x4d, 0x41, 0x43, 0x10,
	0x00, 0x12, 0x06, 0x0a, 0x02, 0x49, 0x50, 0x10, 0x01, 0x32, 0xda, 0x04, 0x0a, 0x11, 0x43, 0x6f,
	0x6e, 0x74, 0x72, 0x6f, 0x6c, 0x6c, 0x65, 0x72, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12,
	0x4e, 0x0a, 0x0d, 0x52, 0x65, 0x67, 0x69, 0x73, 0x74, 0x65, 0x72, 0x4d, 0x61, 0x63, 0x49, 0x50,
	0x12, 0x1e, 0x2e, 0x76, 0x70, 0x63, 0x2e, 0x6c, 0x32, 0x43, 0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c,
	0x6c, 0x65, 0x72, 0x2e, 0x4d, 0x41, 0x43, 0x49, 0x50, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x1a, 0x1b, 0x2e, 0x76, 0x70, 0x63, 0x2e, 0x6c, 0x32, 0x43, 0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c,
	0x6c, 0x65, 0x72, 0x2e, 0x4d, 0x41, 0x43, 0x49, 0x50, 0x52, 0x65, 0x73, 0x70, 0x22, 0x00, 0x12,
	0x50, 0x0a, 0x0f, 0x44, 0x65, 0x72, 0x65, 0x67, 0x69, 0x73, 0x74, 0x65, 0x72, 0x4d, 0x61, 0x63,
	0x49, 0x50, 0x12, 0x1e, 0x2e, 0x76, 0x70, 0x63, 0x2e, 0x6c, 0x32, 0x43, 0x6f, 0x6e, 0x74, 0x72,
	0x6f, 0x6c, 0x6c, 0x65, 0x72, 0x2e, 0x4d, 0x41, 0x43, 0x49, 0x50, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x1a, 0x1b, 0x2e, 0x76, 0x70, 0x63, 0x2e, 0x6c, 0x32, 0x43, 0x6f, 0x6e, 0x74, 0x72,
	0x6f, 0x6c, 0x6c, 0x65, 0x72, 0x2e, 0x4d, 0x41, 0x43, 0x49, 0x50, 0x52, 0x65, 0x73, 0x70, 0x22,
	0x00, 0x12, 0x4f, 0x0a, 0x08, 0x4c, 0x6f, 0x6f, 0x6b, 0x75, 0x70, 0x49, 0x50, 0x12, 0x1f, 0x2e,
	0x76, 0x70, 0x63, 0x2e, 0x6c, 0x32, 0x43, 0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c, 0x6c, 0x65, 0x72,
	0x2e, 0x4c, 0x6f, 0x6f, 0x6b, 0x75, 0x70, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x20,
	0x2e, 0x76, 0x70, 0x63, 0x2e, 0x6c, 0x32, 0x43, 0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c, 0x6c, 0x65,
	0x72, 0x2e, 0x4c, 0x6f, 0x6f, 0x6b, 0x75, 0x70, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x22, 0x00, 0x12, 0x50, 0x0a, 0x09, 0x4c, 0x6f, 0x6f, 0x6b, 0x75, 0x70, 0x4d, 0x61, 0x63, 0x12,
	0x1f, 0x2e, 0x76, 0x70, 0x63, 0x2e, 0x6c, 0x32, 0x43, 0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c, 0x6c,
	0x65, 0x72, 0x2e, 0x4c, 0x6f, 0x6f, 0x6b, 0x75, 0x70, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x1a, 0x20, 0x2e, 0x76, 0x70, 0x63, 0x2e, 0x6c, 0x32, 0x43, 0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c,
	0x6c, 0x65, 0x72, 0x2e, 0x4c, 0x6f, 0x6f, 0x6b, 0x75, 0x70, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e,
	0x73, 0x65, 0x22, 0x00, 0x12, 0x62, 0x0a, 0x12, 0x42, 0x72, 0x6f, 0x61, 0x64, 0x63, 0x61, 0x73,
	0x74, 0x45, 0x6e, 0x64, 0x70, 0x6f, 0x69, 0x6e, 0x74, 0x73, 0x12, 0x1d, 0x2e, 0x76, 0x70, 0x63,
	0x2e, 0x6c, 0x32, 0x43, 0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c, 0x6c, 0x65, 0x72, 0x2e, 0x56, 0x4e,
	0x49, 0x44, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x2b, 0x2e, 0x76, 0x70, 0x63, 0x2e,
	0x6c, 0x32, 0x43, 0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c, 0x6c, 0x65, 0x72, 0x2e, 0x42, 0x72, 0x6f,
	0x61, 0x64, 0x63, 0x61, 0x73, 0x74, 0x45, 0x6e, 0x64, 0x70, 0x6f, 0x69, 0x6e, 0x74, 0x52, 0x65,
	0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x12, 0x4c, 0x0a, 0x0a, 0x52, 0x65, 0x67, 0x69,
	0x73, 0x74, 0x65, 0x72, 0x45, 0x50, 0x12, 0x1d, 0x2e, 0x76, 0x70, 0x63, 0x2e, 0x6c, 0x32, 0x43,
	0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c, 0x6c, 0x65, 0x72, 0x2e, 0x56, 0x4e, 0x49, 0x44, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1d, 0x2e, 0x76, 0x70, 0x63, 0x2e, 0x6c, 0x32, 0x43, 0x6f,
	0x6e, 0x74, 0x72, 0x6f, 0x6c, 0x6c, 0x65, 0x72, 0x2e, 0x52, 0x65, 0x67, 0x52, 0x65, 0x73, 0x70,
	0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x12, 0x4e, 0x0a, 0x0c, 0x44, 0x65, 0x72, 0x65, 0x67, 0x69,
	0x73, 0x74, 0x65, 0x72, 0x45, 0x50, 0x12, 0x1d, 0x2e, 0x76, 0x70, 0x63, 0x2e, 0x6c, 0x32, 0x43,
	0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c, 0x6c, 0x65, 0x72, 0x2e, 0x56, 0x4e, 0x49, 0x44, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1d, 0x2e, 0x76, 0x70, 0x63, 0x2e, 0x6c, 0x32, 0x43, 0x6f,
	0x6e, 0x74, 0x72, 0x6f, 0x6c, 0x6c, 0x65, 0x72, 0x2e, 0x52, 0x65, 0x67, 0x52, 0x65, 0x73, 0x70,
	0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x42, 0x10, 0x5a, 0x0e, 0x2e, 0x3b, 0x6c, 0x32, 0x43, 0x6f,
	0x6e, 0x74, 0x72, 0x6f, 0x6c, 0x6c, 0x65, 0x72, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_l2Controller_proto_rawDescOnce sync.Once
	file_l2Controller_proto_rawDescData = file_l2Controller_proto_rawDesc
)

func file_l2Controller_proto_rawDescGZIP() []byte {
	file_l2Controller_proto_rawDescOnce.Do(func() {
		file_l2Controller_proto_rawDescData = protoimpl.X.CompressGZIP(file_l2Controller_proto_rawDescData)
	})
	return file_l2Controller_proto_rawDescData
}

var file_l2Controller_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_l2Controller_proto_msgTypes = make([]protoimpl.MessageInfo, 7)
var file_l2Controller_proto_goTypes = []interface{}{
	(LookupType)(0),                   // 0: vpc.l2Controller.LookupType
	(*MACIPRequest)(nil),              // 1: vpc.l2Controller.MACIPRequest
	(*MACIPResp)(nil),                 // 2: vpc.l2Controller.MACIPResp
	(*LookupRequest)(nil),             // 3: vpc.l2Controller.LookupRequest
	(*LookupResponse)(nil),            // 4: vpc.l2Controller.LookupResponse
	(*VNIDRequest)(nil),               // 5: vpc.l2Controller.VNIDRequest
	(*RegResponse)(nil),               // 6: vpc.l2Controller.RegResponse
	(*BroadcastEndpointResponse)(nil), // 7: vpc.l2Controller.BroadcastEndpointResponse
}
var file_l2Controller_proto_depIdxs = []int32{
	0, // 0: vpc.l2Controller.LookupRequest.LookupType:type_name -> vpc.l2Controller.LookupType
	1, // 1: vpc.l2Controller.ControllerService.RegisterMacIP:input_type -> vpc.l2Controller.MACIPRequest
	1, // 2: vpc.l2Controller.ControllerService.DeregisterMacIP:input_type -> vpc.l2Controller.MACIPRequest
	3, // 3: vpc.l2Controller.ControllerService.LookupIP:input_type -> vpc.l2Controller.LookupRequest
	3, // 4: vpc.l2Controller.ControllerService.LookupMac:input_type -> vpc.l2Controller.LookupRequest
	5, // 5: vpc.l2Controller.ControllerService.BroadcastEndpoints:input_type -> vpc.l2Controller.VNIDRequest
	5, // 6: vpc.l2Controller.ControllerService.RegisterEP:input_type -> vpc.l2Controller.VNIDRequest
	5, // 7: vpc.l2Controller.ControllerService.DeregisterEP:input_type -> vpc.l2Controller.VNIDRequest
	2, // 8: vpc.l2Controller.ControllerService.RegisterMacIP:output_type -> vpc.l2Controller.MACIPResp
	2, // 9: vpc.l2Controller.ControllerService.DeregisterMacIP:output_type -> vpc.l2Controller.MACIPResp
	4, // 10: vpc.l2Controller.ControllerService.LookupIP:output_type -> vpc.l2Controller.LookupResponse
	4, // 11: vpc.l2Controller.ControllerService.LookupMac:output_type -> vpc.l2Controller.LookupResponse
	7, // 12: vpc.l2Controller.ControllerService.BroadcastEndpoints:output_type -> vpc.l2Controller.BroadcastEndpointResponse
	6, // 13: vpc.l2Controller.ControllerService.RegisterEP:output_type -> vpc.l2Controller.RegResponse
	6, // 14: vpc.l2Controller.ControllerService.DeregisterEP:output_type -> vpc.l2Controller.RegResponse
	8, // [8:15] is the sub-list for method output_type
	1, // [1:8] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_l2Controller_proto_init() }
func file_l2Controller_proto_init() {
	if File_l2Controller_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_l2Controller_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MACIPRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_l2Controller_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MACIPResp); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_l2Controller_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*LookupRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_l2Controller_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*LookupResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_l2Controller_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*VNIDRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_l2Controller_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RegResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_l2Controller_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*BroadcastEndpointResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_l2Controller_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   7,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_l2Controller_proto_goTypes,
		DependencyIndexes: file_l2Controller_proto_depIdxs,
		EnumInfos:         file_l2Controller_proto_enumTypes,
		MessageInfos:      file_l2Controller_proto_msgTypes,
	}.Build()
	File_l2Controller_proto = out.File
	file_l2Controller_proto_rawDesc = nil
	file_l2Controller_proto_goTypes = nil
	file_l2Controller_proto_depIdxs = nil
}
