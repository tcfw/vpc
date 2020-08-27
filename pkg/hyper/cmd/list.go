package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	hyperAPI "github.com/tcfw/vpc/pkg/api/v1/hyper"
	"github.com/tcfw/vpc/pkg/hyper"
)

//newListCmd provides a command to delete vpcs
func newListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "lists local vms",
		Run: func(cmd *cobra.Command, args []string) {
			l, err := hyper.NewLocalLibVirtConn()
			if err != nil {
				log.Fatalf("failed: %s", err)
			}

			defer l.Close()

			if err := hyper.ApplyDesiredState(l, "7897fad0-e6ff-44df-9e3c-cb6c939f8bb6", hyperAPI.PowerState_SHUTDOWN); err != nil {
				log.Fatalf("%s", err)
			}

			f, err := hyper.AvailableDisks()
			if err != nil {
				log.Fatalf("%s", err)
			}
			fmt.Printf("%+v", f)

			// d, _ := l.LookupDomainByUUIDString("7897fad0-e6ff-44df-9e3c-cb6c939f8bb6")
			// if err := hyper.Snapshot(d); err != nil {
			// 	fmt.Printf("ERR: %s", err)
			// }
		},
	}

	return cmd
}
