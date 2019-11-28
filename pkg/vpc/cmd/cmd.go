package cmd

import (
	"github.com/spf13/cobra"
)

// NewDefaultCommand creates the `l2` command and its nested children.
func NewDefaultCommand() *cobra.Command {
	// Parent command to which all subcommands are added.
	cmds := &cobra.Command{
		Use:   "vpc",
		Short: "vpc",
		Long:  `vpc providers vpc and subnet configuration`,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	cmds.AddCommand(NewServeCmd())
	cmds.AddCommand(NewMigrateCmd())

	return cmds
}
