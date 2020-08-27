package sbs

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"

	volumesAPI "github.com/tcfw/vpc/pkg/api/v1/volumes"
	"github.com/tcfw/vpc/pkg/sbs/control"
)

func TestRegister(t *testing.T) {
	controller := control.NewInMemController()

	srv := NewServerWithPort(rand.Intn(1750) + 34000)
	srv.SetPeerID(randomID())
	srv.SetController(controller)
	srv.Listen()

	assert.NotEmpty(t, controller.PeerIDs())
}

func TestDiscovery(t *testing.T) {
	controller := control.NewInMemController()
	ids, servers, err := makeTestDisconnectedCluster(2, controller)
	if err != nil {
		t.Fatal(err)
	}

	peer, err := servers[0].discoverPeer(ids[1])
	if assert.NoError(t, err) {
		assert.Equal(t, peer.PeerID, ids[1])
		assert.Len(t, servers[0].peers, 1)
	}
}

func TestAddVolume(t *testing.T) {
	controller := control.NewInMemController()

	controller.DefineVolume(&volumesAPI.Volume{
		Id:   "vol-test",
		Size: 10,
	}, false)

	srv := NewServerWithPort(rand.Intn(1750) + 34000)
	srv.SetPeerID(randomID())
	srv.SetController(controller)

	//One pass to get the update
	srv.watchController(true)

	assert.NotEmpty(t, srv.volumes)
}
