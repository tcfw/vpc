package sbs

import (
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	volumesAPI "github.com/tcfw/vpc/pkg/api/v1/volumes"
)

func TestRaftTransport(t *testing.T) {
	ids, servers, err := makeTestCluster(3, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanUpServerData(servers)

	obsCh, observer := makePeerObserver()

	for i, srv := range servers {
		if err := srv.AddVolume(&volumesAPI.Volume{
			Id:   "vol-test",
			Size: 10,
		}, ids); err != nil {
			t.Fatal(err)
		}
		srv.Volume("vol-test").Raft.RegisterObserver(observer)
		log.Printf("added volume to %d", i)
	}

	waitForNPeers(2, obsCh)
}

func TestRaftApply(t *testing.T) {
	ids, servers, err := makeTestCluster(3, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanUpServerData(servers)

	obsCh, observer := makePeerObserver()

	for i, srv := range servers {
		if err := srv.AddVolume(&volumesAPI.Volume{
			Id:   "vol-test",
			Size: 10,
		}, ids); err != nil {
			t.Fatal(err)
		}
		srv.Volume("vol-test").Raft.RegisterObserver(observer)
		log.Printf("added volume to %d", i)
	}

	waitForNPeers(2, obsCh)

	leader := getLeader(servers)

	cmd := &StoreCommand{
		Op:     OpWrite,
		Offset: 0,
		Length: BlockSize,
		Data:   []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0},
	}

	cmdBytes, _ := cmd.Encode()

	apply := leader.Volume("vol-test").Raft.Apply(cmdBytes, 5*time.Second)
	err = apply.Error()
	if assert.NoError(t, err) {
		err, isErr := apply.Response().(error)
		if isErr {
			assert.NoError(t, err)
		} else {
			assert.Nil(t, err)
		}
	}
}

func TestRaftSnapshot(t *testing.T) {
	ids, servers, err := makeTestCluster(3, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanUpServerData(servers)

	obsCh, observer := makePeerObserver()

	for i, srv := range servers {
		if err := srv.AddVolume(&volumesAPI.Volume{
			Id:   "vol-test",
			Size: 1,
		}, ids); err != nil {
			t.Fatal(err)
		}
		srv.Volume("vol-test").Raft.RegisterObserver(observer)
		log.Printf("added volume to %d", i)
	}

	waitForNPeers(2, obsCh)

	leader := getLeader(servers)

	cmd := &StoreCommand{
		Op:     OpWrite,
		Offset: 0,
		Length: BlockSize,
		Data:   []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0},
	}

	cmdBytes, _ := cmd.Encode()

	apply := leader.Volume("vol-test").Raft.Apply(cmdBytes, 5*time.Second)
	err = apply.Error()
	if assert.NoError(t, err) {
		future := leader.Volume("vol-test").Raft.Snapshot()
		if err := future.Error(); err != nil {
			t.Fatal(err)
		}
	}
}

func BenchmarkRaftApply(b *testing.B) {
	ids, servers, err := makeTestCluster(3, nil)
	if err != nil {
		b.Fatal(err)
	}
	defer cleanUpServerData(servers)

	obsCh, observer := makePeerObserver()

	for i, srv := range servers {
		if err := srv.AddVolume(&volumesAPI.Volume{
			Id:   "vol-test",
			Size: 10,
		}, ids); err != nil {
			b.Fatal(err)
		}
		srv.Volume("vol-test").Raft.RegisterObserver(observer)
		log.Printf("added volume to %d", i)
	}

	waitForNPeers(2, obsCh)
	leader := getLeader(servers)

	cmd := &StoreCommand{
		Op:     OpWrite,
		Offset: 0,
		Length: BlockSize,
		Data:   []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0},
	}

	cmdBytes, _ := cmd.Encode()

	b.Run("apply", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			leader.Volume("vol-test").Raft.Apply(cmdBytes, 5*time.Second)
		}
	})
}
