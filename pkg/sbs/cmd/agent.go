package cmd

import (
	"log"

	"github.com/spf13/cobra"
	"github.com/tcfw/vpc/pkg/sbs/config"
)

//newAgentCmd provides a command to delete vpcs
func newAgentCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "agent",
		Short: "Starts listening for NBD connections",
		Run: func(cmd *cobra.Command, args []string) {
			if err := config.Read(); err != nil {
				log.Fatalf("Failed to read config files: %s", err)
			}
		},
	}

	return cmd
}
