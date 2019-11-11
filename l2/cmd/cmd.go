package cmd

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/spf13/cobra"
	"github.com/tcfw/vpc/l2"
)

// NewDefaultCommand creates the `l2` command and its nested children.
func NewDefaultCommand() *cobra.Command {
	// Parent command to which all subcommands are added.
	cmds := &cobra.Command{
		Use:   "l2",
		Short: "l2",
		Long:  `l2 providers vxlan bridges`,
		Run: func(cmd *cobra.Command, args []string) {
			vpcs, _ := cmd.Flags().GetIntSlice("vpc")

			stacks := []*l2.Stack{}

			for _, vpcID := range vpcs {
				stack, err := l2.CreateVPCStack(int32(vpcID))
				if err != nil {
					fmt.Printf("Failed to create stack: %s", err)
					return
				}
				fmt.Printf("Created VPC %d\n", vpcID)

				stacks = append(stacks, stack)
			}

			waitForExit()

			for _, stack := range stacks {
				if err := l2.DeleteVPCStack(stack); err != nil {
					fmt.Printf("Failed to delete stack: %s", err)
					return
				}
			}
		},
	}

	cmds.Flags().IntSlice("vpc", []int{}, "vpc to add local router too")
	cmds.MarkFlagRequired("vpc")

	return cmds
}

func waitForExit() {
	var sigChan chan os.Signal
	sigChan = make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	<-sigChan
}
