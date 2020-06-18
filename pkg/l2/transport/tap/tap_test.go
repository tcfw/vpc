package tap

import (
	"bytes"
	"log"
	"testing"
	"time"

	"github.com/google/gopacket/pcap"
	"github.com/stretchr/testify/assert"
	"github.com/tcfw/vpc/pkg/l2/transport/tap/protocol"
	"github.com/tcfw/vpc/pkg/l2/xdp"
	"github.com/vishvananda/netlink"
)

type testTap struct {
	vnid  uint32
	tx    protocol.HandlerFunc
	packs []protocol.Packet
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

func (t *testTap) Receive(b []protocol.Packet) {
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

func TestXDPWrite(t *testing.T) {
	//requires root
	handler := &protocol.TestHandler{}
	fdb := NewFDB()

	lis := &Listener{
		taps: map[uint32]Nic{},
		mtu:  1500,
		FDB:  fdb,
		conn: handler,
	}

	br := &netlink.Bridge{
		LinkAttrs: netlink.LinkAttrs{
			Name: "b-1",
		},
	}

	netlink.LinkAdd(br)
	netlink.LinkSetUp(br)
	defer netlink.LinkDel(br)

	tap, err := NewTap(lis, 1, 1500, br)
	if assert.NoError(t, err) {
		defer tap.Delete()
		lis.taps[1] = tap

		packet, _ := constructPacket()

		n, err := tap.Write(packet)

		assert.NoError(t, err)
		assert.Equal(t, len(packet), n)

		s, _ := tap.xdp[0].Stats()
		assert.Equal(t, uint64(1), s.Completed)
	}
}

func TestXDPRead(t *testing.T) {
	handler := &protocol.TestHandler{}
	fdb := NewFDB()

	lis := &Listener{
		taps: map[uint32]Nic{},
		mtu:  1500,
		FDB:  fdb,
		conn: handler,
	}

	br := &netlink.Bridge{
		LinkAttrs: netlink.LinkAttrs{
			Name: "b-1",
		},
	}

	netlink.LinkAdd(br)
	netlink.LinkSetUp(br)
	defer netlink.LinkDel(br)

	tap, err := NewTap(lis, 1, 1500, br)
	if assert.NoError(t, err) {
		defer tap.Delete()
		lis.taps[1] = tap

		recv := 0
		recvData := [][]byte{}

		tap.SetHandler(func(ps []protocol.Packet) {
			t.Logf("RX N: %d", len(ps))
			for _, p := range ps {
				if bytes.Compare(p.Frame[:6], []byte{0xe6, 0x15, 0x9c, 0x38, 0xf6, 0xe7}) != 0 {
					// t.Logf("PKDATA: [% X]\n", p.Frame)
					continue
				}
				t.Logf("Y")
				recv++
				recvData = append(recvData, p.Frame)
			}
		})

		go tap.Start()

		frame, err := constructVPCPacket()
		if err != nil {
			log.Println(err)
			return
		}

		//Inject Packet into Bridge
		handle, err := pcap.OpenLive("b-1", 1024, true, 10*time.Second)
		if err != nil {
			log.Fatal(err)
			return
		}
		defer handle.Close()

		for i := 0; i < 50; i++ {
			handle.WritePacketData(frame)
			t.Log("-")
		}

		for {
			if recv >= 50 {
				break
			}
			time.Sleep(100 * time.Millisecond)
		}

		log.Printf("C: %d\n", recv)
		assert.Equal(t, 50, recv)
		log.Printf("D: % X", recvData[0])
	}
}

func BenchmarkTapWrite(b *testing.B) {
	//requires root
	handler := &protocol.TestHandler{}
	fdb := NewFDB()

	lis := &Listener{
		taps: map[uint32]Nic{},
		mtu:  1500,
		FDB:  fdb,
		conn: handler,
	}

	br := &netlink.Bridge{
		LinkAttrs: netlink.LinkAttrs{
			Name: "b-1",
		},
	}

	netlink.LinkAdd(br)
	netlink.LinkSetUp(br)
	defer netlink.LinkDel(br)

	tap, err := NewTap(lis, 1, 1500, br)
	if err != nil {
		panic(err)
	}
	defer tap.Delete()
	lis.taps[1] = tap

	packet, err := constructPacket()
	if err != nil {
		panic(err)
	}

	complete := false

	go func() {
		for {
			time.Sleep(5 * time.Second)
			s, _ := tap.xdp[0].Stats()
			log.Printf("STATS: %+v\n", s)
			if complete {
				return
			}
		}
	}()

	var t time.Duration
	b.Run("write", func(b *testing.B) {
		b.SetBytes(int64(len(packet)))
		st := time.Now()
		for i := 0; i < b.N; i++ {
			if _, err := tap.Write(packet); err != nil {
				panic(err)
			}
		}
		t = time.Since(st)
	})

	complete = true

	s, _ := tap.xdp[0].Stats()
	log.Printf("Effective PPS: %d/s", s.Completed/uint64(t.Seconds()))
}

func BenchmarkTapBatchWrite(b *testing.B) {
	//requires root
	handler := &protocol.TestHandler{}
	fdb := NewFDB()

	lis := &Listener{
		taps: map[uint32]Nic{},
		mtu:  1500,
		FDB:  fdb,
		conn: handler,
	}

	br := &netlink.Bridge{
		LinkAttrs: netlink.LinkAttrs{
			Name: "b-1",
		},
	}

	netlink.LinkAdd(br)
	netlink.LinkSetUp(br)
	defer netlink.LinkDel(br)

	tap, err := NewTap(lis, 1, 1500, br)
	if err != nil {
		panic(err)
	}
	defer tap.Delete()
	lis.taps[1] = tap

	packet, err := constructPacket()
	if err != nil {
		panic(err)
	}

	packets := []xdp.BatchDesc{
		{Data: packet, Len: len(packet)},
		{Data: packet, Len: len(packet)},
		{Data: packet, Len: len(packet)},
		{Data: packet, Len: len(packet)},
		{Data: packet, Len: len(packet)},
		{Data: packet, Len: len(packet)},
		{Data: packet, Len: len(packet)},
		{Data: packet, Len: len(packet)},
	}

	complete := false

	go func() {
		for {
			time.Sleep(5 * time.Second)
			s, _ := tap.xdp[0].Stats()
			log.Printf("STATS: %+v\n", s)
			if complete {
				return
			}
		}
	}()

	var t time.Duration
	b.Run("write", func(b *testing.B) {
		b.SetBytes(int64(len(packet) * len(packets)))
		st := time.Now()
		for i := 0; i < b.N; i++ {
			if _, err := tap.WriteBatch(packets); err != nil {
				panic(err)
			}
		}
		t = time.Since(st)
	})

	complete = true

	s, _ := tap.xdp[0].Stats()
	log.Printf("Effective PPS: %d/s", s.Completed/uint64(t.Seconds()))
}
