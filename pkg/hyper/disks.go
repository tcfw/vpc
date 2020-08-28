package hyper

import (
	"encoding/xml"
	"fmt"

	"github.com/esiqveland/gogfapi/gfapi"

	libvirt "libvirt.org/libvirt-go"
)

//AvailableDisks lists available disks
func AvailableDisks() ([]string, error) {
	files := []string{}

	vol := &gfapi.Volume{}
	if err := vol.Init("disks", "localhost"); err != nil {
		return nil, fmt.Errorf("failed to connect: %s", err)
	}

	if err := vol.Mount(); err != nil {
		return nil, fmt.Errorf("failed to mount: %s", err)
	}
	defer vol.Unmount()

	f, err := vol.Open("/")
	if err != nil {
		return nil, err
	}

	volFiles, err := f.Readdir(0)
	if err != nil {
		return nil, err
	}

	for _, f := range volFiles {
		name := f.Name()
		if name == "." || name == ".." {
			continue
		}

		files = append(files, name)
	}

	return files, nil
}

//Snapshot attempts to create a snapshot of a running VM
func Snapshot(d *libvirt.Domain) error {
	desc := &DomainBackup{
		Disks: []DomainBackupDisk{
			{
				Name: "vda",
				Type: "file",
				Target: DomainBackupDiskTarget{
					File: "/home/tom/Desktop/backup.backup",
				},
				Driver: DomainBackupDiskDriver{
					Type: "raw",
				},
			},
		},
	}

	checkpoint := &DomainCheckpoint{
		Description: "test",
		Disks: []DomainCheckpointDisk{
			{
				Name: "vda",
			},
		},
	}

	descBytes, _ := xml.Marshal(desc)
	checkpointBytes, _ := xml.Marshal(checkpoint)

	fmt.Printf("BU: %s\nCP: %s", string(descBytes), string(checkpointBytes))

	return d.BackupBegin(string(descBytes), string(checkpointBytes), libvirt.DOMAIN_BACKUP_BEGIN_REUSE_EXTERNAL)
}
