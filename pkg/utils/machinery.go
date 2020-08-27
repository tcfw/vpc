package utils

import (
	"github.com/denisbrodbeck/machineid"
)

//MachineID gets the static machine ID
func MachineID() string {
	id, _ := machineid.ProtectedID("VPC")
	return id
}
