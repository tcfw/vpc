package machines

import (
	"log"

	"github.com/denisbrodbeck/machineid"
)

//UID returns the unique ID of the local machine
func UID() string {
	id, err := machineid.ProtectedID("tcfw.vpc")
	if err != nil {
		log.Fatal(err)
	}
	return id
}
