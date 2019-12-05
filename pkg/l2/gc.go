package l2

import (
	"log"
	"time"

	l2API "github.com/tcfw/vpc/pkg/api/v1/l2"
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
				err = deleteStack(stack)
				if err != nil {
					log.Println(err)
					continue
				}

				s.transport.DelEP(uint32(vpcID))

				delete(s.stacks, vpcID)

				s.sdn.DeregisterEP(uint32(vpcID))

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
