package ns

import (
	"fmt"
	"log"
	"os"
	"path"
	"runtime"
	"syscall"

	virshnetns "github.com/vishvananda/netns"
)

const (
	netnsDir = "/var/run/netns/"
)

//NetNS allows communicating with a network namespace
type NetNS struct {
	handle virshnetns.NsHandle
	name   string
}

//Fd provides the file decriptor of the namespace
func (n *NetNS) Fd() int {
	return int(n.handle)
}

//CreateNetNS creates a new network namespace and sets a name
func CreateNetNS(name string) (*NetNS, error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	origns, _ := virshnetns.Get()
	defer origns.Close()

	routerNetNs, err := virshnetns.New()
	if err != nil {
		return nil, err
	}

	ns := &NetNS{
		handle: routerNetNs,
	}

	if len(name) != 0 {
		if err := ns.SetNSName(name); err != nil {
			log.Printf("Failed to add a name to NS: %s\n", err)
			return ns, fmt.Errorf("failed to add name to ns: %s", err)
		}
	}

	virshnetns.Set(origns)

	return ns, nil
}

//SetNSName sets a name for the netns which can be accessible via 'ip netns' commands
func (n *NetNS) SetNSName(name string) error {
	//Attempt to make the alias dir if does not exists
	if _, err := os.Stat(netnsDir); os.IsNotExist(err) {
		os.MkdirAll(netnsDir, 0777)
	}

	p := path.Join(netnsDir, name)

	f, err := os.OpenFile(p, os.O_CREATE|os.O_EXCL, 0444)
	if err != nil {
		return fmt.Errorf("failed to create netns name: %s", err)
	}

	f.Close()

	nspath := fmt.Sprintf("/proc/%d/task/%d/ns/net", os.Getpid(), syscall.Gettid())

	if err := syscall.Mount(nspath, p, "bind", syscall.MS_BIND, ""); err != nil {
		return err
	}

	n.name = name

	return nil
}

//Unbind removes the mount created from setNSName
func (n *NetNS) Unbind() error {
	if len(n.name) == 0 {
		return fmt.Errorf("ns has no name to unbind")
	}

	p := path.Join(netnsDir, n.name)
	if err := syscall.Unmount(p, 0); err != nil {
		return err
	}

	return os.Remove(p)
}

//Exec switches to the specified netns and executes the closure
//WARNING: CANNOT EXECUTE WITH GOROUTINES _INSIDE_ THE CLOSURE
func (n *NetNS) Exec(fn func() error) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	origns, _ := virshnetns.Get()
	defer origns.Close()

	//Switch to our namespace
	virshnetns.Set(n.handle)

	err := fn()

	//Switch back to original namespace
	virshnetns.Set(origns)

	return err
}

//Close unbinds and closes the netns
func (n *NetNS) Close() error {
	n.Unbind()
	return n.handle.Close()
}
