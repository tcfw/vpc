package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tcfw/vpc/pkg/l2"
)

//NewDeleteCmd provides a command to delete vpcs
func NewDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Deletes VPC stacks",
		Run: func(cmd *cobra.Command, args []string) {
			vpcs, _ := cmd.Flags().GetIntSlice("vpc")

			for _, vpcID := range vpcs {
				stack, err := l2.GetVPCStack(int32(vpcID))
				if err != nil {
					fmt.Printf("Failed to find VPC stack: %s\n", err)
				}
				if err := l2.DeleteVPCStack(stack); err != nil {
					fmt.Printf("Failed to delete stack: %s", err)
					return
				}
				fmt.Printf("Deleted VPC %d\n", vpcID)
			}
		},
	}

	cmd.Flags().IntSlice("vpc", []int{}, "vpc to add local router too")
	cmd.MarkFlagRequired("vpc")

	return cmd
}
