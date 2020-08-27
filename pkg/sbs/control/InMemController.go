package control

import (
	"fmt"
	"net"
	"sync"

	volumesAPI "github.com/tcfw/vpc/pkg/api/v1/volumes"
)

//InMemController in-memory controller; useful for tests
//Volumes will be assigned to all peers
type InMemController struct {
	Volumes     map[string]*volumesAPI.Volume
	Attachments map[string]*volumesAPI.Volume

	volumeChs     []chan VolumeAction
	attachmentChs []chan VolumeAction
	peers         map[string]*net.UDPAddr
	mu            sync.Mutex
}

//NewInMemController provides a new in-memory controller instance for testing
func NewInMemController() *InMemController {
	return &InMemController{
		Volumes:       map[string]*volumesAPI.Volume{},
		Attachments:   map[string]*volumesAPI.Volume{},
		volumeChs:     []chan VolumeAction{},
		attachmentChs: []chan VolumeAction{},
		peers:         map[string]*net.UDPAddr{},
	}
}

//VolumesCh provides updates to storage nodes for adding or removing volumes
func (imc *InMemController) VolumesCh(existing bool) <-chan VolumeAction {
	ch := make(chan VolumeAction)
	imc.volumeChs = append(imc.volumeChs, ch)

	if existing {
		go imc.sendExisting(ch)
	}

	return ch
}

//AttachmentsCh provides updates to agent nodes for adding or removing volume attachments
func (imc *InMemController) AttachmentsCh(existing bool) <-chan VolumeAction {
	ch := make(chan VolumeAction)
	imc.attachmentChs = append(imc.attachmentChs, ch)
	if existing {
		go imc.sendExisting(ch)

	}
	return ch
}

func (imc *InMemController) sendExisting(ch chan VolumeAction) {
	imc.mu.Lock()
	defer imc.mu.Unlock()

	peerIDs := imc.PeerIDs()

	for _, vol := range imc.Volumes {
		ch <- VolumeAction{
			Op:    OpAdd,
			Desc:  vol,
			Peers: peerIDs,
		}
	}
}

//PeerIDs provides a list of connected peers
func (imc *InMemController) PeerIDs() []string {
	peerIDs := []string{}
	for peer := range imc.peers {
		peerIDs = append(peerIDs, peer)
	}
	return peerIDs
}

//Discover provides a mechanism for finding peer endpoints
func (imc *InMemController) Discover(peer string) (*net.UDPAddr, error) {
	imc.mu.Lock()
	defer imc.mu.Unlock()

	addr, ok := imc.peers[peer]
	if !ok {
		return nil, fmt.Errorf("not found")
	}

	return addr, nil
}

//Register allows a server to register it's endpoint
func (imc *InMemController) Register(id string, addr *net.UDPAddr) error {
	imc.mu.Lock()
	defer imc.mu.Unlock()

	imc.peers[id] = addr
	return nil
}

//DefineVolume adds a new volume and notifies all peers of new volume
func (imc *InMemController) DefineVolume(vol *volumesAPI.Volume, isUpdate bool) {
	imc.mu.Lock()
	defer imc.mu.Unlock()

	imc.Volumes[vol.Id] = vol

	peerIDs := imc.PeerIDs()

	op := OpAdd
	if isUpdate {
		op = OpUpdate
	}
	for _, ch := range imc.volumeChs {
		ch <- VolumeAction{
			Op:    op,
			Desc:  vol,
			Peers: peerIDs,
		}
	}
}

//DefineAttachment adds a new attachment and notifies all peers of new volume
func (imc *InMemController) DefineAttachment(vol *volumesAPI.Volume, isUpdate bool) {
	imc.mu.Lock()
	defer imc.mu.Unlock()

	imc.Attachments[vol.Id] = vol

	peerIDs := imc.PeerIDs()

	op := OpAdd
	if isUpdate {
		op = OpUpdate
	}
	for _, ch := range imc.attachmentChs {
		ch <- VolumeAction{
			Op:    op,
			Desc:  vol,
			Peers: peerIDs,
		}
	}
}

//RemoveVolume removes a volme along with attachments and notifies all peers of the change
func (imc *InMemController) RemoveVolume(id string) {
	imc.mu.Lock()
	defer imc.mu.Unlock()

	vol, ok := imc.Volumes[id]
	if !ok {
		return
	}

	delete(imc.Volumes, vol.Id)
	delete(imc.Attachments, vol.Id)

	for _, ch := range imc.attachmentChs {
		ch <- VolumeAction{
			Op:   OpRemove,
			Desc: vol,
		}
	}

	for _, ch := range imc.volumeChs {
		ch <- VolumeAction{
			Op:   OpRemove,
			Desc: vol,
		}
	}
}

//Peers provides information about which peers should contain volumes
func (imc *InMemController) Peers(volumeID string) ([]*PeerInfo, error) {
	peerInfo := []*PeerInfo{}

	for peer, addr := range imc.peers {
		peerInfo = append(peerInfo, &PeerInfo{
			ID:   peer,
			Addr: addr,
		})
	}

	return peerInfo, nil
}
