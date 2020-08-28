package hyper

import (
	"encoding/json"
	"fmt"
	"os"

	diskfs "github.com/diskfs/go-diskfs"
	"github.com/diskfs/go-diskfs/disk"
	"github.com/diskfs/go-diskfs/filesystem"
	"github.com/diskfs/go-diskfs/filesystem/iso9660"
	hyperAPI "github.com/tcfw/vpc/pkg/api/v1/hyper"
)

const (
	cloudName = "vpc"
	diskStore = "/var/local/vpc/cidisks"
)

//InstanceMetadata basic instance info
type InstanceMetadata struct {
	InstanceID string `json:"instance-id"`
	CloudName  string `json:"cloud-name"`
}

func createCloudInitISO(vm *hyperAPI.VM) (*os.File, error) {
	instanceMD := &InstanceMetadata{
		InstanceID: vm.Id,
		CloudName:  cloudName,
	}

	diskImg, err := ciDataDisk(vm)
	if err != nil {
		return nil, err
	}

	mdbytes, err := json.Marshal(instanceMD)
	if err != nil {
		defer os.Remove(diskImg.Name())
		return diskImg, err
	}

	size := int64(len(vm.UserData) + len(mdbytes) + 1024)
	if size < 32768 {
		//min ISO file size
		size = 32768 + 2048 + 2048 + 2048
	}

	diskImg.Truncate(size)

	ciDisk, err := diskfs.OpenWithMode(diskImg.Name(), diskfs.ReadWriteExclusive)
	if err != nil {
		defer os.Remove(diskImg.Name())
		return diskImg, err
	}

	ciDisk.LogicalBlocksize = 2048
	fspec := disk.FilesystemSpec{Partition: 0, FSType: filesystem.TypeISO9660, VolumeLabel: "cidata"}

	fs, err := ciDisk.CreateFilesystem(fspec)
	if err != nil {
		defer os.Remove(diskImg.Name())
		return diskImg, err
	}

	//Metadata file
	fmd, err := fs.OpenFile("meta-data", os.O_CREATE|os.O_RDWR)
	if err != nil {
		defer os.Remove(diskImg.Name())
		return diskImg, fmt.Errorf("failed to open md: %s", err)
	}
	if _, err := fmd.Write(mdbytes); err != nil {
		defer os.Remove(diskImg.Name())
		return nil, fmt.Errorf("failed to write md: %s", err)
	}

	//User-data
	if vm.UserData != nil {
		fud, err := fs.OpenFile("user-data", os.O_CREATE|os.O_RDWR)
		if err != nil {
			defer os.Remove(diskImg.Name())
			return nil, fmt.Errorf("failed to write ud: %s", err)
		}

		if _, err := fud.Write(vm.UserData); err != nil {
			defer os.Remove(diskImg.Name())
			return nil, err
		}
	}

	//Finalise
	iso, ok := fs.(*iso9660.FileSystem)
	isows := iso.Workspace()
	if !ok {
		defer os.Remove(diskImg.Name())
		return nil, fmt.Errorf("not an iso9660 filesystem")
	}
	if err = iso.Finalize(iso9660.FinalizeOptions{}); err != nil {
		defer os.Remove(diskImg.Name())
		return nil, err
	}

	//Remove the temp workspace since diskfs does not
	os.RemoveAll(isows)

	return diskImg, diskImg.Sync()

}

func ciDataDisk(vm *hyperAPI.VM) (*os.File, error) {
	if _, err := os.Stat(diskStore); os.IsNotExist(err) {
		err := os.MkdirAll(diskStore, os.ModePerm)
		if err != nil {
			return nil, fmt.Errorf("failed to create local disk store: %s", err)
		}
	}

	return os.OpenFile(ciDiskName(vm), os.O_RDWR|os.O_CREATE, os.ModePerm)
}

func ciDiskName(vm *hyperAPI.VM) string {
	return fmt.Sprintf("%s/ci.%s.iso", diskStore, vm.Id)
}
