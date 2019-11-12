package l3

import (
	"fmt"
	"os"
	"path"
	"runtime"
	"syscall"

	"github.com/vishvananda/netns"
)

func createNetNS(name string) (netns.NsHandle, error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	origns, _ := netns.Get()
	defer origns.Close()

	routerNetNs, err := netns.New()
	if err != nil {
		return 0, err
	}

	if err := setNSName(name); err != nil {
		fmt.Printf("Failed to add a name to NS: %s\n", err)
	}

	netns.Set(origns)

	return routerNetNs, nil
}

func setNSName(name string) error {
	p := path.Join("/var/run/netns/", name)
	f, err := os.OpenFile(p, os.O_CREATE|os.O_EXCL, 0444)
	if err != nil {
		return err
	}
	f.Close()
	nspath := fmt.Sprintf("/proc/%d/task/%d/ns/net", os.Getpid(), syscall.Gettid())
	if err := syscall.Mount(nspath, p, "bind", syscall.MS_BIND, ""); err != nil {
		return err
	}
	return nil
}

func unbindNSName(name string) error {
	p := path.Join("/var/run/netns/", name)
	if err := syscall.Unmount(p, 0); err != nil {
		return err
	}

	return os.Remove(p)
}

func execInNetNs(ns netns.NsHandle, fn func() error) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	origns, _ := netns.Get()
	defer origns.Close()

	netns.Set(ns)

	err := fn()

	netns.Set(origns)

	return err
}
