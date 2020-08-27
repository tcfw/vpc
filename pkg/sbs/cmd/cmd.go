package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// NewDefaultCommand creates the `hyper` command and its nested children.
func NewDefaultCommand() *cobra.Command {
	// Parent command to which all subcommands are added.
	cmds := &cobra.Command{
		Use:   "sbs",
		Short: "sbs",
		Long:  `sbs providers simple block storage resources`,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	cmds.AddCommand(newStartCmd())
	cmds.AddCommand(newAgentCmd())

	cmds.PersistentFlags().Int16P("log-level", "v", 3, "log level (5=trace, 4=debug, 3=warning, 2=error, 1=fatal, 0=panic")
	viper.BindPFlag("LogLevel", cmds.PersistentFlags().Lookup("log-level"))

	return cmds
}
