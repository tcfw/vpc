package sbs

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"

	badger "github.com/dgraph-io/badger/v2"
	"github.com/hashicorp/raft"
)

//LogStore provides a way for raft to store logs
type LogStore struct {
	volumeID string
	db       *badger.DB
}

//NewLogStore creates a new log store used in raft
func NewLogStore(id string, db *badger.DB) *LogStore {
	return &LogStore{
		volumeID: id,
		db:       db,
	}
}

const (
	keyFirstIndex = "fidx"
	keyLastIndex  = "lidx"
)

//FirstIndex returns the first index written. 0 for no entries.
func (rls *LogStore) FirstIndex() (uint64, error) {
	var v uint64

	err := rls.db.View(func(tx *badger.Txn) error {
		vRaw, err := tx.Get(rls.key(keyFirstIndex))
		if err == badger.ErrKeyNotFound {
			return nil
		}
		if err != nil {
			return err
		}
		buf, err := vRaw.ValueCopy(nil)
		if err != nil {
			return err
		}
		v = binary.LittleEndian.Uint64(buf)
		return nil
	})

	return v, err
}

//LastIndex returns the last index written. 0 for no entries.
func (rls *LogStore) LastIndex() (uint64, error) {
	var v uint64

	err := rls.db.View(func(tx *badger.Txn) error {
		vRaw, err := tx.Get(rls.key(keyLastIndex))
		if err == badger.ErrKeyNotFound {
			return nil
		}
		if err != nil {
			return err
		}
		buf, err := vRaw.ValueCopy(nil)
		if err != nil {
			return err
		}
		v = binary.LittleEndian.Uint64(buf)
		return nil
	})

	return v, err
}

//GetLog gets a log entry at a given index.
func (rls *LogStore) GetLog(index uint64, l *raft.Log) error {
	var lRaw []byte

	err := rls.db.View(func(tx *badger.Txn) error {
		vRaw, err := tx.Get(rls.indexKey(index))
		if err != nil {
			return err
		}
		buf, err := vRaw.ValueCopy(nil)
		if err != nil {
			return err
		}
		lRaw = buf
		return nil
	})
	if err != nil {
		return err
	}

	return rls.decode(lRaw, l)
}

//StoreLog stores a log entry.
func (rls *LogStore) StoreLog(log *raft.Log) error {
	return rls.StoreLogs([]*raft.Log{log})
}

//StoreLogs stores multiple log entries.
func (rls *LogStore) StoreLogs(logs []*raft.Log) error {
	err := rls.db.Update(func(tx *badger.Txn) error {
		for _, l := range logs {
			fIndex := rls.key(keyFirstIndex)
			lIndex := rls.key(keyLastIndex)
			//Set first index if it doesn't exist
			_, err := tx.Get(fIndex)
			if err == badger.ErrKeyNotFound {
				fInb := make([]byte, 8)
				binary.LittleEndian.PutUint64(fInb, l.Index)
				err = tx.Set(fIndex, fInb)
			}
			if err != nil {
				return err
			}

			//Update last index
			lInb := make([]byte, 8)
			binary.LittleEndian.PutUint64(lInb, l.Index)
			if err := tx.Set(lIndex, lInb); err != nil {
				return err
			}

			lRaw, err := rls.encode(l)
			if err != nil {
				return err
			}
			err = tx.Set(rls.indexKey(l.Index), lRaw)
			if err != nil {
				return err
			}
		}

		return nil
	})
	return err
}

//DeleteRange deletes a range of log entries. The range is inclusive.
func (rls *LogStore) DeleteRange(min, max uint64) error {
	return rls.db.Update(func(tx *badger.Txn) error {
		for i := min; i <= max; i++ {
			tx.Delete(rls.indexKey(i))
		}

		//Update first index assuming last index is further ahead
		fInb := make([]byte, 8)
		binary.LittleEndian.PutUint64(fInb, max+1)
		if err := tx.Set(rls.key(keyFirstIndex), fInb); err != nil {
			return err
		}

		return nil
	})
}

func (rls *LogStore) indexKey(index uint64) []byte {
	return []byte(fmt.Sprintf("%s:i:%d", rls.volumeID, index))
}

func (rls *LogStore) key(k string) []byte {
	return []byte(fmt.Sprintf("%s:%s", rls.volumeID, k))
}

func (rls *LogStore) keyRaw(k []byte) []byte {
	return rls.key(string(k))
}

func (rls *LogStore) encode(log *raft.Log) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(log)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (rls *LogStore) decode(b []byte, l *raft.Log) error {
	buf := bytes.NewBuffer(b)
	enc := gob.NewDecoder(buf)

	err := enc.Decode(l)
	return err
}
