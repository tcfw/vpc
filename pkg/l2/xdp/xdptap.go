package xdp

import (
	"fmt"
	"sync/atomic"
	"time"

	"github.com/vishvananda/netlink"
	"golang.org/x/sys/unix"
)

//Tap provides communication between taps in bridges and other endpoints, via XDP/eBFP redirects
//implements io.ReadWriteCloser
type Tap struct {
	xsk   *Socket
	iface netlink.Link

	recvWaiting atomic.Value
}

//BatchDesc holds descriptors and data for batch reading
type BatchDesc struct {
	Len  int
	Data []byte
}

//NewTap creates a new XDP attachment to the given iface queue
func NewTap(iface netlink.Link, queue int, prog ProgRef) (*Tap, error) {
	xdpt := &Tap{
		iface: iface,
	}

	xsk, err := NewSocket(iface.Attrs().Index, queue, prog)
	if err != nil {
		return nil, fmt.Errorf("xdp sock: %s", err)
	}

	xdpt.xsk = xsk

	xdpt.recvWaiting.Store(int(0))

	return xdpt, nil
}

func (xdpt *Tap) Write(p []byte) (int, error) {
	for {
		if xdpt.xsk.NumFreeTxSlots() == 0 {
			if err := xdpt.pollTx(); err != nil {
				return 0, err
			}
			continue
		}
		break
	}

	txSlots := xdpt.xsk.GetDescs(1)

	n := copy(xdpt.xsk.GetFrame(txSlots[0]), p)
	txSlots[0].Len = uint32(n)

	//TX
	if txn := xdpt.xsk.Transmit(txSlots); txn == 0 {
		return 0, fmt.Errorf("Failed to send")
	}

	//Wait
	if err := xdpt.pollTx(); err != nil {
		return 0, err
	}

	//Mark
	xdpt.xsk.Complete(xdpt.xsk.NumCompleted())

	return n, nil
}

//BatchWrite takes an array of byte arrays to be transmitted
func (xdpt *Tap) BatchWrite(ps []*BatchDesc) (int, error) {
	for {
		if xdpt.xsk.NumFreeTxSlots() <= len(ps) {
			if err := xdpt.pollTx(); err != nil {
				return 0, err
			}
			continue
		}
		break
	}

	var n int

	txSlots := xdpt.xsk.GetDescs(len(ps))
	for i := range txSlots {
		ni := copy(xdpt.xsk.GetFrame(txSlots[i]), ps[i].Data[ps[i].Len:])
		txSlots[i].Len = uint32(ps[i].Len)
		n += ni
	}

	if txn := xdpt.xsk.Transmit(txSlots); txn == 0 {
		return 0, fmt.Errorf("Failed to send")
	}

	if err := xdpt.pollTx(); err != nil {
		return 0, err
	}

	xdpt.xsk.Complete(xdpt.xsk.NumCompleted())

	return n, nil
}

func (xdpt *Tap) Read(p []byte) (int, error) {
	//prevent unessesary polling if we already know there's more to receive
	ready := xdpt.recvWaiting.Load().(int)
	if ready == 0 {
		for {
			xdpt.xsk.Fill(xdpt.xsk.GetDescs(xdpt.xsk.NumFreeFillSlots()))
			numrecv, err := xdpt.pollRx()
			if err != nil {
				return 0, err
			}
			//Check if poll was for recieving
			if numrecv != 0 {
				numrecv--
				xdpt.recvWaiting.Store(numrecv)
				ready = numrecv
				break
			}
			time.Sleep(1 * time.Microsecond)
		}
	} else {
		ready--
		xdpt.recvWaiting.Store(ready)
	}
	rxDescs := xdpt.xsk.Receive(1)
	n := copy(p, xdpt.xsk.GetFrame(rxDescs[0]))

	return n, nil
}

//BatchRead batch reads from the tap
func (xdpt *Tap) BatchRead(ps []*BatchDesc) (bytesRead int, nRead int, err error) {
	ready := xdpt.recvWaiting.Load().(int)

	if ready == 0 {
		for {
			xdpt.xsk.Fill(xdpt.xsk.GetDescs(xdpt.xsk.NumFreeFillSlots()))
			numrecv, err := xdpt.pollRx()
			if err != nil {
				return bytesRead, nRead, err
			}
			if numrecv != 0 {
				xdpt.recvWaiting.Store(numrecv)
				ready = numrecv
				break
			}
			time.Sleep(1 * time.Microsecond)
		}
	}

	toRead := ready
	if toRead > len(ps) {
		toRead = len(ps)
	}

	rxDescs := xdpt.xsk.Receive(ready)
	for i := range rxDescs {
		pktData := xdpt.xsk.GetFrame(rxDescs[i])
		n := copy(ps[i].Data[0:], pktData)
		ps[i].Len = n
		bytesRead += n
		nRead++
		ready--
	}

	xdpt.recvWaiting.Store(ready)

	return
}

//Close the underlying xdp socket
func (xdpt *Tap) Close() error {
	return xdpt.xsk.Close()
}

//Stats provides statistics for the interface (XDP)
func (xdpt *Tap) Stats() (Stats, error) {
	return xdpt.xsk.Stats()
}

//pollTx polls the fd only for TX events
func (xdpt *Tap) pollTx() error {
	var pfds [1]unix.PollFd
	pfds[0].Fd = int32(xdpt.xsk.FD())
	pfds[0].Events = unix.POLLOUT
	_, err := unix.Poll(pfds[:], -1)
	if err != nil {
		return err
	}
	return nil
}

//pollRx polls the fd only for RX events and provides number of received frames
func (xdpt *Tap) pollRx() (int, error) {
	var pfds [1]unix.PollFd
	pfds[0].Fd = int32(xdpt.xsk.FD())
	pfds[0].Events = unix.POLLIN
	_, err := unix.Poll(pfds[:], -1)
	if err != nil {
		return 0, err
	}
	return xdpt.xsk.NumReceived(), nil
}
