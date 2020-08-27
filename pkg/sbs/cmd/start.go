package cmd

import (
	"log"
	"time"

	"github.com/spf13/cobra"
	"github.com/tcfw/vpc/pkg/sbs"
	"github.com/tcfw/vpc/pkg/sbs/config"
	"github.com/tcfw/vpc/pkg/utils"

	volumesAPI "github.com/tcfw/vpc/pkg/api/v1/volumes"
)

//newStartCmd provides a command to delete vpcs
func newStartCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Starts listening for cluster connections",
		Run: func(cmd *cobra.Command, args []string) {
			if err := config.Read(); err != nil {
				log.Fatalf("Failed to read config files: %s", err)
			}

			port, _ := cmd.Flags().GetInt("port")
			id, _ := cmd.Flags().GetString("id")
			bootPeers, _ := cmd.Flags().GetStringSlice("boot-peer")

			s := sbs.NewServerWithPort(port)
			if id != "" {
				s.SetPeerID(id)
			}
			defer s.Shutdown()

			if err := s.Listen(); err != nil {
				log.Fatalf("failed to start listening to remote connections: %s", err)
				return
			}

			s.BootPeers(bootPeers)

			time.Sleep(2 * time.Second)

			if err := s.AddVolume(&volumesAPI.Volume{
				Id:      "vol-test",
				Account: 1,
				Size:    10,
			}, s.PeerIDs()); err != nil {
				log.Fatalf("failed to start volume: %s", err)
				return
			}

			s.Export("vol-test")

			//Wait and gracefully shutdown
			utils.BlockUntilSigTerm()
		},
	}

	cmd.Flags().IntP("port", "p", sbs.DefaultListenPort, "quic listening port")
	cmd.Flags().String("id", "", "manual peer ID. Defaults to machine ID")
	cmd.Flags().StringSlice("boot-peer", []string{}, "a list of peers to connect to on boot in format 'peerID@addr:port'")

	return cmd
}
