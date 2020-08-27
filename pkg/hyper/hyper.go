package hyper

import (
	"fmt"
	"log"
	"time"

	hyperAPI "github.com/tcfw/vpc/pkg/api/v1/hyper"
	libvirt "libvirt.org/libvirt-go"
)

func List(c *libvirt.Connect) {
	domains, err := c.ListAllDomains(libvirt.CONNECT_LIST_DOMAINS_ACTIVE | libvirt.CONNECT_LIST_DOMAINS_INACTIVE)
	if err != nil {
		log.Fatalf("failed to retrieve domains: %v", err)
	}

	fmt.Println("ID\tName\t\tUUID\tState")
	fmt.Printf("--------------------------------------------------------\n")
	for _, d := range domains {
		state, _, err := d.GetState()
		if err != nil {
			log.Fatalln(err)
		}
		id, _ := d.GetID()
		name, _ := d.GetName()
		uuid, _ := d.GetUUID()
		fmt.Printf("%d\t%s\t%x\t%v\n", id, name, uuid, state)
	}
}

func ApplyDesiredState(c *libvirt.Connect, uuid string, desiredState hyperAPI.PowerState) error {
	d, err := c.LookupDomainByUUIDString(uuid)
	if err != nil {
		return fmt.Errorf("failed to lookup: %s", err)
	}

	current, err := setDesirePower(d, desiredState)

	fmt.Printf("%+v", current)

	return err
}

func Stats(c *libvirt.Connect, uuid string) error {
	//Check exists
	d, err := c.LookupDomainByUUIDString(uuid)
	if err != nil {
		return fmt.Errorf("failed to lookup: %s", err)
	}

	//Check is running
	s, _, err := d.GetState()
	if err != nil {
		return fmt.Errorf("failed check state: %s", err)
	}
	if s != libvirt.DOMAIN_RUNNING {
		return fmt.Errorf("domain not running: %d", s)
	}

	memStats, err := memStats(d)
	if err != nil {
		return err
	}

	fmt.Printf("STAT: %+v\n", memStats)
	fmt.Printf("STAT Used: %+v\n", memStats.Available-memStats.Unused)
	fmt.Printf("STAT Last: %s\n", time.Unix(int64(memStats.LastUpdate), 0))

	cpuStats, total, err := cpuStats(d)
	if err != nil {
		return err
	}

	fmt.Printf("%+v\n", cpuStats)
	fmt.Printf("Total: %+#v\n", total)

	netStats, err := netStats(d)
	if err != nil {
		return err
	}

	fmt.Printf("NET: %+v\n", netStats)

	_, dTotal, err := diskStats(d)
	if err != nil {
		return err
	}

	fmt.Printf("DISK: %+v\n", dTotal)

	return nil
}
