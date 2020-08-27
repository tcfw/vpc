package sbs

import (
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	volumesAPI "github.com/tcfw/vpc/pkg/api/v1/volumes"
)

func TestPing(t *testing.T) {
	ids, servers, err := makeTestCluster(2, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanUpServerData(servers)

	resp, err := servers[1].sendReq(servers[1].peers[ids[0]], &volumesAPI.PeerPing{Ts: time.Now().Unix()})
	if err != nil {
		t.Fatal(err)
	}

	assert.NotEmpty(t, resp.(*volumesAPI.PeerPing).Ts)
}

func BenchmarkPing(b *testing.B) {
	ids, servers, err := makeTestCluster(2, nil)
	if err != nil {
		b.Fatal(err)
	}
	defer cleanUpServerData(servers)

	peer := servers[1].peers[ids[0]]
	msg := &volumesAPI.PeerPing{Ts: time.Now().Unix()}

	failures := 0

	b.Run("ping", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := servers[1].SendReqViaChannel(peer, "ping", msg)
			if err != nil {
				failures++
			}
		}
	})

	b.Logf("failures: %d", failures)
}

func TestBlockCommandWrite(t *testing.T) {
	ids, servers, err := makeTestCluster(2, nil)
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
		srv.volumes["vol-test"].Raft.RegisterObserver(observer)
		log.Printf("added volume to %d", i)
	}

	waitForNPeers(1, obsCh)
	leader := getLeader(servers)
	notLeader := getNotLeader(servers)

	//Send request from not leader to leader
	resp, err := notLeader.sendReq(notLeader.peers[leader.PeerID()], &volumesAPI.BlockCommandRequest{
		Volume: "vol-test",
	})
	if err != nil {
		t.Fatal(err)
	}

	assert.NotNil(t, resp)
	bcr, isBCR := resp.(*volumesAPI.BlockCommandResponse)
	assert.True(t, isBCR)
	assert.NotNil(t, bcr.Volume)
	assert.Empty(t, bcr.Error)

}

func TestBlockCommandWriteForwarding(t *testing.T) {
	ids, servers, err := makeTestCluster(2, nil)
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
		srv.volumes["vol-test"].Raft.RegisterObserver(observer)
		log.Printf("added volume to %d", i)
	}

	waitForNPeers(1, obsCh)
	leader := getLeader(servers)
	notLeader := getNotLeader(servers)

	//Send request from leader to not leader - this should then
	//be forwarded back to the leader for processing
	resp, err := leader.sendReq(leader.peers[notLeader.PeerID()], &volumesAPI.PeerRPC{
		Metadata: &volumesAPI.RPCMetadata{
			AllowForwarding: true,
			TTL:             2,
		},
		RPC: &volumesAPI.PeerRPC_BlockCommand{BlockCommand: &volumesAPI.BlockCommandRequest{
			Volume: "vol-test",
		},
		}})
	if err != nil {
		t.Fatal(err)
	}

	assert.NotNil(t, resp)
	bcr, isBCR := resp.(*volumesAPI.BlockCommandResponse)
	assert.True(t, isBCR)
	assert.NotNil(t, bcr.Volume)
	assert.Empty(t, bcr.Error)
}
