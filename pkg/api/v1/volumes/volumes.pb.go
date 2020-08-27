// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0-devel
// 	protoc        v3.6.1
// source: volumes.proto

package volumes

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

type VolumeStatus int32

const (
	VolumeStatus_PENDING   VolumeStatus = 0
	VolumeStatus_CREATING  VolumeStatus = 1
	VolumeStatus_AVAILABLE VolumeStatus = 2
	VolumeStatus_INUSE     VolumeStatus = 3
	VolumeStatus_BACKINGUP VolumeStatus = 4
	VolumeStatus_RESTORING VolumeStatus = 5
	VolumeStatus_RESIZING  VolumeStatus = 6
	VolumeStatus_DELETING  VolumeStatus = 7
	VolumeStatus_DELETED   VolumeStatus = 8
	VolumeStatus_ERROR     VolumeStatus = 99
)

// Enum value maps for VolumeStatus.
var (
	VolumeStatus_name = map[int32]string{
		0:  "PENDING",
		1:  "CREATING",
		2:  "AVAILABLE",
		3:  "INUSE",
		4:  "BACKINGUP",
		5:  "RESTORING",
		6:  "RESIZING",
		7:  "DELETING",
		8:  "DELETED",
		99: "ERROR",
	}
	VolumeStatus_value = map[string]int32{
		"PENDING":   0,
		"CREATING":  1,
		"AVAILABLE": 2,
		"INUSE":     3,
		"BACKINGUP": 4,
		"RESTORING": 5,
		"RESIZING":  6,
		"DELETING":  7,
		"DELETED":   8,
		"ERROR":     99,
	}
)

func (x VolumeStatus) Enum() *VolumeStatus {
	p := new(VolumeStatus)
	*p = x
	return p
}

func (x VolumeStatus) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (VolumeStatus) Descriptor() protoreflect.EnumDescriptor {
	return file_volumes_proto_enumTypes[0].Descriptor()
}

func (VolumeStatus) Type() protoreflect.EnumType {
	return &file_volumes_proto_enumTypes[0]
}

func (x VolumeStatus) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use VolumeStatus.Descriptor instead.
func (VolumeStatus) EnumDescriptor() ([]byte, []int) {
	return file_volumes_proto_rawDescGZIP(), []int{0}
}

type BlockCommandRequest_Cmd int32

const (
	BlockCommandRequest_WRITE BlockCommandRequest_Cmd = 0
	BlockCommandRequest_READ  BlockCommandRequest_Cmd = 1
	BlockCommandRequest_ZERO  BlockCommandRequest_Cmd = 2
)

// Enum value maps for BlockCommandRequest_Cmd.
var (
	BlockCommandRequest_Cmd_name = map[int32]string{
		0: "WRITE",
		1: "READ",
		2: "ZERO",
	}
	BlockCommandRequest_Cmd_value = map[string]int32{
		"WRITE": 0,
		"READ":  1,
		"ZERO":  2,
	}
)

func (x BlockCommandRequest_Cmd) Enum() *BlockCommandRequest_Cmd {
	p := new(BlockCommandRequest_Cmd)
	*p = x
	return p
}

func (x BlockCommandRequest_Cmd) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (BlockCommandRequest_Cmd) Descriptor() protoreflect.EnumDescriptor {
	return file_volumes_proto_enumTypes[1].Descriptor()
}

func (BlockCommandRequest_Cmd) Type() protoreflect.EnumType {
	return &file_volumes_proto_enumTypes[1]
}

func (x BlockCommandRequest_Cmd) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use BlockCommandRequest_Cmd.Descriptor instead.
func (BlockCommandRequest_Cmd) EnumDescriptor() ([]byte, []int) {
	return file_volumes_proto_rawDescGZIP(), []int{5, 0}
}

type Volume struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Account          int32             `protobuf:"varint,1,opt,name=account,proto3" json:"account,omitempty"`
	Id               string            `protobuf:"bytes,2,opt,name=id,proto3" json:"id,omitempty"`
	Status           VolumeStatus      `protobuf:"varint,3,opt,name=status,proto3,enum=vpc.volumes.VolumeStatus" json:"status,omitempty"`
	Metadata         map[string][]byte `protobuf:"bytes,4,rep,name=metadata,proto3" json:"metadata,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	CreatedAt        string            `protobuf:"bytes,5,opt,name=created_at,json=createdAt,proto3" json:"created_at,omitempty"`
	UpdatedAt        string            `protobuf:"bytes,6,opt,name=updated_at,json=updatedAt,proto3" json:"updated_at,omitempty"`
	Size             int64             `protobuf:"varint,7,opt,name=size,proto3" json:"size,omitempty"`
	Store            string            `protobuf:"bytes,8,opt,name=store,proto3" json:"store,omitempty"`
	AvailabilityArea string            `protobuf:"bytes,9,opt,name=availability_area,json=availabilityArea,proto3" json:"availability_area,omitempty"`
}

func (x *Volume) Reset() {
	*x = Volume{}
	if protoimpl.UnsafeEnabled {
		mi := &file_volumes_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Volume) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Volume) ProtoMessage() {}

func (x *Volume) ProtoReflect() protoreflect.Message {
	mi := &file_volumes_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Volume.ProtoReflect.Descriptor instead.
func (*Volume) Descriptor() ([]byte, []int) {
	return file_volumes_proto_rawDescGZIP(), []int{0}
}

func (x *Volume) GetAccount() int32 {
	if x != nil {
		return x.Account
	}
	return 0
}

func (x *Volume) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *Volume) GetStatus() VolumeStatus {
	if x != nil {
		return x.Status
	}
	return VolumeStatus_PENDING
}

func (x *Volume) GetMetadata() map[string][]byte {
	if x != nil {
		return x.Metadata
	}
	return nil
}

func (x *Volume) GetCreatedAt() string {
	if x != nil {
		return x.CreatedAt
	}
	return ""
}

func (x *Volume) GetUpdatedAt() string {
	if x != nil {
		return x.UpdatedAt
	}
	return ""
}

func (x *Volume) GetSize() int64 {
	if x != nil {
		return x.Size
	}
	return 0
}

func (x *Volume) GetStore() string {
	if x != nil {
		return x.Store
	}
	return ""
}

func (x *Volume) GetAvailabilityArea() string {
	if x != nil {
		return x.AvailabilityArea
	}
	return ""
}

type Snapshot struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id               string            `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Volume           string            `protobuf:"bytes,2,opt,name=volume,proto3" json:"volume,omitempty"`
	Status           VolumeStatus      `protobuf:"varint,3,opt,name=status,proto3,enum=vpc.volumes.VolumeStatus" json:"status,omitempty"`
	Metadata         map[string][]byte `protobuf:"bytes,4,rep,name=metadata,proto3" json:"metadata,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	CreatedAt        string            `protobuf:"bytes,5,opt,name=created_at,json=createdAt,proto3" json:"created_at,omitempty"`
	UpdatedAt        string            `protobuf:"bytes,6,opt,name=updated_at,json=updatedAt,proto3" json:"updated_at,omitempty"`
	Size             int64             `protobuf:"varint,7,opt,name=size,proto3" json:"size,omitempty"`
	Store            string            `protobuf:"bytes,8,opt,name=store,proto3" json:"store,omitempty"`
	AvailabilityArea string            `protobuf:"bytes,9,opt,name=availability_area,json=availabilityArea,proto3" json:"availability_area,omitempty"`
}

func (x *Snapshot) Reset() {
	*x = Snapshot{}
	if protoimpl.UnsafeEnabled {
		mi := &file_volumes_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Snapshot) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Snapshot) ProtoMessage() {}

func (x *Snapshot) ProtoReflect() protoreflect.Message {
	mi := &file_volumes_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Snapshot.ProtoReflect.Descriptor instead.
func (*Snapshot) Descriptor() ([]byte, []int) {
	return file_volumes_proto_rawDescGZIP(), []int{1}
}

func (x *Snapshot) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *Snapshot) GetVolume() string {
	if x != nil {
		return x.Volume
	}
	return ""
}

func (x *Snapshot) GetStatus() VolumeStatus {
	if x != nil {
		return x.Status
	}
	return VolumeStatus_PENDING
}

func (x *Snapshot) GetMetadata() map[string][]byte {
	if x != nil {
		return x.Metadata
	}
	return nil
}

func (x *Snapshot) GetCreatedAt() string {
	if x != nil {
		return x.CreatedAt
	}
	return ""
}

func (x *Snapshot) GetUpdatedAt() string {
	if x != nil {
		return x.UpdatedAt
	}
	return ""
}

func (x *Snapshot) GetSize() int64 {
	if x != nil {
		return x.Size
	}
	return 0
}

func (x *Snapshot) GetStore() string {
	if x != nil {
		return x.Store
	}
	return ""
}

func (x *Snapshot) GetAvailabilityArea() string {
	if x != nil {
		return x.AvailabilityArea
	}
	return ""
}

type PeerNegotiate struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	IAm    string `protobuf:"bytes,1,opt,name=IAm,proto3" json:"IAm,omitempty"`
	YouAre string `protobuf:"bytes,2,opt,name=YouAre,proto3" json:"YouAre,omitempty"`
}

func (x *PeerNegotiate) Reset() {
	*x = PeerNegotiate{}
	if protoimpl.UnsafeEnabled {
		mi := &file_volumes_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PeerNegotiate) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PeerNegotiate) ProtoMessage() {}

func (x *PeerNegotiate) ProtoReflect() protoreflect.Message {
	mi := &file_volumes_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PeerNegotiate.ProtoReflect.Descriptor instead.
func (*PeerNegotiate) Descriptor() ([]byte, []int) {
	return file_volumes_proto_rawDescGZIP(), []int{2}
}

func (x *PeerNegotiate) GetIAm() string {
	if x != nil {
		return x.IAm
	}
	return ""
}

func (x *PeerNegotiate) GetYouAre() string {
	if x != nil {
		return x.YouAre
	}
	return ""
}

type RaftRPC struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Volume  string `protobuf:"bytes,1,opt,name=volume,proto3" json:"volume,omitempty"`
	Type    int32  `protobuf:"varint,2,opt,name=type,proto3" json:"type,omitempty"`
	Command []byte `protobuf:"bytes,3,opt,name=command,proto3" json:"command,omitempty"`
	Error   string `protobuf:"bytes,4,opt,name=error,proto3" json:"error,omitempty"`
}

func (x *RaftRPC) Reset() {
	*x = RaftRPC{}
	if protoimpl.UnsafeEnabled {
		mi := &file_volumes_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RaftRPC) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RaftRPC) ProtoMessage() {}

func (x *RaftRPC) ProtoReflect() protoreflect.Message {
	mi := &file_volumes_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RaftRPC.ProtoReflect.Descriptor instead.
func (*RaftRPC) Descriptor() ([]byte, []int) {
	return file_volumes_proto_rawDescGZIP(), []int{3}
}

func (x *RaftRPC) GetVolume() string {
	if x != nil {
		return x.Volume
	}
	return ""
}

func (x *RaftRPC) GetType() int32 {
	if x != nil {
		return x.Type
	}
	return 0
}

func (x *RaftRPC) GetCommand() []byte {
	if x != nil {
		return x.Command
	}
	return nil
}

func (x *RaftRPC) GetError() string {
	if x != nil {
		return x.Error
	}
	return ""
}

type PeerPing struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Ts int64 `protobuf:"varint,1,opt,name=ts,proto3" json:"ts,omitempty"`
}

func (x *PeerPing) Reset() {
	*x = PeerPing{}
	if protoimpl.UnsafeEnabled {
		mi := &file_volumes_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PeerPing) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PeerPing) ProtoMessage() {}

func (x *PeerPing) ProtoReflect() protoreflect.Message {
	mi := &file_volumes_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PeerPing.ProtoReflect.Descriptor instead.
func (*PeerPing) Descriptor() ([]byte, []int) {
	return file_volumes_proto_rawDescGZIP(), []int{4}
}

func (x *PeerPing) GetTs() int64 {
	if x != nil {
		return x.Ts
	}
	return 0
}

type BlockCommandRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Volume string                  `protobuf:"bytes,1,opt,name=volume,proto3" json:"volume,omitempty"`
	Cmd    BlockCommandRequest_Cmd `protobuf:"varint,2,opt,name=cmd,proto3,enum=vpc.volumes.BlockCommandRequest_Cmd" json:"cmd,omitempty"`
	Offset int64                   `protobuf:"varint,3,opt,name=offset,proto3" json:"offset,omitempty"`
	Length int32                   `protobuf:"varint,4,opt,name=length,proto3" json:"length,omitempty"`
	Data   []byte                  `protobuf:"bytes,5,opt,name=data,proto3" json:"data,omitempty"`
}

func (x *BlockCommandRequest) Reset() {
	*x = BlockCommandRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_volumes_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *BlockCommandRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BlockCommandRequest) ProtoMessage() {}

func (x *BlockCommandRequest) ProtoReflect() protoreflect.Message {
	mi := &file_volumes_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BlockCommandRequest.ProtoReflect.Descriptor instead.
func (*BlockCommandRequest) Descriptor() ([]byte, []int) {
	return file_volumes_proto_rawDescGZIP(), []int{5}
}

func (x *BlockCommandRequest) GetVolume() string {
	if x != nil {
		return x.Volume
	}
	return ""
}

func (x *BlockCommandRequest) GetCmd() BlockCommandRequest_Cmd {
	if x != nil {
		return x.Cmd
	}
	return BlockCommandRequest_WRITE
}

func (x *BlockCommandRequest) GetOffset() int64 {
	if x != nil {
		return x.Offset
	}
	return 0
}

func (x *BlockCommandRequest) GetLength() int32 {
	if x != nil {
		return x.Length
	}
	return 0
}

func (x *BlockCommandRequest) GetData() []byte {
	if x != nil {
		return x.Data
	}
	return nil
}

type BlockCommandResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Volume  string `protobuf:"bytes,1,opt,name=volume,proto3" json:"volume,omitempty"`
	RetryAt string `protobuf:"bytes,2,opt,name=retryAt,proto3" json:"retryAt,omitempty"` //peer ID to retry to request at - i.e. peer not leader
	Data    []byte `protobuf:"bytes,3,opt,name=data,proto3" json:"data,omitempty"`
	Error   string `protobuf:"bytes,4,opt,name=error,proto3" json:"error,omitempty"`
}

func (x *BlockCommandResponse) Reset() {
	*x = BlockCommandResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_volumes_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *BlockCommandResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BlockCommandResponse) ProtoMessage() {}

func (x *BlockCommandResponse) ProtoReflect() protoreflect.Message {
	mi := &file_volumes_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BlockCommandResponse.ProtoReflect.Descriptor instead.
func (*BlockCommandResponse) Descriptor() ([]byte, []int) {
	return file_volumes_proto_rawDescGZIP(), []int{6}
}

func (x *BlockCommandResponse) GetVolume() string {
	if x != nil {
		return x.Volume
	}
	return ""
}

func (x *BlockCommandResponse) GetRetryAt() string {
	if x != nil {
		return x.RetryAt
	}
	return ""
}

func (x *BlockCommandResponse) GetData() []byte {
	if x != nil {
		return x.Data
	}
	return nil
}

func (x *BlockCommandResponse) GetError() string {
	if x != nil {
		return x.Error
	}
	return ""
}

type RPCMetadata struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	LeaderOnly      bool  `protobuf:"varint,1,opt,name=leaderOnly,proto3" json:"leaderOnly,omitempty"` //to prevent stale reads if no caching on client side - all writes must still go to leader
	AllowForwarding bool  `protobuf:"varint,2,opt,name=allowForwarding,proto3" json:"allowForwarding,omitempty"`
	TTL             int32 `protobuf:"varint,3,opt,name=TTL,proto3" json:"TTL,omitempty"`
}

func (x *RPCMetadata) Reset() {
	*x = RPCMetadata{}
	if protoimpl.UnsafeEnabled {
		mi := &file_volumes_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RPCMetadata) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RPCMetadata) ProtoMessage() {}

func (x *RPCMetadata) ProtoReflect() protoreflect.Message {
	mi := &file_volumes_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RPCMetadata.ProtoReflect.Descriptor instead.
func (*RPCMetadata) Descriptor() ([]byte, []int) {
	return file_volumes_proto_rawDescGZIP(), []int{7}
}

func (x *RPCMetadata) GetLeaderOnly() bool {
	if x != nil {
		return x.LeaderOnly
	}
	return false
}

func (x *RPCMetadata) GetAllowForwarding() bool {
	if x != nil {
		return x.AllowForwarding
	}
	return false
}

func (x *RPCMetadata) GetTTL() int32 {
	if x != nil {
		return x.TTL
	}
	return 0
}

type PeerRPC struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Metadata *RPCMetadata `protobuf:"bytes,1,opt,name=metadata,proto3" json:"metadata,omitempty"`
	// Types that are assignable to RPC:
	//	*PeerRPC_RaftRPC
	//	*PeerRPC_PeerPing
	//	*PeerRPC_BlockCommand
	//	*PeerRPC_BlockCommandResponse
	RPC   isPeerRPC_RPC `protobuf_oneof:"RPC"`
	Async bool          `protobuf:"varint,101,opt,name=async,proto3" json:"async,omitempty"`
}

func (x *PeerRPC) Reset() {
	*x = PeerRPC{}
	if protoimpl.UnsafeEnabled {
		mi := &file_volumes_proto_msgTypes[8]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PeerRPC) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PeerRPC) ProtoMessage() {}

func (x *PeerRPC) ProtoReflect() protoreflect.Message {
	mi := &file_volumes_proto_msgTypes[8]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PeerRPC.ProtoReflect.Descriptor instead.
func (*PeerRPC) Descriptor() ([]byte, []int) {
	return file_volumes_proto_rawDescGZIP(), []int{8}
}

func (x *PeerRPC) GetMetadata() *RPCMetadata {
	if x != nil {
		return x.Metadata
	}
	return nil
}

func (m *PeerRPC) GetRPC() isPeerRPC_RPC {
	if m != nil {
		return m.RPC
	}
	return nil
}

func (x *PeerRPC) GetRaftRPC() *RaftRPC {
	if x, ok := x.GetRPC().(*PeerRPC_RaftRPC); ok {
		return x.RaftRPC
	}
	return nil
}

func (x *PeerRPC) GetPeerPing() *PeerPing {
	if x, ok := x.GetRPC().(*PeerRPC_PeerPing); ok {
		return x.PeerPing
	}
	return nil
}

func (x *PeerRPC) GetBlockCommand() *BlockCommandRequest {
	if x, ok := x.GetRPC().(*PeerRPC_BlockCommand); ok {
		return x.BlockCommand
	}
	return nil
}

func (x *PeerRPC) GetBlockCommandResponse() *BlockCommandResponse {
	if x, ok := x.GetRPC().(*PeerRPC_BlockCommandResponse); ok {
		return x.BlockCommandResponse
	}
	return nil
}

func (x *PeerRPC) GetAsync() bool {
	if x != nil {
		return x.Async
	}
	return false
}

type isPeerRPC_RPC interface {
	isPeerRPC_RPC()
}

type PeerRPC_RaftRPC struct {
	RaftRPC *RaftRPC `protobuf:"bytes,2,opt,name=RaftRPC,proto3,oneof"`
}

type PeerRPC_PeerPing struct {
	PeerPing *PeerPing `protobuf:"bytes,3,opt,name=PeerPing,proto3,oneof"`
}

type PeerRPC_BlockCommand struct {
	BlockCommand *BlockCommandRequest `protobuf:"bytes,4,opt,name=BlockCommand,proto3,oneof"`
}

type PeerRPC_BlockCommandResponse struct {
	BlockCommandResponse *BlockCommandResponse `protobuf:"bytes,5,opt,name=BlockCommandResponse,proto3,oneof"`
}

func (*PeerRPC_RaftRPC) isPeerRPC_RPC() {}

func (*PeerRPC_PeerPing) isPeerRPC_RPC() {}

func (*PeerRPC_BlockCommand) isPeerRPC_RPC() {}

func (*PeerRPC_BlockCommandResponse) isPeerRPC_RPC() {}

var File_volumes_proto protoreflect.FileDescriptor

var file_volumes_proto_rawDesc = []byte{
	0x0a, 0x0d, 0x76, 0x6f, 0x6c, 0x75, 0x6d, 0x65, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12,
	0x0b, 0x76, 0x70, 0x63, 0x2e, 0x76, 0x6f, 0x6c, 0x75, 0x6d, 0x65, 0x73, 0x22, 0xf6, 0x02, 0x0a,
	0x06, 0x56, 0x6f, 0x6c, 0x75, 0x6d, 0x65, 0x12, 0x18, 0x0a, 0x07, 0x61, 0x63, 0x63, 0x6f, 0x75,
	0x6e, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x05, 0x52, 0x07, 0x61, 0x63, 0x63, 0x6f, 0x75, 0x6e,
	0x74, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69,
	0x64, 0x12, 0x31, 0x0a, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x18, 0x03, 0x20, 0x01, 0x28,
	0x0e, 0x32, 0x19, 0x2e, 0x76, 0x70, 0x63, 0x2e, 0x76, 0x6f, 0x6c, 0x75, 0x6d, 0x65, 0x73, 0x2e,
	0x56, 0x6f, 0x6c, 0x75, 0x6d, 0x65, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x52, 0x06, 0x73, 0x74,
	0x61, 0x74, 0x75, 0x73, 0x12, 0x3d, 0x0a, 0x08, 0x6d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61,
	0x18, 0x04, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x21, 0x2e, 0x76, 0x70, 0x63, 0x2e, 0x76, 0x6f, 0x6c,
	0x75, 0x6d, 0x65, 0x73, 0x2e, 0x56, 0x6f, 0x6c, 0x75, 0x6d, 0x65, 0x2e, 0x4d, 0x65, 0x74, 0x61,
	0x64, 0x61, 0x74, 0x61, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x08, 0x6d, 0x65, 0x74, 0x61, 0x64,
	0x61, 0x74, 0x61, 0x12, 0x1d, 0x0a, 0x0a, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x64, 0x5f, 0x61,
	0x74, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x64,
	0x41, 0x74, 0x12, 0x1d, 0x0a, 0x0a, 0x75, 0x70, 0x64, 0x61, 0x74, 0x65, 0x64, 0x5f, 0x61, 0x74,
	0x18, 0x06, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x75, 0x70, 0x64, 0x61, 0x74, 0x65, 0x64, 0x41,
	0x74, 0x12, 0x12, 0x0a, 0x04, 0x73, 0x69, 0x7a, 0x65, 0x18, 0x07, 0x20, 0x01, 0x28, 0x03, 0x52,
	0x04, 0x73, 0x69, 0x7a, 0x65, 0x12, 0x14, 0x0a, 0x05, 0x73, 0x74, 0x6f, 0x72, 0x65, 0x18, 0x08,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x73, 0x74, 0x6f, 0x72, 0x65, 0x12, 0x2b, 0x0a, 0x11, 0x61,
	0x76, 0x61, 0x69, 0x6c, 0x61, 0x62, 0x69, 0x6c, 0x69, 0x74, 0x79, 0x5f, 0x61, 0x72, 0x65, 0x61,
	0x18, 0x09, 0x20, 0x01, 0x28, 0x09, 0x52, 0x10, 0x61, 0x76, 0x61, 0x69, 0x6c, 0x61, 0x62, 0x69,
	0x6c, 0x69, 0x74, 0x79, 0x41, 0x72, 0x65, 0x61, 0x1a, 0x3b, 0x0a, 0x0d, 0x4d, 0x65, 0x74, 0x61,
	0x64, 0x61, 0x74, 0x61, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x14, 0x0a, 0x05, 0x76,
	0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75,
	0x65, 0x3a, 0x02, 0x38, 0x01, 0x22, 0xf8, 0x02, 0x0a, 0x08, 0x53, 0x6e, 0x61, 0x70, 0x73, 0x68,
	0x6f, 0x74, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02,
	0x69, 0x64, 0x12, 0x16, 0x0a, 0x06, 0x76, 0x6f, 0x6c, 0x75, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x06, 0x76, 0x6f, 0x6c, 0x75, 0x6d, 0x65, 0x12, 0x31, 0x0a, 0x06, 0x73, 0x74,
	0x61, 0x74, 0x75, 0x73, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x19, 0x2e, 0x76, 0x70, 0x63,
	0x2e, 0x76, 0x6f, 0x6c, 0x75, 0x6d, 0x65, 0x73, 0x2e, 0x56, 0x6f, 0x6c, 0x75, 0x6d, 0x65, 0x53,
	0x74, 0x61, 0x74, 0x75, 0x73, 0x52, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x3f, 0x0a,
	0x08, 0x6d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x18, 0x04, 0x20, 0x03, 0x28, 0x0b, 0x32,
	0x23, 0x2e, 0x76, 0x70, 0x63, 0x2e, 0x76, 0x6f, 0x6c, 0x75, 0x6d, 0x65, 0x73, 0x2e, 0x53, 0x6e,
	0x61, 0x70, 0x73, 0x68, 0x6f, 0x74, 0x2e, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x45,
	0x6e, 0x74, 0x72, 0x79, 0x52, 0x08, 0x6d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x12, 0x1d,
	0x0a, 0x0a, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x64, 0x5f, 0x61, 0x74, 0x18, 0x05, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x09, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x64, 0x41, 0x74, 0x12, 0x1d, 0x0a,
	0x0a, 0x75, 0x70, 0x64, 0x61, 0x74, 0x65, 0x64, 0x5f, 0x61, 0x74, 0x18, 0x06, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x09, 0x75, 0x70, 0x64, 0x61, 0x74, 0x65, 0x64, 0x41, 0x74, 0x12, 0x12, 0x0a, 0x04,
	0x73, 0x69, 0x7a, 0x65, 0x18, 0x07, 0x20, 0x01, 0x28, 0x03, 0x52, 0x04, 0x73, 0x69, 0x7a, 0x65,
	0x12, 0x14, 0x0a, 0x05, 0x73, 0x74, 0x6f, 0x72, 0x65, 0x18, 0x08, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x05, 0x73, 0x74, 0x6f, 0x72, 0x65, 0x12, 0x2b, 0x0a, 0x11, 0x61, 0x76, 0x61, 0x69, 0x6c, 0x61,
	0x62, 0x69, 0x6c, 0x69, 0x74, 0x79, 0x5f, 0x61, 0x72, 0x65, 0x61, 0x18, 0x09, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x10, 0x61, 0x76, 0x61, 0x69, 0x6c, 0x61, 0x62, 0x69, 0x6c, 0x69, 0x74, 0x79, 0x41,
	0x72, 0x65, 0x61, 0x1a, 0x3b, 0x0a, 0x0d, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x45,
	0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01,
	0x22, 0x39, 0x0a, 0x0d, 0x50, 0x65, 0x65, 0x72, 0x4e, 0x65, 0x67, 0x6f, 0x74, 0x69, 0x61, 0x74,
	0x65, 0x12, 0x10, 0x0a, 0x03, 0x49, 0x41, 0x6d, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03,
	0x49, 0x41, 0x6d, 0x12, 0x16, 0x0a, 0x06, 0x59, 0x6f, 0x75, 0x41, 0x72, 0x65, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x06, 0x59, 0x6f, 0x75, 0x41, 0x72, 0x65, 0x22, 0x65, 0x0a, 0x07, 0x52,
	0x61, 0x66, 0x74, 0x52, 0x50, 0x43, 0x12, 0x16, 0x0a, 0x06, 0x76, 0x6f, 0x6c, 0x75, 0x6d, 0x65,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x76, 0x6f, 0x6c, 0x75, 0x6d, 0x65, 0x12, 0x12,
	0x0a, 0x04, 0x74, 0x79, 0x70, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x05, 0x52, 0x04, 0x74, 0x79,
	0x70, 0x65, 0x12, 0x18, 0x0a, 0x07, 0x63, 0x6f, 0x6d, 0x6d, 0x61, 0x6e, 0x64, 0x18, 0x03, 0x20,
	0x01, 0x28, 0x0c, 0x52, 0x07, 0x63, 0x6f, 0x6d, 0x6d, 0x61, 0x6e, 0x64, 0x12, 0x14, 0x0a, 0x05,
	0x65, 0x72, 0x72, 0x6f, 0x72, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x65, 0x72, 0x72,
	0x6f, 0x72, 0x22, 0x1a, 0x0a, 0x08, 0x50, 0x65, 0x65, 0x72, 0x50, 0x69, 0x6e, 0x67, 0x12, 0x0e,
	0x0a, 0x02, 0x74, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x52, 0x02, 0x74, 0x73, 0x22, 0xcf,
	0x01, 0x0a, 0x13, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x43, 0x6f, 0x6d, 0x6d, 0x61, 0x6e, 0x64, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x16, 0x0a, 0x06, 0x76, 0x6f, 0x6c, 0x75, 0x6d, 0x65,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x76, 0x6f, 0x6c, 0x75, 0x6d, 0x65, 0x12, 0x36,
	0x0a, 0x03, 0x63, 0x6d, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x24, 0x2e, 0x76, 0x70,
	0x63, 0x2e, 0x76, 0x6f, 0x6c, 0x75, 0x6d, 0x65, 0x73, 0x2e, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x43,
	0x6f, 0x6d, 0x6d, 0x61, 0x6e, 0x64, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x2e, 0x43, 0x6d,
	0x64, 0x52, 0x03, 0x63, 0x6d, 0x64, 0x12, 0x16, 0x0a, 0x06, 0x6f, 0x66, 0x66, 0x73, 0x65, 0x74,
	0x18, 0x03, 0x20, 0x01, 0x28, 0x03, 0x52, 0x06, 0x6f, 0x66, 0x66, 0x73, 0x65, 0x74, 0x12, 0x16,
	0x0a, 0x06, 0x6c, 0x65, 0x6e, 0x67, 0x74, 0x68, 0x18, 0x04, 0x20, 0x01, 0x28, 0x05, 0x52, 0x06,
	0x6c, 0x65, 0x6e, 0x67, 0x74, 0x68, 0x12, 0x12, 0x0a, 0x04, 0x64, 0x61, 0x74, 0x61, 0x18, 0x05,
	0x20, 0x01, 0x28, 0x0c, 0x52, 0x04, 0x64, 0x61, 0x74, 0x61, 0x22, 0x24, 0x0a, 0x03, 0x43, 0x6d,
	0x64, 0x12, 0x09, 0x0a, 0x05, 0x57, 0x52, 0x49, 0x54, 0x45, 0x10, 0x00, 0x12, 0x08, 0x0a, 0x04,
	0x52, 0x45, 0x41, 0x44, 0x10, 0x01, 0x12, 0x08, 0x0a, 0x04, 0x5a, 0x45, 0x52, 0x4f, 0x10, 0x02,
	0x22, 0x72, 0x0a, 0x14, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x43, 0x6f, 0x6d, 0x6d, 0x61, 0x6e, 0x64,
	0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x16, 0x0a, 0x06, 0x76, 0x6f, 0x6c, 0x75,
	0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x76, 0x6f, 0x6c, 0x75, 0x6d, 0x65,
	0x12, 0x18, 0x0a, 0x07, 0x72, 0x65, 0x74, 0x72, 0x79, 0x41, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x07, 0x72, 0x65, 0x74, 0x72, 0x79, 0x41, 0x74, 0x12, 0x12, 0x0a, 0x04, 0x64, 0x61,
	0x74, 0x61, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x04, 0x64, 0x61, 0x74, 0x61, 0x12, 0x14,
	0x0a, 0x05, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x65,
	0x72, 0x72, 0x6f, 0x72, 0x22, 0x69, 0x0a, 0x0b, 0x52, 0x50, 0x43, 0x4d, 0x65, 0x74, 0x61, 0x64,
	0x61, 0x74, 0x61, 0x12, 0x1e, 0x0a, 0x0a, 0x6c, 0x65, 0x61, 0x64, 0x65, 0x72, 0x4f, 0x6e, 0x6c,
	0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x08, 0x52, 0x0a, 0x6c, 0x65, 0x61, 0x64, 0x65, 0x72, 0x4f,
	0x6e, 0x6c, 0x79, 0x12, 0x28, 0x0a, 0x0f, 0x61, 0x6c, 0x6c, 0x6f, 0x77, 0x46, 0x6f, 0x72, 0x77,
	0x61, 0x72, 0x64, 0x69, 0x6e, 0x67, 0x18, 0x02, 0x20, 0x01, 0x28, 0x08, 0x52, 0x0f, 0x61, 0x6c,
	0x6c, 0x6f, 0x77, 0x46, 0x6f, 0x72, 0x77, 0x61, 0x72, 0x64, 0x69, 0x6e, 0x67, 0x12, 0x10, 0x0a,
	0x03, 0x54, 0x54, 0x4c, 0x18, 0x03, 0x20, 0x01, 0x28, 0x05, 0x52, 0x03, 0x54, 0x54, 0x4c, 0x22,
	0xea, 0x02, 0x0a, 0x07, 0x50, 0x65, 0x65, 0x72, 0x52, 0x50, 0x43, 0x12, 0x34, 0x0a, 0x08, 0x6d,
	0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x18, 0x2e,
	0x76, 0x70, 0x63, 0x2e, 0x76, 0x6f, 0x6c, 0x75, 0x6d, 0x65, 0x73, 0x2e, 0x52, 0x50, 0x43, 0x4d,
	0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x52, 0x08, 0x6d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74,
	0x61, 0x12, 0x30, 0x0a, 0x07, 0x52, 0x61, 0x66, 0x74, 0x52, 0x50, 0x43, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x14, 0x2e, 0x76, 0x70, 0x63, 0x2e, 0x76, 0x6f, 0x6c, 0x75, 0x6d, 0x65, 0x73,
	0x2e, 0x52, 0x61, 0x66, 0x74, 0x52, 0x50, 0x43, 0x48, 0x00, 0x52, 0x07, 0x52, 0x61, 0x66, 0x74,
	0x52, 0x50, 0x43, 0x12, 0x33, 0x0a, 0x08, 0x50, 0x65, 0x65, 0x72, 0x50, 0x69, 0x6e, 0x67, 0x18,
	0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x15, 0x2e, 0x76, 0x70, 0x63, 0x2e, 0x76, 0x6f, 0x6c, 0x75,
	0x6d, 0x65, 0x73, 0x2e, 0x50, 0x65, 0x65, 0x72, 0x50, 0x69, 0x6e, 0x67, 0x48, 0x00, 0x52, 0x08,
	0x50, 0x65, 0x65, 0x72, 0x50, 0x69, 0x6e, 0x67, 0x12, 0x46, 0x0a, 0x0c, 0x42, 0x6c, 0x6f, 0x63,
	0x6b, 0x43, 0x6f, 0x6d, 0x6d, 0x61, 0x6e, 0x64, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x20,
	0x2e, 0x76, 0x70, 0x63, 0x2e, 0x76, 0x6f, 0x6c, 0x75, 0x6d, 0x65, 0x73, 0x2e, 0x42, 0x6c, 0x6f,
	0x63, 0x6b, 0x43, 0x6f, 0x6d, 0x6d, 0x61, 0x6e, 0x64, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x48, 0x00, 0x52, 0x0c, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x43, 0x6f, 0x6d, 0x6d, 0x61, 0x6e, 0x64,
	0x12, 0x57, 0x0a, 0x14, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x43, 0x6f, 0x6d, 0x6d, 0x61, 0x6e, 0x64,
	0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x18, 0x05, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x21,
	0x2e, 0x76, 0x70, 0x63, 0x2e, 0x76, 0x6f, 0x6c, 0x75, 0x6d, 0x65, 0x73, 0x2e, 0x42, 0x6c, 0x6f,
	0x63, 0x6b, 0x43, 0x6f, 0x6d, 0x6d, 0x61, 0x6e, 0x64, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x48, 0x00, 0x52, 0x14, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x43, 0x6f, 0x6d, 0x6d, 0x61, 0x6e,
	0x64, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x14, 0x0a, 0x05, 0x61, 0x73, 0x79,
	0x6e, 0x63, 0x18, 0x65, 0x20, 0x01, 0x28, 0x08, 0x52, 0x05, 0x61, 0x73, 0x79, 0x6e, 0x63, 0x42,
	0x05, 0x0a, 0x03, 0x52, 0x50, 0x43, 0x4a, 0x04, 0x08, 0x06, 0x10, 0x65, 0x2a, 0x95, 0x01, 0x0a,
	0x0c, 0x56, 0x6f, 0x6c, 0x75, 0x6d, 0x65, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x0b, 0x0a,
	0x07, 0x50, 0x45, 0x4e, 0x44, 0x49, 0x4e, 0x47, 0x10, 0x00, 0x12, 0x0c, 0x0a, 0x08, 0x43, 0x52,
	0x45, 0x41, 0x54, 0x49, 0x4e, 0x47, 0x10, 0x01, 0x12, 0x0d, 0x0a, 0x09, 0x41, 0x56, 0x41, 0x49,
	0x4c, 0x41, 0x42, 0x4c, 0x45, 0x10, 0x02, 0x12, 0x09, 0x0a, 0x05, 0x49, 0x4e, 0x55, 0x53, 0x45,
	0x10, 0x03, 0x12, 0x0d, 0x0a, 0x09, 0x42, 0x41, 0x43, 0x4b, 0x49, 0x4e, 0x47, 0x55, 0x50, 0x10,
	0x04, 0x12, 0x0d, 0x0a, 0x09, 0x52, 0x45, 0x53, 0x54, 0x4f, 0x52, 0x49, 0x4e, 0x47, 0x10, 0x05,
	0x12, 0x0c, 0x0a, 0x08, 0x52, 0x45, 0x53, 0x49, 0x5a, 0x49, 0x4e, 0x47, 0x10, 0x06, 0x12, 0x0c,
	0x0a, 0x08, 0x44, 0x45, 0x4c, 0x45, 0x54, 0x49, 0x4e, 0x47, 0x10, 0x07, 0x12, 0x0b, 0x0a, 0x07,
	0x44, 0x45, 0x4c, 0x45, 0x54, 0x45, 0x44, 0x10, 0x08, 0x12, 0x09, 0x0a, 0x05, 0x45, 0x52, 0x52,
	0x4f, 0x52, 0x10, 0x63, 0x42, 0x0b, 0x5a, 0x09, 0x2e, 0x3b, 0x76, 0x6f, 0x6c, 0x75, 0x6d, 0x65,
	0x73, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_volumes_proto_rawDescOnce sync.Once
	file_volumes_proto_rawDescData = file_volumes_proto_rawDesc
)

func file_volumes_proto_rawDescGZIP() []byte {
	file_volumes_proto_rawDescOnce.Do(func() {
		file_volumes_proto_rawDescData = protoimpl.X.CompressGZIP(file_volumes_proto_rawDescData)
	})
	return file_volumes_proto_rawDescData
}

var file_volumes_proto_enumTypes = make([]protoimpl.EnumInfo, 2)
var file_volumes_proto_msgTypes = make([]protoimpl.MessageInfo, 11)
var file_volumes_proto_goTypes = []interface{}{
	(VolumeStatus)(0),            // 0: vpc.volumes.VolumeStatus
	(BlockCommandRequest_Cmd)(0), // 1: vpc.volumes.BlockCommandRequest.Cmd
	(*Volume)(nil),               // 2: vpc.volumes.Volume
	(*Snapshot)(nil),             // 3: vpc.volumes.Snapshot
	(*PeerNegotiate)(nil),        // 4: vpc.volumes.PeerNegotiate
	(*RaftRPC)(nil),              // 5: vpc.volumes.RaftRPC
	(*PeerPing)(nil),             // 6: vpc.volumes.PeerPing
	(*BlockCommandRequest)(nil),  // 7: vpc.volumes.BlockCommandRequest
	(*BlockCommandResponse)(nil), // 8: vpc.volumes.BlockCommandResponse
	(*RPCMetadata)(nil),          // 9: vpc.volumes.RPCMetadata
	(*PeerRPC)(nil),              // 10: vpc.volumes.PeerRPC
	nil,                          // 11: vpc.volumes.Volume.MetadataEntry
	nil,                          // 12: vpc.volumes.Snapshot.MetadataEntry
}
var file_volumes_proto_depIdxs = []int32{
	0,  // 0: vpc.volumes.Volume.status:type_name -> vpc.volumes.VolumeStatus
	11, // 1: vpc.volumes.Volume.metadata:type_name -> vpc.volumes.Volume.MetadataEntry
	0,  // 2: vpc.volumes.Snapshot.status:type_name -> vpc.volumes.VolumeStatus
	12, // 3: vpc.volumes.Snapshot.metadata:type_name -> vpc.volumes.Snapshot.MetadataEntry
	1,  // 4: vpc.volumes.BlockCommandRequest.cmd:type_name -> vpc.volumes.BlockCommandRequest.Cmd
	9,  // 5: vpc.volumes.PeerRPC.metadata:type_name -> vpc.volumes.RPCMetadata
	5,  // 6: vpc.volumes.PeerRPC.RaftRPC:type_name -> vpc.volumes.RaftRPC
	6,  // 7: vpc.volumes.PeerRPC.PeerPing:type_name -> vpc.volumes.PeerPing
	7,  // 8: vpc.volumes.PeerRPC.BlockCommand:type_name -> vpc.volumes.BlockCommandRequest
	8,  // 9: vpc.volumes.PeerRPC.BlockCommandResponse:type_name -> vpc.volumes.BlockCommandResponse
	10, // [10:10] is the sub-list for method output_type
	10, // [10:10] is the sub-list for method input_type
	10, // [10:10] is the sub-list for extension type_name
	10, // [10:10] is the sub-list for extension extendee
	0,  // [0:10] is the sub-list for field type_name
}

func init() { file_volumes_proto_init() }
func file_volumes_proto_init() {
	if File_volumes_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_volumes_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Volume); i {
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
		file_volumes_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Snapshot); i {
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
		file_volumes_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PeerNegotiate); i {
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
		file_volumes_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RaftRPC); i {
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
		file_volumes_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PeerPing); i {
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
		file_volumes_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*BlockCommandRequest); i {
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
		file_volumes_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*BlockCommandResponse); i {
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
		file_volumes_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RPCMetadata); i {
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
		file_volumes_proto_msgTypes[8].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PeerRPC); i {
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
	file_volumes_proto_msgTypes[8].OneofWrappers = []interface{}{
		(*PeerRPC_RaftRPC)(nil),
		(*PeerRPC_PeerPing)(nil),
		(*PeerRPC_BlockCommand)(nil),
		(*PeerRPC_BlockCommandResponse)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_volumes_proto_rawDesc,
			NumEnums:      2,
			NumMessages:   11,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_volumes_proto_goTypes,
		DependencyIndexes: file_volumes_proto_depIdxs,
		EnumInfos:         file_volumes_proto_enumTypes,
		MessageInfos:      file_volumes_proto_msgTypes,
	}.Build()
	File_volumes_proto = out.File
	file_volumes_proto_rawDesc = nil
	file_volumes_proto_goTypes = nil
	file_volumes_proto_depIdxs = nil
}
