package tap

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
	entries map[string][]*FDBEntry
	mu      sync.RWMutex

	broadcastMac net.HardwareAddr
}

//NewFDB inits a new FDB table and start GC timer
func NewFDB() *FDB {
	bc, _ := net.ParseMAC("00:00:00:00:00:00")

	tbl := &FDB{
		entries:      map[string][]*FDBEntry{},
		broadcastMac: bc,
	}
	go tbl.gc()

	return tbl
}

//LookupMac finds dst VTEP for a given MAC
func (tbl *FDB) LookupMac(vnid uint32, mac net.HardwareAddr) net.IP {
	var rdst net.IP

	tbl.mu.RLock()
	defer tbl.mu.RUnlock()

	hkey := hmackey(vnid, mac)

	if entries, ok := tbl.entries[hkey]; ok {
		for k, val := range entries {
			if bytes.Compare(val.mac, mac) == 0 && val.vnid == vnid && bytes.Compare(val.mac, tbl.broadcastMac) != 0 {
				//Update cache TS
				val.updated = time.Now()
				tbl.entries[hkey][k] = val

				rdst = val.rdst

				break
			}
		}
	}

	return rdst
}

//ListBroadcast finds all broadcast dst VTEPs
func (tbl *FDB) ListBroadcast(vnid uint32) []net.IP {
	tbl.mu.RLock()
	defer tbl.mu.RUnlock()

	dsts := []net.IP{}

	for _, entries := range tbl.entries {
		for _, val := range entries {
			if bytes.Compare(val.mac, tbl.broadcastMac) == 0 && val.vnid == vnid {
				dsts = append(dsts, val.rdst)
			}
		}
	}

	return dsts
}

//AddEntry adds a forwarding entry to the table
func (tbl *FDB) AddEntry(vnid uint32, mac net.HardwareAddr, rdst net.IP) {
	loc := -1

	tbl.mu.RLock()

	hkey := hmackey(vnid, mac)

	if _, ok := tbl.entries[hkey]; !ok {
		tbl.entries[hkey] = make([]*FDBEntry, 0, 10)
	}

	for k, entry := range tbl.entries[hkey] {
		if entry.vnid == vnid && bytes.Compare(entry.mac, mac) == 0 && entry.rdst.Equal(rdst) {
			loc = k
			break
		}
	}

	tbl.mu.RUnlock()

	if loc == -1 {
		// log.Printf("added FDB rec: %s %s", mac, rdst)
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
		tbl.entries[hkey] = append(tbl.entries[hkey], entry)
	} else {
		tbl.entries[hkey][loc] = entry
	}
}

func (tbl *FDB) gc() {
	for {
		time.Sleep(2 * time.Minute)
		tbl.gcOnce()
	}
}

func (tbl *FDB) gcOnce() {
	tbl.mu.Lock()

	for hkey, entries := range tbl.entries {
		for k, entry := range entries {
			//Delete entries older than 1 minute
			if entry.updated.Unix() < time.Now().Add(-3*time.Minute).Unix() {
				log.Printf("FDB GC: %s %s", entry.mac, entry.rdst)
				tbl.delEntry(hkey, k)
			}
		}
		if len(entries) == 0 {
			delete(tbl.entries, hkey)
		}
	}

	tbl.mu.Unlock()
}

func (tbl *FDB) delEntry(hkey string, i int) {
	tbl.entries[hkey][len(tbl.entries[hkey])-1], tbl.entries[hkey][i] = tbl.entries[hkey][i], tbl.entries[hkey][len(tbl.entries[hkey])-1]
	tbl.entries[hkey] = tbl.entries[hkey][:len(tbl.entries[hkey])-1]
}

func hmackey(vnid uint32, mac net.HardwareAddr) string {
	return string(vnid) + "/" + string(mac)
}
