package vxlan

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPacketFromBytes(t *testing.T) {
	rawBytes := []byte{
		0x8,           //flags
		0x0, 0x0, 0x0, //group policy
		0x0, 0x0, 0x5, //vnid
		0x0,           //resv 2
		0x1, 0x2, 0x3, //inner packet
	}

	p, err := FromBytes(bytes.NewBuffer(rawBytes))
	assert.NoError(t, err, "failed to parse packet")

	assert.Equal(t, uint32(5), p.VNID, "VNID not same")
	assert.Equal(t, uint8(0), p.resv, "RES not same")
	assert.Equal(t, rawBytes[8:], []byte(p.InnerFrame))
}

func TestInvalidHeader(t *testing.T) {
	rawBytes := []byte{
		0x0,           //flags
		0x0, 0x0, 0x0, //group policy
		0x0, 0x0, 0x5, //vnid
		0x0,           //resv 2
		0x1, 0x2, 0x3, //inner packet
	}

	_, err := FromBytes(bytes.NewBuffer(rawBytes))
	assert.Error(t, err)
}

func TestPacketToBytes(t *testing.T) {
	p := &Packet{
		Flags:      0x08,
		VNID:       5,
		InnerFrame: []byte{0x1, 0x2, 0x3},
	}

	rawBytes := p.Bytes()

	assert.Equal(t, []byte{
		0x8,           //flags
		0x0, 0x0, 0x0, //group policy
		0x0, 0x0, 0x5, //vnid
		0x0,           //resv 2
		0x1, 0x2, 0x3, //inner packet
	}, rawBytes)
}

func TestPacketNewPacketToBytes(t *testing.T) {
	p := NewPacket(5, []byte{
		0x1, 0x2, 0x3,
	})
	raw := p.Bytes()
	assert.Equal(t, []byte{
		0x8,           //flags
		0x0, 0x0, 0x0, //group policy
		0x0, 0x0, 0x5, //vnid
		0x0,           //resv 2
		0x1, 0x2, 0x3, //inner packet
	}, raw)
}

func BenchmarkToByte(b *testing.B) {
	innerFrame := []byte{
		0xff, 0x2, 0x3, 0x1, 0x2, 0x3, 0x1, 0x2, 0x3, 0x1, 0x2, 0x3, 0x1, 0x2, 0x3, 0xf4,
		0xff, 0x2, 0x3, 0x1, 0x2, 0x3, 0x1, 0x2, 0x3, 0x1, 0x2, 0x3, 0x1, 0x2, 0x3, 0xf4,
	}
	vni := uint32(5)

	b.Run("convert", func(b *testing.B) {
		c := int64(0)
		for i := 0; i < b.N; i++ {
			p := NewPacket(vni, innerFrame)
			bytes := p.Bytes()
			c = c + int64(len(bytes))
		}
		b.SetBytes(c)
	})
}
