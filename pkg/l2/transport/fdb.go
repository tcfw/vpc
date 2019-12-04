package transport

import (
	"bytes"
	"log"
	"net"
	"sync"
	"time"
)

//FDBEntry forwarding DB entry
type FDBEntry struct {
	vnid    uint32
	mac     net.HardwareAddr
	rdst    net.IP
	updated time.Time
}

//FDB forwarding DB
type FDB struct {
	entries []*FDBEntry
	mu      sync.RWMutex

	broadcastMac net.HardwareAddr
}

//NewFDB inits a new FDB table and start GC timer
func NewFDB() *FDB {
	bc, _ := net.ParseMAC("00:00:00:00:00:00")

	tbl := &FDB{
		entries:      []*FDBEntry{},
		broadcastMac: bc,
	}
	go tbl.gc()

	return tbl
}

//LookupMac finds dst VTEP for a given MAC
func (tbl *FDB) LookupMac(vnid uint32, dst net.HardwareAddr) net.IP {
	var rdst net.IP

	tbl.mu.RLock()
	defer tbl.mu.RUnlock()

	for k, val := range tbl.entries {
		if bytes.Compare(val.mac, dst) == 0 && val.vnid == vnid && bytes.Compare(val.mac, tbl.broadcastMac) != 0 {
			//Update cache TS
			val.updated = time.Now()
			tbl.entries[k] = val

			rdst = val.rdst

			break
		}
	}

	return rdst
}

//ListBroadcast finds all broadcast dst VTEPs
func (tbl *FDB) ListBroadcast(vnid uint32) []net.IP {
	tbl.mu.RLock()
	defer tbl.mu.RUnlock()

	dsts := []net.IP{}

	for _, val := range tbl.entries {
		if bytes.Compare(val.mac, tbl.broadcastMac) == 0 && val.vnid == vnid {
			dsts = append(dsts, val.rdst)
		}
	}

	return dsts
}

//AddEntry adds a forwarding entry to the table
func (tbl *FDB) AddEntry(vnid uint32, mac net.HardwareAddr, rdst net.IP) {
	loc := -1

	tbl.mu.RLock()

	for k, entry := range tbl.entries {
		if entry.vnid == vnid && bytes.Compare(entry.mac, mac) == 0 && entry.rdst.Equal(rdst) {
			loc = k
			break
		}
	}

	tbl.mu.RUnlock()

	if loc == -1 {
		log.Printf("added FDB rec: %s %s", mac, rdst)
	}

	tbl.mu.Lock()
	defer tbl.mu.Unlock()

	entry := &FDBEntry{
		vnid:    vnid,
		rdst:    rdst,
		mac:     mac,
		updated: time.Now(),
	}

	if loc == -1 {
		tbl.entries = append(tbl.entries, entry)
	} else {
		tbl.entries[loc] = entry
	}
}

func (tbl *FDB) gc() {
	for {
		time.Sleep(2 * time.Minute)

		expired := []int{}

		for k, entry := range tbl.entries {
			//Delete entries older than 1 minute
			if entry.updated.Unix() < time.Now().Add(-3*time.Minute).Unix() {

				log.Printf("FDB GC: %s %s", entry.mac, entry.rdst)
				expired = append(expired, k)

			}
		}

		tbl.mu.Lock()

		for _, k := range expired {
			tbl.delEntry(k)
		}

		tbl.mu.Unlock()
	}
}

func (tbl *FDB) delEntry(i int) {
	tbl.entries[len(tbl.entries)-1], tbl.entries[i] = tbl.entries[i], tbl.entries[len(tbl.entries)-1]
	tbl.entries = tbl.entries[:len(tbl.entries)-1]
}
