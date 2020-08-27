package sbs

import (
	"testing"

	badger "github.com/dgraph-io/badger/v2"
	"github.com/hashicorp/raft"
	"github.com/stretchr/testify/assert"
)

func TestLogAdd(t *testing.T) {
	db, err := badger.Open(badger.DefaultOptions("").WithInMemory(true))
	if err != nil {
		t.Fatal(err)
	}

	logStore := &LogStore{
		volumeID: "vol-test",
		db:       db,
	}

	//Check index bounds
	i, err := logStore.FirstIndex()
	if err != nil {
		t.Fatal(err)
	}
	assert.Zero(t, i)

	i, err = logStore.LastIndex()
	if err != nil {
		t.Fatal(err)
	}
	assert.Zero(t, i)

	err = logStore.StoreLog(&raft.Log{
		Term:  1,
		Index: 1,
		Type:  raft.LogCommand,
		Data:  []byte{1, 2, 3, 4, 5},
	})

	if assert.NoError(t, err) {
		//Check index bounds after store
		i, err = logStore.FirstIndex()
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, uint64(1), i)

		i, err = logStore.LastIndex()
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, uint64(1), i)

		err = logStore.StoreLog(&raft.Log{
			Term:  1,
			Index: 2,
			Type:  raft.LogCommand,
			Data:  []byte{1, 2, 3, 4, 5},
		})

		if assert.NoError(t, err) {
			i, err = logStore.FirstIndex()
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, uint64(1), i)

			i, err = logStore.LastIndex()
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, uint64(2), i)
		}
	}
}
