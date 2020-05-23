package vpctp

import (
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tcfw/vpc/pkg/l2/transport/tap/protocol"
)

func TestMarshal(t *testing.T) {
	v := &VPCTP{listenAddr: "::1"}
	p := &protocol.Packet{
		VNID:   5,
		Source: []byte(""),
		Frame:  []byte{0x1, 0x2},
	}

	b := v.toBytes(p)
	pfb := v.toPacket(b)

	assert.Equal(t, p.VNID, pfb.VNID)
	assert.Equal(t, p.Source, pfb.Source)
	assert.Equal(t, p.Frame, pfb.Frame)
}

func TestSendRecv(t *testing.T) {
	//Needs root
	v := &VPCTP{listenAddr: "::1"}
	p := &protocol.Packet{
		VNID:   5,
		Source: []byte("abcdef"),
		Frame:  []byte{0x1, 0x2},
	}

	received := 0
	handledSomething := false

	v.SetHandler(func(ps []*protocol.Packet) {
		received++
		handledSomething = true
	})

	err := v.Start()
	if assert.NoError(t, err) {
		defer v.Stop()

		v.SendOne(p, net.ParseIP("::1"))
	}

	for {
		if handledSomething {
			break
		}
		time.Sleep(1 * time.Millisecond)
	}

	assert.Equal(t, 1, received)
}

func BenchmarkToBytes(b *testing.B) {
	v := &VPCTP{listenAddr: "::1"}
	p := &protocol.Packet{
		VNID:   5,
		Source: []byte("abcdef"),
		Frame: []byte{
			0xCC, 0x03, 0x04, 0xDC, 0x00, 0x10, 0xFF, 0xAA, 0xFA, 0xAA, 0xFF, 0xAA, 0x81, 0x00,
			0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x7F, 0x00, 0x00, 0x01, 0x08, 0x08, 0x08, 0x08, 0x00, 0x08, 0x00, 0x00,
			0x00, 0x01, 0x00, 0x01, 0x30, 0x31, 0x32, 0x33, 0x34, 0x35, 0x36, 0x37, 0x38, 0x39,
			0x61, 0x62, 0x63, 0x64, 0x65, 0x66,
		},
	}

	b.Run("toBytes", func(b *testing.B) {
		b.SetBytes(int64(4 + 2 + len(p.Source) + len(p.Frame)))
		for i := 0; i < b.N; i++ {
			v.toBytes(p)
		}
	})
}

func BenchmarkSend(b *testing.B) {
	//Needs root
	v := &VPCTP{listenAddr: "::1"}
	p := &protocol.Packet{
		VNID:   5,
		Source: []byte("abcdef"),
		Frame: []byte{
			0xCC, 0x03, 0x04, 0xDC, 0x00, 0x10, 0xFF, 0xAA, 0xFA, 0xAA, 0xFF, 0xAA, 0x81, 0x00,
			0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x7F, 0x00, 0x00, 0x01, 0x08, 0x08, 0x08, 0x08, 0x00, 0x08, 0x00, 0x00,
			0x00, 0x01, 0x00, 0x01, 0x30, 0x31, 0x32, 0x33, 0x34, 0x35, 0x36, 0x37, 0x38, 0x39,
			0x61, 0x62, 0x63, 0x64, 0x65, 0x66,
		},
	}

	received := 0

	v.SetHandler(func(ps []*protocol.Packet) {
		received++
	})

	if err := v.Start(); err != nil {
		panic(err)
	}

	defer v.Stop()
	dst := net.ParseIP("::1")

	b.Run("Send", func(b *testing.B) {
		b.SetBytes(int64(4 + 2 + len(p.Source) + len(p.Frame)))
		for i := 0; i < b.N; i++ {
			v.SendOne(p, dst)
		}
	})
}
