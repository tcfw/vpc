package cmd

import (
	"github.com/spf13/cobra"
	"github.com/tcfw/vpc/pkg/vpc"
)

//NewServeCmd provides a command to delete vpcs
func NewServeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Starts the vpc api service",
		Run: func(cmd *cobra.Command, args []string) {
			port, _ := cmd.Flags().GetUint("port")
			vpc.Serve(port)
		},
	}

	cmd.Flags().UintP("port", "p", 18254, "GRPC port")

	return cmd
}
