package l2

import (
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
			links, err := getBridgeLinks(stack.Bridge.Index)
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

				delete(s.stacks, vpcID)

				s.logChange(&l2API.StackChange{
					VpcId:  vpcID,
					Action: "gc",
				})
			}
		}
		s.m.Unlock()

		time.Sleep(1 * time.Minute)
	}
}

func getBridgeLinks(brIndex int) ([]netlink.Link, error) {
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
	for _, link := range links {
		if link.Attrs().MasterIndex == brIndex && link.Type() != "vxlan" {
			slaveLinks = append(slaveLinks, link)
		}
	}

	return slaveLinks, nil
}
