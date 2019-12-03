package l2

import (
	"fmt"
	"log"
	"time"

	l2API "github.com/tcfw/vpc/pkg/api/v1/l2"
	"github.com/vishvananda/netlink"
)

//Gc watches for vpcs which have no nics anymore and closes them off
func (s *Server) Gc() {
	for {
		s.m.Lock()

		for _, stack := range s.stacks {
			if stack.Bridge == nil {
				continue
			}
			links, err := getBridgeLinks(stack.Bridge.Index, stack.VPCID)
			if err != nil {
				log.Println(err)
				continue
			}
			if len(links) == 0 {
				//delete the stack since no devices apart from vtep remaining
				vpcID := stack.VPCID
				err = DeleteVPCStack(stack)
				if err != nil {
					log.Println(err)
					continue
				}

				s.transport.DelVTEP(uint32(vpcID))

				delete(s.stacks, vpcID)

				s.bgp.DeregisterEP(uint32(vpcID))

				s.logChange(&l2API.StackChange{
					VpcId:  vpcID,
					Action: "gc",
				})
			}
		}

		s.m.Unlock()
		time.Sleep(5 * time.Second)
	}
}

func getBridgeLinks(brIndex int, vpcID int32) ([]netlink.Link, error) {
	slaveLinks := []netlink.Link{}

	handle, err := netlink.NewHandle(netlink.FAMILY_ALL)
	if err != nil {
		return nil, err
	}
	defer func() {
		handle.Delete()
	}()

	links, err := handle.LinkList()
	if err != nil {
		return nil, err
	}

	vtepName := fmt.Sprintf("t-%d", vpcID)

	for _, link := range links {
		if link.Attrs().MasterIndex == brIndex && link.Attrs().Name != vtepName {
			slaveLinks = append(slaveLinks, link)
		}
	}

	return slaveLinks, nil
}
