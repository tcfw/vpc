package transport

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFDBAddEntry(t *testing.T) {
	fdb := NewFDB()

	rdst := net.ParseIP("1.1.1.1")
	mac, _ := net.ParseMAC("10:20:30:40:50:60")
	var vnid uint32 = 1

	fdb.AddEntry(vnid, mac, rdst)

	ip := fdb.LookupMac(vnid, mac)

	assert.Equal(t, rdst, ip)
}

func TestFDBListBroadcast(t *testing.T) {
	fdb := NewFDB()

	rdst := net.ParseIP("1.1.1.1")
	mac, _ := net.ParseMAC("00:00:00:00:00:00")
	var vnid uint32 = 1

	fdb.AddEntry(vnid, mac, rdst)

	ips := fdb.ListBroadcast(vnid)

	assert.Contains(t, ips, rdst)
}

func TestFDBLookupMacWithBroadcast(t *testing.T) {
	fdb := NewFDB()

	rdst1 := net.ParseIP("1.1.1.1")
	rdst2 := net.ParseIP("2.2.2.2")
	mac, _ := net.ParseMAC("10:20:30:40:50:60")
	bcmac, _ := net.ParseMAC("00:00:00:00:00:00")
	var vnid uint32 = 1

	fdb.AddEntry(vnid, bcmac, rdst2)
	fdb.AddEntry(vnid, mac, rdst1)

	ip := fdb.LookupMac(vnid, mac)

	assert.Equal(t, rdst1, ip)
}
