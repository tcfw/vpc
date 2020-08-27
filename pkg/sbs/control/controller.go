package control

import (
	"net"

	volumesAPI "github.com/tcfw/vpc/pkg/api/v1/volumes"
)

//Controller provides cluster level management of individual SBS nodes
type Controller interface {
	//VolumesCh provides updates to storage nodes for adding or removing volumes
	//If existing is true, all existing volumes wil be streamed to the chan
	VolumesCh(exiting bool) <-chan VolumeAction

	//AttachmentsCh provides updates to agent nodes for adding or removing volume attachments
	//If existing is true, all existing attachments wil be streamed to the chan
	AttachmentsCh(exiting bool) <-chan VolumeAction

	//Discover provides a mechanism for finding peer endpoints
	Discover(peer string) (*net.UDPAddr, error)

	//Register allows a server to register it's endpoint
	Register(peer string, addr *net.UDPAddr) error

	//Peers provides information about which peers should contain volumes
	Peers(volumeID string) ([]*PeerInfo, error)
}

//ActionOp operation being applied to a volume
type ActionOp int32

const (
	//OpAdd add volume
	OpAdd ActionOp = iota
	//OpRemove remove volume
	OpRemove
	//OpUpdate update config of volume
	OpUpdate
)

//VolumeAction provides update data of volume changes
type VolumeAction struct {
	Op    ActionOp
	Desc  *volumesAPI.Volume
	Peers []string
}

//PeerInfo contains basic info about contacting a peer
type PeerInfo struct {
	ID   string
	Addr *net.UDPAddr
}
