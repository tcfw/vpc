package cmd

import (
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
			cmd.Help()
		},
	}

	cmds.AddCommand(NewCreateCmd())

	return cmds
}
