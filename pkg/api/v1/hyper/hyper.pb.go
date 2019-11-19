// Code generated by protoc-gen-go. DO NOT EDIT.
// source: hyper.proto

package hyper

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
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

type VM_PowerState int32

const (
	VM_NONE     VM_PowerState = 0
	VM_RUNNING  VM_PowerState = 1
	VM_SHUTDOWN VM_PowerState = 4
	VM_SHUTOFF  VM_PowerState = 5
	VM_CRASHED  VM_PowerState = 6
)

var VM_PowerState_name = map[int32]string{
	0: "NONE",
	1: "RUNNING",
	4: "SHUTDOWN",
	5: "SHUTOFF",
	6: "CRASHED",
}

var VM_PowerState_value = map[string]int32{
	"NONE":     0,
	"RUNNING":  1,
	"SHUTDOWN": 4,
	"SHUTOFF":  5,
	"CRASHED":  6,
}

func (x VM_PowerState) String() string {
	return proto.EnumName(VM_PowerState_name, int32(x))
}

func (VM_PowerState) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_c3cc9b77647888f3, []int{0, 0}
}

type StorageDevice_Driver int32

const (
	StorageDevice_LOCAL StorageDevice_Driver = 0
	StorageDevice_RDB   StorageDevice_Driver = 1
)

var StorageDevice_Driver_name = map[int32]string{
	0: "LOCAL",
	1: "RDB",
}

var StorageDevice_Driver_value = map[string]int32{
	"LOCAL": 0,
	"RDB":   1,
}

func (x StorageDevice_Driver) String() string {
	return proto.EnumName(StorageDevice_Driver_name, int32(x))
}

func (StorageDevice_Driver) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_c3cc9b77647888f3, []int{2, 0}
}

type VM struct {
	Id                   string            `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	TemplateId           string            `protobuf:"bytes,2,opt,name=template_id,json=templateId,proto3" json:"template_id,omitempty"`
	Metadata             map[string][]byte `protobuf:"bytes,3,rep,name=metadata,proto3" json:"metadata,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	PlacementId          string            `protobuf:"bytes,4,opt,name=placement_id,json=placementId,proto3" json:"placement_id,omitempty"`
	SubnetId             string            `protobuf:"bytes,5,opt,name=subnet_id,json=subnetId,proto3" json:"subnet_id,omitempty"`
	Storage              []*StorageDevice  `protobuf:"bytes,6,rep,name=storage,proto3" json:"storage,omitempty"`
	FwRuleSets           []string          `protobuf:"bytes,7,rep,name=fw_rule_sets,json=fwRuleSets,proto3" json:"fw_rule_sets,omitempty"`
	Nics                 []string          `protobuf:"bytes,8,rep,name=nics,proto3" json:"nics,omitempty"`
	PowerState           VM_PowerState     `protobuf:"varint,9,opt,name=power_state,json=powerState,proto3,enum=vpc.hyper.VM_PowerState" json:"power_state,omitempty"`
	PowerStateLastUpdate string            `protobuf:"bytes,10,opt,name=power_state_last_update,json=powerStateLastUpdate,proto3" json:"power_state_last_update,omitempty"`
	XXX_NoUnkeyedLiteral struct{}          `json:"-"`
	XXX_unrecognized     []byte            `json:"-"`
	XXX_sizecache        int32             `json:"-"`
}

func (m *VM) Reset()         { *m = VM{} }
func (m *VM) String() string { return proto.CompactTextString(m) }
func (*VM) ProtoMessage()    {}
func (*VM) Descriptor() ([]byte, []int) {
	return fileDescriptor_c3cc9b77647888f3, []int{0}
}

func (m *VM) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_VM.Unmarshal(m, b)
}
func (m *VM) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_VM.Marshal(b, m, deterministic)
}
func (m *VM) XXX_Merge(src proto.Message) {
	xxx_messageInfo_VM.Merge(m, src)
}
func (m *VM) XXX_Size() int {
	return xxx_messageInfo_VM.Size(m)
}
func (m *VM) XXX_DiscardUnknown() {
	xxx_messageInfo_VM.DiscardUnknown(m)
}

var xxx_messageInfo_VM proto.InternalMessageInfo

func (m *VM) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

func (m *VM) GetTemplateId() string {
	if m != nil {
		return m.TemplateId
	}
	return ""
}

func (m *VM) GetMetadata() map[string][]byte {
	if m != nil {
		return m.Metadata
	}
	return nil
}

func (m *VM) GetPlacementId() string {
	if m != nil {
		return m.PlacementId
	}
	return ""
}

func (m *VM) GetSubnetId() string {
	if m != nil {
		return m.SubnetId
	}
	return ""
}

func (m *VM) GetStorage() []*StorageDevice {
	if m != nil {
		return m.Storage
	}
	return nil
}

func (m *VM) GetFwRuleSets() []string {
	if m != nil {
		return m.FwRuleSets
	}
	return nil
}

func (m *VM) GetNics() []string {
	if m != nil {
		return m.Nics
	}
	return nil
}

func (m *VM) GetPowerState() VM_PowerState {
	if m != nil {
		return m.PowerState
	}
	return VM_NONE
}

func (m *VM) GetPowerStateLastUpdate() string {
	if m != nil {
		return m.PowerStateLastUpdate
	}
	return ""
}

type VMTemplate struct {
	Id                   string   `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Vcpu                 int32    `protobuf:"varint,2,opt,name=vcpu,proto3" json:"vcpu,omitempty"`
	Ram                  int64    `protobuf:"varint,3,opt,name=ram,proto3" json:"ram,omitempty"`
	LibvirtXml           string   `protobuf:"bytes,4,opt,name=libvirt_xml,json=libvirtXml,proto3" json:"libvirt_xml,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *VMTemplate) Reset()         { *m = VMTemplate{} }
func (m *VMTemplate) String() string { return proto.CompactTextString(m) }
func (*VMTemplate) ProtoMessage()    {}
func (*VMTemplate) Descriptor() ([]byte, []int) {
	return fileDescriptor_c3cc9b77647888f3, []int{1}
}

func (m *VMTemplate) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_VMTemplate.Unmarshal(m, b)
}
func (m *VMTemplate) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_VMTemplate.Marshal(b, m, deterministic)
}
func (m *VMTemplate) XXX_Merge(src proto.Message) {
	xxx_messageInfo_VMTemplate.Merge(m, src)
}
func (m *VMTemplate) XXX_Size() int {
	return xxx_messageInfo_VMTemplate.Size(m)
}
func (m *VMTemplate) XXX_DiscardUnknown() {
	xxx_messageInfo_VMTemplate.DiscardUnknown(m)
}

var xxx_messageInfo_VMTemplate proto.InternalMessageInfo

func (m *VMTemplate) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

func (m *VMTemplate) GetVcpu() int32 {
	if m != nil {
		return m.Vcpu
	}
	return 0
}

func (m *VMTemplate) GetRam() int64 {
	if m != nil {
		return m.Ram
	}
	return 0
}

func (m *VMTemplate) GetLibvirtXml() string {
	if m != nil {
		return m.LibvirtXml
	}
	return ""
}

type StorageDevice struct {
	Id                   string               `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Device               string               `protobuf:"bytes,2,opt,name=device,proto3" json:"device,omitempty"`
	Driver               StorageDevice_Driver `protobuf:"varint,3,opt,name=driver,proto3,enum=vpc.hyper.StorageDevice_Driver" json:"driver,omitempty"`
	Size                 int64                `protobuf:"varint,4,opt,name=size,proto3" json:"size,omitempty"`
	XXX_NoUnkeyedLiteral struct{}             `json:"-"`
	XXX_unrecognized     []byte               `json:"-"`
	XXX_sizecache        int32                `json:"-"`
}

func (m *StorageDevice) Reset()         { *m = StorageDevice{} }
func (m *StorageDevice) String() string { return proto.CompactTextString(m) }
func (*StorageDevice) ProtoMessage()    {}
func (*StorageDevice) Descriptor() ([]byte, []int) {
	return fileDescriptor_c3cc9b77647888f3, []int{2}
}

func (m *StorageDevice) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_StorageDevice.Unmarshal(m, b)
}
func (m *StorageDevice) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_StorageDevice.Marshal(b, m, deterministic)
}
func (m *StorageDevice) XXX_Merge(src proto.Message) {
	xxx_messageInfo_StorageDevice.Merge(m, src)
}
func (m *StorageDevice) XXX_Size() int {
	return xxx_messageInfo_StorageDevice.Size(m)
}
func (m *StorageDevice) XXX_DiscardUnknown() {
	xxx_messageInfo_StorageDevice.DiscardUnknown(m)
}

var xxx_messageInfo_StorageDevice proto.InternalMessageInfo

func (m *StorageDevice) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

func (m *StorageDevice) GetDevice() string {
	if m != nil {
		return m.Device
	}
	return ""
}

func (m *StorageDevice) GetDriver() StorageDevice_Driver {
	if m != nil {
		return m.Driver
	}
	return StorageDevice_LOCAL
}

func (m *StorageDevice) GetSize() int64 {
	if m != nil {
		return m.Size
	}
	return 0
}

func init() {
	proto.RegisterEnum("vpc.hyper.VM_PowerState", VM_PowerState_name, VM_PowerState_value)
	proto.RegisterEnum("vpc.hyper.StorageDevice_Driver", StorageDevice_Driver_name, StorageDevice_Driver_value)
	proto.RegisterType((*VM)(nil), "vpc.hyper.VM")
	proto.RegisterMapType((map[string][]byte)(nil), "vpc.hyper.VM.MetadataEntry")
	proto.RegisterType((*VMTemplate)(nil), "vpc.hyper.VMTemplate")
	proto.RegisterType((*StorageDevice)(nil), "vpc.hyper.StorageDevice")
}

func init() { proto.RegisterFile("hyper.proto", fileDescriptor_c3cc9b77647888f3) }

var fileDescriptor_c3cc9b77647888f3 = []byte{
	// 519 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x74, 0x93, 0xcd, 0x6e, 0xda, 0x40,
	0x10, 0xc7, 0x63, 0xfc, 0x01, 0x8c, 0x09, 0xb2, 0x46, 0x51, 0x6b, 0x35, 0x95, 0xe2, 0x72, 0xe2,
	0xe4, 0x03, 0x55, 0x95, 0x7e, 0x9c, 0x92, 0x40, 0x1a, 0x54, 0x30, 0xd5, 0x3a, 0xd0, 0xaa, 0x17,
	0x6b, 0xb1, 0x37, 0xad, 0x55, 0x1b, 0x2c, 0x7b, 0x6d, 0x4a, 0x9f, 0xa3, 0x4f, 0xd0, 0x27, 0xad,
	0x76, 0x71, 0x48, 0x50, 0xd4, 0xdb, 0x7f, 0xe6, 0xf7, 0x67, 0x66, 0x96, 0x19, 0x83, 0xf9, 0x63,
	0x9b, 0xb1, 0xdc, 0xcd, 0xf2, 0x35, 0x5f, 0x63, 0xbb, 0xca, 0x42, 0x57, 0x26, 0x7a, 0x7f, 0x34,
	0x68, 0x2c, 0xa6, 0xd8, 0x85, 0x46, 0x1c, 0xd9, 0x8a, 0xa3, 0xf4, 0xdb, 0xa4, 0x11, 0x47, 0x78,
	0x06, 0x26, 0x67, 0x69, 0x96, 0x50, 0xce, 0x82, 0x38, 0xb2, 0x1b, 0x12, 0xc0, 0x7d, 0x6a, 0x1c,
	0xe1, 0x39, 0xb4, 0x52, 0xc6, 0x69, 0x44, 0x39, 0xb5, 0x55, 0x47, 0xed, 0x9b, 0x83, 0x53, 0x77,
	0x5f, 0xd5, 0x5d, 0x4c, 0xdd, 0x69, 0x4d, 0x47, 0x2b, 0x9e, 0x6f, 0xc9, 0xde, 0x8c, 0xaf, 0xa0,
	0x93, 0x25, 0x34, 0x64, 0x29, 0x5b, 0x71, 0x51, 0x5a, 0x93, 0xa5, 0xcd, 0x7d, 0x6e, 0x1c, 0xe1,
	0x29, 0xb4, 0x8b, 0x72, 0xb9, 0x62, 0x92, 0xeb, 0x92, 0xb7, 0x76, 0x89, 0x71, 0x84, 0x03, 0x68,
	0x16, 0x7c, 0x9d, 0xd3, 0xef, 0xcc, 0x36, 0x64, 0x5f, 0xfb, 0x51, 0x5f, 0x7f, 0x47, 0x86, 0xac,
	0x8a, 0x43, 0x46, 0xee, 0x8d, 0xe8, 0x40, 0xe7, 0x6e, 0x13, 0xe4, 0x65, 0xc2, 0x82, 0x82, 0xf1,
	0xc2, 0x6e, 0x3a, 0xaa, 0x78, 0xce, 0xdd, 0x86, 0x94, 0x09, 0xf3, 0x19, 0x2f, 0x10, 0x41, 0x5b,
	0xc5, 0x61, 0x61, 0xb7, 0x24, 0x91, 0x1a, 0xdf, 0x81, 0x99, 0xad, 0x37, 0x2c, 0x0f, 0x0a, 0x4e,
	0x39, 0xb3, 0xdb, 0x8e, 0xd2, 0xef, 0x1e, 0x74, 0x5b, 0x4c, 0xdd, 0xcf, 0xc2, 0xe0, 0x0b, 0x4e,
	0x20, 0xdb, 0x6b, 0x7c, 0x03, 0xcf, 0x1f, 0xfd, 0x34, 0x48, 0x68, 0xc1, 0x83, 0x32, 0x8b, 0x44,
	0x19, 0x90, 0xef, 0x39, 0x79, 0x30, 0x4f, 0x68, 0xc1, 0xe7, 0x92, 0xbd, 0xf8, 0x00, 0xc7, 0x07,
	0x7f, 0x1b, 0x5a, 0xa0, 0xfe, 0x64, 0xdb, 0x7a, 0x2f, 0x42, 0xe2, 0x09, 0xe8, 0x15, 0x4d, 0x4a,
	0x26, 0x57, 0xd2, 0x21, 0xbb, 0xe0, 0x7d, 0xe3, 0xad, 0xd2, 0xfb, 0x04, 0xf0, 0x30, 0x0d, 0xb6,
	0x40, 0xf3, 0x66, 0xde, 0xc8, 0x3a, 0x42, 0x13, 0x9a, 0x64, 0xee, 0x79, 0x63, 0xef, 0xa3, 0xa5,
	0x60, 0x07, 0x5a, 0xfe, 0xcd, 0xfc, 0x76, 0x38, 0xfb, 0xe2, 0x59, 0x9a, 0x40, 0x22, 0x9a, 0x5d,
	0x5f, 0x5b, 0xba, 0x08, 0xae, 0xc8, 0x85, 0x7f, 0x33, 0x1a, 0x5a, 0x46, 0x2f, 0x04, 0x58, 0x4c,
	0x6f, 0xeb, 0x75, 0x3f, 0xb9, 0x0e, 0x04, 0xad, 0x0a, 0xb3, 0x52, 0xce, 0xa0, 0x13, 0xa9, 0xc5,
	0xa8, 0x39, 0x4d, 0x6d, 0xd5, 0x51, 0xfa, 0x2a, 0x11, 0x52, 0xdc, 0x50, 0x12, 0x2f, 0xab, 0x38,
	0xe7, 0xc1, 0xaf, 0x34, 0xa9, 0x17, 0x0d, 0x75, 0xea, 0x6b, 0x9a, 0xf4, 0xfe, 0x2a, 0x70, 0x7c,
	0xb0, 0xb1, 0x27, 0x8d, 0x9e, 0x81, 0x11, 0x49, 0x52, 0x5f, 0x60, 0x1d, 0xe1, 0x39, 0x18, 0x51,
	0x1e, 0x57, 0x2c, 0x97, 0xfd, 0xba, 0x83, 0xb3, 0xff, 0xdd, 0x80, 0x3b, 0x94, 0x36, 0x52, 0xdb,
	0xc5, 0xe4, 0x45, 0xfc, 0x9b, 0xc9, 0x61, 0x54, 0x22, 0x75, 0xef, 0x25, 0x18, 0x3b, 0x17, 0xb6,
	0x41, 0x9f, 0xcc, 0xae, 0x2e, 0x26, 0xd6, 0x11, 0x36, 0x41, 0x25, 0xc3, 0x4b, 0x4b, 0xb9, 0x6c,
	0x7e, 0xd3, 0x65, 0xdd, 0xa5, 0x21, 0xbf, 0x9d, 0xd7, 0xff, 0x02, 0x00, 0x00, 0xff, 0xff, 0xb9,
	0xd7, 0x2d, 0x06, 0x4a, 0x03, 0x00, 0x00,
}
