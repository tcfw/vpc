package hyper

import (
	"encoding/xml"
	"net"
)

/*
Based from the libvirt XML format
See https://libvirt.org/formatdomain.html for more info or potental additions
*/

//DomainDesc main domain description
type DomainDesc struct {
	XMLName  xml.Name       `xml:"domain"`
	Type     string         `xml:"type,attr"`
	Name     string         `xml:"name"`
	UUID     string         `xml:"uuid,omtiempty"`
	Metadata DomainMetadata `xml:"metadata"`

	OnPoweroff string `xml:"on_poweroff"`
	OnReboot   string `xml:"on_reboot"`
	OnCrash    string `xml:"on_crash"`

	OS       DomainOS       `xml:"os"`
	Features DomainFeatures `xml:"features"`

	Memory        DomainMemory       `xml:"memory"`
	CurrentMemory *DomainMemory      `xml:"currentMemory,omitempty"`
	VCPU          DomainVCPU         `xml:"vcpu"`
	CPU           DomainCPU          `xml:"cpu"`
	Clock         DomainClock        `xml:"clock"`
	Blkiotune     *DomainBlockIOTune `xml:"blkiotune,omitempty"`

	PowerManagement DomainPM `xml:"pm"`

	Devices DomainDevices `xml:"devices"`
}

//DomainMetadata main domain metadata
type DomainMetadata struct {
	Text string `xml:",innerxml"`
}

//DomainFeatures Hypervisor Features
type DomainFeatures struct {
	Acpi   *DomainFeatureAcpi  `xml:"acpi"`
	Apic   *DomainFeatureApic  `xml:"apic"`
	Vmport DomainFeatureVmport `xml:"vmport"`
}

//DomainFeatureVmport VMWare IO port support
type DomainFeatureVmport struct {
	State string `xml:"state,attr"`
}

//DomainBlockIOTune Block IO Tune
type DomainBlockIOTune struct {
	Weight int `xml:"weight,omitempty"`

	Devices []DomainBlkIOTuneDevice `xml:"device"`
}

//DomainBlkIOTuneDevice Device level Block IO tuning
type DomainBlkIOTuneDevice struct {
	Path   string `xml:"path"`
	Weight int    `xml:"weight,omitempty"`

	ReadBytesSec  uint32 `xml:"read_bytes_sec,omitempty"`
	WriteBytesSec uint32 `xml:"write_bytes_sec,omitempty"`
	ReadIOPsSec   uint32 `xml:"read_iops_sec,omitempty"`
	WriteIOPsSec  uint32 `xml:"write_iops_sec,omitempty"`
}

//DomainFeatureAcpi Feature ACPI
type DomainFeatureAcpi struct{}

//DomainFeatureApic Feature APIC
type DomainFeatureApic struct{}

//DomainMemory Memory
type DomainMemory struct {
	Amount uint64 `xml:",chardata"`
	Unit   string `xml:"unit,attr"`
}

//DomainVCPU VCPU
type DomainVCPU struct {
	Count     uint16 `xml:",chardata"`
	Placement string `xml:"placement,attr,omitempty"`
	CPUSet    string `xml:"cpuset,attr,omitempty"`
}

//DomainCPU CPU
type DomainCPU struct {
	Mode  string `xml:"mode,attr"`
	Check string `xml:"check,attr"`
}

//DomainOS OS
type DomainOS struct {
	Type DomainOSType `xml:"type"`
	Boot DomainOSBoot `xml:"boot"`
}

//DomainOSType OS Type
type DomainOSType struct {
	Type    string `xml:",chardata"`
	Arch    string `xml:"arch,attr"`
	Machine string `xml:"machine,attr"`
}

//DomainOSBoot OS Boot Device
type DomainOSBoot struct {
	Dev string `xml:"dev,attr"`
}

//DomainClock Clock
type DomainClock struct {
	Offset string             `xml:"offset,attr"`
	Timer  []DomainClockTimer `xml:"timer"`
}

//DomainClockTimer Clock Timer
type DomainClockTimer struct {
	Name       string `xml:"name,attr"`
	Tickpolicy string `xml:"tickpolicy,attr,omitempty"`
	Present    string `xml:"present,attr,omitempty"`
}

//DomainPM Power Management
type DomainPM struct {
	SuspendToMem  DomainPMSuspendType `xml:"suspend-to-mem"`
	SuspendToDisk DomainPMSuspendType `xml:"suspend-to-disk"`
}

//DomainPMSuspendType PM Suspend Type
type DomainPMSuspendType struct {
	Enabled string `xml:"enabled,attr"`
}

//DomainDevices Devices
type DomainDevices struct {
	Emulator string `xml:"emulator"`

	Disks       []DomainDisk        `xml:"disk"`
	Console     []DomainConsole     `xml:"console"`
	Serial      []DomainSerial      `xml:"serial"`
	Controllers []DomainController  `xml:"controller"`
	Channel     []DomainChannel     `xml:"channel"`
	Input       []DomainInput       `xml:"input,omitempty"`
	Graphics    []DomainGraphics    `xml:"graphics,omitempty"`
	Video       []DomainVideo       `xml:"video,omitempty"`
	Rng         []DomainRNG         `xml:"rng"`
	MemBalloon  DomainMemoryBalloon `xml:"memballoon"`
}

//DomainDisk Disk
type DomainDisk struct {
	XMLName  xml.Name            `xml:"disk"`
	Type     string              `xml:"type,attr"`
	Device   string              `xml:"device,attr"`
	ReadOnly *DomainDiskReadOnly `xml:"readonly"`
	IOTune   *DomainDiskIOTune   `xml:"iotune"`

	Driver       DomainDiskDriver        `xml:"driver"`
	Source       DomainDiskSource        `xml:"source"`
	BackingStore *DomainDiskBackingStore `xml:"backingStore,omitempty"`
	Target       DomainDiskTarget        `xml:"target"`
	Address      DomainDeviceAddr        `xml:"address"`
}

//DomainDiskIOTune Disk IO Tune
type DomainDiskIOTune struct {
	TotalBytesSec    uint64 `xml:"total_bytes_sec,omitempty"`
	ReadBytesSec     uint64 `xml:"read_bytes_sec,omitempty"`
	WriteBytesSec    uint64 `xml:"write_bytes_sec,omitempty"`
	TotalIopsSec     uint64 `xml:"total_iops_sec,omitempty"`
	ReadIopsSec      uint64 `xml:"read_iops_sec,omitempty"`
	WriteIopsSec     uint64 `xml:"write_iops_sec,omitempty"`
	TotalBytesSecMax uint64 `xml:"total_bytes_sec_max,omitempty"`
	ReadBytesSecMax  uint64 `xml:"read_bytes_sec_max,omitempty"`
	WriteBytesSecMax uint64 `xml:"write_bytes_sec_max,omitempty"`
	TotalIopsSecMax  uint64 `xml:"total_iops_sec_max,omitempty"`
	ReadIopsSecMax   uint64 `xml:"read_iops_sec_max,omitempty"`
	WriteIopsSecMax  uint64 `xml:"write_iops_sec_max,omitempty"`
	SizeIopsSec      uint64 `xml:"size_iops_sec,omitempty"`

	GroupName uint64 `xml:"group_name,omitempty"`

	TotalBytesSecMaxLength uint64 `xml:"total_bytes_sec_max_length,omitempty"`
	ReadBytesSecMaxLength  uint64 `xml:"read_bytes_sec_max_length,omitempty"`
	WriteBytesSecMaxLength uint64 `xml:"write_bytes_sec_max_length,omitempty"`
	TotalIopsSecMaxLength  uint64 `xml:"total_iops_sec_max_length,omitempty"`
	ReadIopsSecMaxLength   uint64 `xml:"read_iops_sec_max_length,omitempty"`
	WriteIopsSecMaxLength  uint64 `xml:"write_iops_sec_max_length,omitempty"`
}

//DomainDiskReadOnly Disk Read Only
type DomainDiskReadOnly struct {
	XMLName xml.Name `xml:"readonly"`
}

//DomainDiskTarget Disk Target
type DomainDiskTarget struct {
	Dev string `xml:"dev,attr"`
	Bus string `xml:"bus,attr"`
}

//DomainDiskDriver Disk Driver
type DomainDiskDriver struct {
	Name string `xml:"name,attr"`
	Type string `xml:"type,attr"`
}

//DomainDiskSource Disk Source
type DomainDiskSource struct {
	Name          string `xml:"name,attr,omitempty"`
	File          string `xml:"file,attr,omitempty"`
	StartupPolicy string `xml:"startupPolicy,attr,omitempty"`
	Protocol      string `xml:"protocol,attr,omitempty"`

	Host *DomainDiskSourceHost `xml:"host,omitempty"`
}

//DomainDiskSourceHost Disk Source Host
type DomainDiskSourceHost struct {
	Text string `xml:",chardata"`
	Name string `xml:"name,attr,omitempty"`
	Port string `xml:"port,attr,omitempty"`
}

//DomainDiskBackingStore Disk Backing Store
type DomainDiskBackingStore struct {
}

//DomainDeviceAddr Device Addr
type DomainDeviceAddr struct {
	Type     string `xml:"type,attr,omitempty"`
	Domain   string `xml:"domain,attr,omitempty"`
	Bus      string `xml:"bus,attr,omitempty"`
	Slot     string `xml:"slot,attr,omitempty"`
	Function string `xml:"function,attr,omitempty"`
}

//DomainConsole Console
type DomainConsole struct {
	Type string `xml:"type,attr"`

	Target DomainConsoleTarget `xml:"target"`
}

//DomainConsoleTarget Console Target
type DomainConsoleTarget struct {
	Type string `xml:"type,attr"`
	Port int    `xml:"port,attr"`
}

//DomainSerial Serial
type DomainSerial struct {
	Type string `xml:"type,attr"`

	Target DomainSerialTarget `xml:"target"`
}

//DomainSerialTarget Serial Target
type DomainSerialTarget struct {
	Type string `xml:"type,attr"`
	Port int    `xml:"port,attr"`

	Model DomainSerialTargetModel `xml:"model"`
}

//DomainSerialTargetModel Serial Target Model
type DomainSerialTargetModel struct {
	Name string `xml:"name,attr"`
}

//DomainInterface Interface
type DomainInterface struct {
	Type string `xml:"type,attr"`

	Mac DomainInterfaceMac `xml:"mac"`

	Address DomainDeviceAddr `xml:"address"`
}

//DomainInterfaceMac Interface Mac
type DomainInterfaceMac struct {
	Address net.HardwareAddr `xml:"address,attr"`
}

//DomainController Controller
type DomainController struct {
	Type      string `xml:"type,attr"`
	Index     string `xml:"index,attr"`
	ModelAttr string `xml:"model,attr,omitempty"`

	Master *DomainControllerMaster `xml:"master,omitempty"`
	Model  *DomainControllerModel  `xml:"model,omitempty"`
	Target *DomainControllerTarget `xml:"target,omitempty"`

	Address *DomainDeviceAddr `xml:"address"`
}

//DomainControllerMaster Controller Master
type DomainControllerMaster struct {
	Startport string `xml:"startport,attr,omitempty"`
}

//DomainControllerModel Controller Model
type DomainControllerModel struct {
	Name string `xml:"name,attr,omitempty"`
}

//DomainControllerTarget Controller Target
type DomainControllerTarget struct {
	Chassis string `xml:"chassis,attr,omitempty"`
	Port    string `xml:"port,attr,omitempty"`
}

//DomainMemoryBalloon Memory Balloon
type DomainMemoryBalloon struct {
	Model   string           `xml:"model,attr"`
	Address DomainDeviceAddr `xml:"address"`
}

//DomainRNG RNG (Random Number Generator)
type DomainRNG struct {
	Model   string           `xml:"model,attr"`
	Backend DomainRNGBackend `xml:"backend"`

	Address DomainDeviceAddr `xml:"address"`
}

//DomainRNGBackend RNG Backend
type DomainRNGBackend struct {
	Text  string `xml:",chardata"`
	Model string `xml:"model,attr"`
}

//DomainVideo Video
type DomainVideo struct {
	Model   DomainVideoModel `xml:"model"`
	Address DomainDeviceAddr `xml:"address"`
}

//DomainVideoModel Video Model
type DomainVideoModel struct {
	Text    string `xml:",chardata"`
	Type    string `xml:"type,attr"`
	RAM     string `xml:"ram,attr"`
	VRAM    string `xml:"vram,attr"`
	Vgamem  string `xml:"vgamem,attr"`
	Heads   string `xml:"heads,attr"`
	Primary string `xml:"primary,attr"`
}

//DomainGraphics Graphics
type DomainGraphics struct {
	Type     string `xml:"type,attr"`
	Autoport string `xml:"autoport,attr"`

	Listen DomainGraphicsListen `xml:"listen"`
	Image  DomainGraphicsImage  `xml:"image"`
}

//DomainGraphicsListen Graphics Listen
type DomainGraphicsListen struct {
	Text string `xml:",chardata"`
	Type string `xml:"type,attr"`
}

//DomainGraphicsImage Graphics Image
type DomainGraphicsImage struct {
	Text        string `xml:",chardata"`
	Compression string `xml:"compression,attr"`
}

//DomainChannel Channel
type DomainChannel struct {
	Type string `xml:"type,attr"`

	Target DomainChannelTarget `xml:"target"`

	Address DomainDeviceAddr `xml:"address"`
}

//DomainChannelTarget Channel Target
type DomainChannelTarget struct {
	Type string `xml:"type,attr"`
	Name string `xml:"name,attr"`
}

//DomainInput Input
type DomainInput struct {
	Type    string            `xml:"type,attr"`
	Bus     string            `xml:"bus,attr"`
	Address *DomainDeviceAddr `xml:"address"`
}

/*******************
===== Back ups =====
*******************/

//DomainBackup Backup
type DomainBackup struct {
	XMLName xml.Name `xml:"domainbackup"`
	Mode    string   `xml:"mode,attr,omitempty"`

	Disks       []DomainBackupDisk `xml:"disks>disk"`
	Incremental uint64             `xml:"incremental,omitempty"`
}

//DomainBackupDisk Backup Disk
type DomainBackupDisk struct {
	XMLName xml.Name `xml:"disk"`
	Name    string   `xml:"name,attr"`
	Type    string   `xml:"type,attr"`

	Backup       string `xml:"backup,attr,omitempty"`
	ExportName   string `xml:"exportname,attr,omitempty"`
	ExportBitmap string `xml:"exportbitmap,attr,omitempty"`

	Driver DomainBackupDiskDriver `xml:"driver"`
	Target DomainBackupDiskTarget `xml:"target"`
}

//DomainBackupDiskDriver Backup Disk Driver
type DomainBackupDiskDriver struct {
	Name string `xml:"name,attr,omitempty"`
	Type string `xml:"type,attr"`
}

//DomainBackupDiskTarget Backup Disk Target
type DomainBackupDiskTarget struct {
	File     string `xml:"file,attr,omitempty"`
	Protocol string `xml:"protocol,attr,omitempty"`

	Host *DomainBackupDiskTargetHost `xml:"host,omitempty"`
}

//DomainBackupDiskTargetHost Backup Disk Target Host
type DomainBackupDiskTargetHost struct {
	Text string `xml:",chardata"`
	Name string `xml:"name,attr,omitempty"`
	Port string `xml:"port,attr,omitempty"`
}

//DomainCheckpoint Checkpoint
type DomainCheckpoint struct {
	XMLName     xml.Name `xml:"domaincheckpoint"`
	Description string   `xml:"description"`

	Disks []DomainCheckpointDisk `xml:"disks>disk"`
}

//DomainCheckpointDisk Checkpoint Disk
type DomainCheckpointDisk struct {
	XMLName    xml.Name `xml:"disk"`
	Name       string   `xml:"name,attr"`
	Checkpoint string   `xml:"checkpoint,attr,omitempty"`
	Bitmap     string   `xml:"bitmap,attr,omitempty"`
}
