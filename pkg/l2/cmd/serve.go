package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tcfw/vpc/pkg/l2"
)

//NewServeCmd provides a command to delete vpcs
func NewServeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Starts the l2 agent in daemon",
		Run: func(cmd *cobra.Command, args []string) {
			go l2.StartPProf(8024)

			port, _ := cmd.Flags().GetUint("port")
			l2.Serve(port)
		},
	}

	cmd.Flags().UintP("port", "p", 18254, "GRPC port")
	cmd.Flags().StringSlice("bgp", []string{}, "initial BGP peer")
	viper.BindPFlag("bgp_peers", cmd.Flags().Lookup("bgp"))

	return cmd
}
