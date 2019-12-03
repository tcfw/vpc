package transport

import (
	"fmt"
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
	entries map[string]*FDBEntry
	mu      sync.RWMutex
}

//NewFDB inits a new FDB table and start GC timer
func NewFDB() *FDB {
	tbl := &FDB{
		entries: map[string]*FDBEntry{},
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
		if val.mac.String() == dst.String() && val.vnid == vnid && val.mac.String() != "00:00:00:00:00:00" {
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
	dsts := []net.IP{}

	tbl.mu.RLock()
	defer tbl.mu.RUnlock()

	for _, val := range tbl.entries {
		if val.mac.String() == "00:00:00:00:00:00" && val.vnid == vnid {
			dsts = append(dsts, val.rdst)
		}
	}

	return dsts
}

//AddEntry adds a forwarding entry to the table
func (tbl *FDB) AddEntry(vnid uint32, mac net.HardwareAddr, rdst net.IP) {
	k := fmt.Sprintf("%s-%s", mac.String(), rdst.String())

	_, ok := tbl.entries[k]
	if !ok {
		log.Printf("added FDB rec: %s %s", mac, rdst)
	} else {
		log.Printf("updating FDB rec: %s %s", mac, rdst)
	}

	tbl.mu.Lock()
	defer tbl.mu.Unlock()

	tbl.entries[k] = &FDBEntry{
		vnid:    vnid,
		rdst:    rdst,
		mac:     mac,
		updated: time.Now(),
	}
}

func (tbl *FDB) gc() {
	for {
		time.Sleep(2 * time.Minute)
		for k, entry := range tbl.entries {
			//Delete entries older than 1 minute
			if entry.updated.Unix() < time.Now().Add(-1*time.Minute).Unix() {
				tbl.mu.Lock()

				log.Printf("FDB GC: %s %s", entry.mac, entry.rdst)
				delete(tbl.entries, k)

				tbl.mu.Unlock()
			}
		}
	}
}
