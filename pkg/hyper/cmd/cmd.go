package cmd

import (
	"github.com/spf13/cobra"
)

// NewDefaultCommand creates the `hyper` command and its nested children.
func NewDefaultCommand() *cobra.Command {
	// Parent command to which all subcommands are added.
	cmds := &cobra.Command{
		Use:   "hyper",
		Short: "hyper",
		Long:  `hyper providers hypervised compute resources`,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	cmds.AddCommand(newListCmd())

	return cmds
}
