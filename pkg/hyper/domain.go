package hyper

import (
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"os"

	hyperAPI "github.com/tcfw/vpc/pkg/api/v1/hyper"
	"libvirt.org/libvirt-go"
)

//DefineVM takes in an API desc and creates a hypervisor accepted VM
func DefineVM(c *libvirt.Connect, vm *hyperAPI.VM) error {
	if vm.Id == "" {
		return fmt.Errorf("vm has no id")
	}

	if vm.HyperId != "" {
		d, err := c.LookupDomainByUUIDString(vm.HyperId)
		if err != nil {
			return fmt.Errorf("failed to check for existing: %s", err)
		}
		if d != nil {
			return fmt.Errorf("vm already defined")
		}
	}

	//construct cloud-init disk
	var cidata *os.File
	cidiskn := ciDiskName(vm)
	if _, err := os.Stat(cidiskn); os.IsExist(err) {
		cidata, err = os.OpenFile(cidiskn, os.O_EXCL|os.O_RDWR, os.ModePerm)
		if err != nil {
			return fmt.Errorf("failed to open cidata: %s", err)
		}
	} else {
		cidata, err = createCloudInitISO(vm)
		if err != nil {
			return fmt.Errorf("failed to create cidata: %s", err)
		}
	}

	//Closed for IO
	cidata.Close()

	desc, err := GetTemplate(vm)
	if err != nil {
		return err
	}

	desc.Devices.Disks = append(desc.Devices.Disks, DomainDisk{
		Type:     "file",
		Device:   "disk",
		ReadOnly: &DomainDiskReadOnly{},
		Target:   DomainDiskTarget{Bus: "virtio", Dev: "vdz"},
		Driver:   DomainDiskDriver{Name: "qemu", Type: "raw"},
		Source:   DomainDiskSource{File: cidata.Name()},
		Address: DomainDeviceAddr{ //TODO(tcfw) check if PCI slot is available
			Type:     "pci",
			Domain:   "0x0000",
			Bus:      "0x09",
			Slot:     "0x00",
			Function: "0x00",
		},
		IOTune: &DomainDiskIOTune{
			TotalIopsSec:          100,
			TotalIopsSecMax:       200,
			TotalIopsSecMaxLength: 10,
		},
	})

	uuid, err := createDomain(c, desc)
	if err != nil {
		return fmt.Errorf("failed to define vm: %s", err)
	}
	vm.HyperId = hex.EncodeToString(uuid)
	return nil

}

func createDomain(c *libvirt.Connect, desc interface{}) ([]byte, error) {
	switch desc.(type) {
	case *DomainDesc:
		descBytes, err := xml.Marshal(desc)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal to XML desc: %s", err)
		}
		return createXMLDomain(c, string(descBytes))
	case []byte:
		return createXMLDomain(c, string(desc.([]byte)))
	case string:
		return createXMLDomain(c, desc.(string))
	}

	return nil, fmt.Errorf("unknown desc format")
}

func createXMLDomain(c *libvirt.Connect, descXML string) ([]byte, error) {
	d, err := c.DomainDefineXML(descXML)
	if err != nil {
		return nil, fmt.Errorf("failed to define domain: %s", err)
	}

	uuid, err := d.GetUUID()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve uuid from defined domain: %s", err)
	}

	return uuid, nil
}
