package cmd

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/spf13/cobra"
)

// NewDefaultCommand creates the `l3` command and its nested children.
func NewDefaultCommand() *cobra.Command {
	// Parent command to which all subcommands are added.
	cmds := &cobra.Command{
		Use:   "l3",
		Short: "l3",
		Long:  `l3 providers routing for vxlans`,
		Run: func(cmd *cobra.Command, args []string) {
			vpcs, _ := cmd.Flags().GetUintSlice("vpc")
			fmt.Printf("%v\n", vpcs)

			waitForExit()
		},
	}

	cmds.Flags().UintSlice("vpc", []uint{}, "vpc to add local router too")
	cmds.MarkFlagRequired("vpc")

	return cmds
}

func waitForExit() {
	var sigChan chan os.Signal
	sigChan = make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	<-sigChan
}
