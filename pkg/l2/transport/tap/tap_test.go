package tap

import (
	"github.com/tcfw/vpc/pkg/l2/transport/tap/protocol"
	"github.com/vishvananda/netlink"
)

type testTap struct {
	vnid  uint32
	tx    protocol.HandlerFunc
	packs []*protocol.Packet
}

func (t *testTap) Stop() error {
	return nil
}

func (t *testTap) Start() {
	select {}
}

func (t *testTap) Write(b []byte) (int, error) {
	return len(b), nil
}

func (t *testTap) Receive(b []*protocol.Packet) {
	t.tx(b)
}

func (t *testTap) Delete() error {
	return nil
}

func (t *testTap) IFace() netlink.Link {
	return &netlink.Dummy{LinkAttrs: netlink.LinkAttrs{Name: "foo"}}
}

func (t *testTap) SetHandler(h protocol.HandlerFunc) {
	t.tx = h
}
