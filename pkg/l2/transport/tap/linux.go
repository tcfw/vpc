package tap

import (
	"fmt"
	"io"
	"os"
	"strings"
	"unsafe"

	"golang.org/x/sys/unix"
)

type tuntapDev struct {
	io.ReadWriteCloser
	fd     int
	Device string
}

const (
	cIFFTUN        = 0x0001
	cIFFTAP        = 0x0002
	cIFFNOPI       = 0x1000
	cIFFMULTIQUEUE = 0x0100
)

type ifReq struct {
	Name  [16]byte
	Flags uint16
	pad   [8]byte
}

func ioctl(a1, a2, a3 uintptr) error {
	_, _, errno := unix.Syscall(unix.SYS_IOCTL, a1, a2, a3)
	if errno != 0 {
		return os.NewSyscallError("ioctl", errno)
	}
	return nil
}

func (c *tuntapDev) WriteRaw(b []byte) (int, error) {
	var nn int
	for {
		max := len(b)
		n, err := unix.Write(c.fd, b[nn:max])
		if n > 0 {
			nn += n
		}
		if nn == len(b) {
			return nn, err
		}

		if err != nil {
			return nn, err
		}

		if n == 0 {
			return nn, io.ErrUnexpectedEOF
		}
	}
}

func (v *Tap) openDev(deviceName string) (ifce *tuntapDev, err error) {
	fd, err := unix.Open("/dev/net/tun", os.O_RDWR, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to open linux tun dev: %s", err)
	}

	var req ifReq
	req.Flags = cIFFNOPI | cIFFTAP | cIFFMULTIQUEUE
	copy(req.Name[:], deviceName)

	if err = ioctl(uintptr(fd), unix.TUNSETIFF, uintptr(unsafe.Pointer(&req))); err != nil {
		return nil, err
	}

	name := strings.Trim(string(req.Name[:]), "\x00")

	file := os.NewFile(uintptr(fd), "/dev/net/tun")

	ifce = &tuntapDev{
		ReadWriteCloser: file,
		fd:              int(file.Fd()),
		Device:          name,
	}

	return
}
