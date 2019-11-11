package cmd

import (
	"github.com/spf13/cobra"
)

// NewDefaultCommand creates the `l2` command and its nested children.
func NewDefaultCommand() *cobra.Command {
	// Parent command to which all subcommands are added.
	cmds := &cobra.Command{
		Use:   "l2",
		Short: "l2",
		Long:  `l2 providers vxlan bridges`,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	cmds.AddCommand(NewCreateCmd())
	cmds.AddCommand(NewDeleteCmd())
	cmds.AddCommand(NewCheckCmd())

	return cmds
}
