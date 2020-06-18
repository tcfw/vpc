package vxlan

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/songgao/packets/ethernet"
)

//Packet represents the VxLAN headers and inner packet
type Packet struct {
	Flags       uint8
	GroupPolicy uint32 //wire: uint24
	VNID        uint32 //wire: uint24
	resv        uint8
	InnerFrame  ethernet.Frame

	_buf bytes.Buffer
}

//NewPacket encaps a inner packet with VXlan headers
func NewPacket(vnid uint32, innerFrame []byte) Packet {
	return Packet{
		Flags:       0x08,
		GroupPolicy: 0,
		VNID:        vnid,
		resv:        0,
		InnerFrame:  innerFrame,
	}
}

//FromBytes converts a byte array to a vxlan packet
func FromBytes(bytesReader io.Reader) (*Packet, error) {
	p := &Packet{}
	//Flags
	flags := uint8(0)
	binary.Read(bytesReader, binary.BigEndian, &flags)
	p.Flags = flags

	if p.Flags != 0x8 {
		return nil, fmt.Errorf("invalid magic integer")
	}

	//Group Policy
	groupPolicy := make([]byte, 3)
	binary.Read(bytesReader, binary.BigEndian, &groupPolicy)
	groupPolicy = append(groupPolicy, []byte{0x0}...)
	p.GroupPolicy = binary.BigEndian.Uint32(groupPolicy)

	//VNID expanding 3 bytes to 4
	vnid := make([]byte, 3)
	binary.Read(bytesReader, binary.BigEndian, &vnid)
	vnid = append([]byte{0x0}, vnid...)
	p.VNID = binary.BigEndian.Uint32(vnid)

	//Resv
	resv := uint8(0)
	binary.Read(bytesReader, binary.BigEndian, &resv)
	p.resv = resv

	//Inner frame
	var innerFrame bytes.Buffer
	io.Copy(&innerFrame, bytesReader)
	p.InnerFrame = innerFrame.Bytes()

	return p, nil
}

//Bytes converts the packet to raw bytes
func (p *Packet) Bytes() []byte {
	p._buf.Reset()
	p.WriteTo(&p._buf)

	return p._buf.Bytes()
}

//WriteTo writes a pakcet to a give writer
func (p *Packet) WriteTo(w *bytes.Buffer) (int64, error) {
	w.WriteByte(p.Flags)

	w.WriteByte(byte(p.GroupPolicy >> 16))
	w.WriteByte(byte(p.GroupPolicy >> 8))
	w.WriteByte(byte(p.GroupPolicy))

	w.WriteByte(byte(p.VNID >> 16))
	w.WriteByte(byte(p.VNID >> 8))
	w.WriteByte(byte(p.VNID))

	w.WriteByte(0) //resv

	n, _ := w.Write(p.InnerFrame)

	return int64(n + 8), nil
}
