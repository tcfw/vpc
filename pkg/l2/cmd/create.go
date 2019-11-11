package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tcfw/vpc/pkg/l2"
)

//NewCreateCmd provides a command to create vpcs
func NewCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Creates VPC stacks",
		Run: func(cmd *cobra.Command, args []string) {
			vpcs, _ := cmd.Flags().GetIntSlice("vpc")

			for _, vpcID := range vpcs {
				if vpcID > 16777215 {
					fmt.Printf("VPC %d id out of range\n", vpcID)
					continue
				}
				vtepdev, _ := cmd.Flags().GetString("vtepdev")
				if _, err := l2.CreateVPCStack(int32(vpcID), vtepdev); err != nil {
					fmt.Printf("Failed to create stack: %s", err)
					return
				}
				fmt.Printf("Created VPC %d\n", vpcID)
			}
		},
	}

	cmd.Flags().IntSlice("vpc", []int{}, "vpc to add local router too")
	cmd.Flags().StringP("vtepdev", "i", "eth0", "VTEP local interface")
	cmd.MarkFlagRequired("vpc")

	return cmd
}
