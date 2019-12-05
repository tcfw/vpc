package tap

import (
	"testing"

	"github.com/vishvananda/netlink"

	"github.com/stretchr/testify/assert"
)

func TestListen(t *testing.T) {
	lis, err := NewListener(4789)
	lis.SetMTU(1000)

	assert.NoError(t, err)

	go lis.Start()

	vtepname := "testVTEP"

	la := netlink.LinkAttrs{}
	la.Name = vtepname

	tuntap := &netlink.Tuntap{LinkAttrs: la, Mode: netlink.TUNTAP_MODE_TUN}
	err = netlink.LinkAdd(tuntap)
	assert.NoError(t, err)

	// err = lis.AddEP(5, vtepname)
	// assert.NoError(t, err)
}
