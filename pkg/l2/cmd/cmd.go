package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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

	cmds.AddCommand(NewCheckCmd())
	cmds.AddCommand(NewServeCmd())
	cmds.AddCommand(NewWatchCmd())

	cmds.PersistentFlags().StringP("vtepdev", "i", "eth0", "Device for VTEP endpoint")
	viper.BindPFlag("vtepdev", cmds.PersistentFlags().Lookup("vtepdev"))

	return cmds
}
