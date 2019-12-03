package cmd

import (
	"fmt"
	"strings"

	"github.com/vishvananda/netlink"

	"github.com/spf13/cobra"
	"github.com/tcfw/vpc/pkg/l2"
)

//NewCheckCmd provides a command check and validate status of vpcs
func NewCheckCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "check",
		Short: "Checks VPC stack status",
		Run: func(cmd *cobra.Command, args []string) {
			vpcs, _ := cmd.Flags().GetIntSlice("vpc")

			for _, vpcID := range vpcs {
				id := int32(vpcID)

				stack, err := l2.GetVPCStack(id)
				if err != nil {
					fmt.Printf("Failed to find VPC stack: %s\n", err)
				}

				if stack.Bridge == nil && stack.Vtep == nil {
					fmt.Printf("VPC %d is missing\n", id)
					continue
				} else {
					fmt.Printf("VPC %d Status:\n", id)
				}

				ok, err := l2.HasVPCBridge(id)
				if err != nil {
					fmt.Printf("Failed to query bridge: %s\n", err)
				}
				if !ok {
					fmt.Println("\xE2\x9D\x8C  Bridge is missing")
				} else if stack.Bridge.OperState != netlink.OperUp {
					fmt.Printf("\xE2\x9D\x93  Bridge is in %s state\n", strings.ToUpper(stack.Bridge.OperState.String()))
				} else {
					fmt.Println("\xE2\x9C\x85  Bridge is UP")
				}

				ok, err = l2.HasVTEP(id)
				if err != nil {
					fmt.Printf("Failed to query VTEP: %s\n", err)
				}
				if !ok {
					fmt.Println("\xE2\x9D\x8C  VTEP is missing")
				} else if stack.Vtep.Attrs().OperState != netlink.OperUp {
					fmt.Printf("\xE2\x9D\x93  VTEP is in %s state\n", strings.ToUpper(stack.Vtep.Attrs().OperState.String()))
				} else {
					fmt.Println("\xE2\x9C\x85  VTEP is UP")
				}

				fmt.Println()
			}
		},
	}

	cmd.Flags().IntSlice("vpc", []int{}, "vpc to add local router too")
	cmd.MarkFlagRequired("vpc")

	return cmd
}
