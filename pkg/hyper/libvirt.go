package hyper

import (
	"fmt"

	libvirt "libvirt.org/libvirt-go"
)

//NewLocalLibVirtConn connects to the local libvirt daemon
func NewLocalLibVirtConn() (*libvirt.Connect, error) {
	l, err := libvirt.NewConnect("qemu:///system")
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %v", err)
	}

	return l, nil
}

func testLibVirtConn() (*libvirt.Connect, error) {
	return libvirt.NewConnect("test:///default")
}
