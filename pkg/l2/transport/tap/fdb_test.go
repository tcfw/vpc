package tap

import (
	"fmt"
	"math/rand"
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
	assert.Len(t, fdb.entries, 1)
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

func BenchmarkFDBLoopkup(b *testing.B) {
	fdb := NewFDB()

	mac, _ := net.ParseMAC("10:20:30:40:50:60")
	var vnid uint32 = 2

	n := 10000
	for i := 0; i < n; i++ {
		fdb.AddEntry(vnid, randomMac(), randomIP4())
	}
	fdb.AddEntry(vnid, mac, randomIP4())

	b.ResetTimer()

	b.Run("Lookup", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			fdb.LookupMac(vnid, mac)
		}
	})
}

func BenchmarkHMacKey(b *testing.B) {
	mac := net.HardwareAddr{0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
	vnid := uint32(1)

	b.Run("hkey", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			hmackey(vnid, mac)
		}
	})
}

func randomIP4() net.IP {
	return net.ParseIP(fmt.Sprintf("%d.%d.%d.%d", rand.Intn(255), rand.Intn(255), rand.Intn(255), rand.Intn(255)))
}

func randomMac() net.HardwareAddr {
	buf := make([]byte, 6)
	_, err := rand.Read(buf)
	if err != nil {
		fmt.Println("error:", err)
		return nil
	}
	// Set the local bit
	buf[0] |= 2

	return net.HardwareAddr(buf)
}
