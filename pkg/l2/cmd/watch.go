package cmd

import (
	"context"
	"fmt"
	"log"

	"github.com/spf13/cobra"
	l2API "github.com/tcfw/vpc/pkg/api/v1/l2"
	"google.golang.org/grpc"
)

//NewWatchCmd provides a command to delete vpcs
func NewWatchCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "watch",
		Short: "Watches for changes in VPC stacks",
		Run: func(cmd *cobra.Command, args []string) {
			port, _ := cmd.Flags().GetUint("port")

			conn, err := grpc.Dial(fmt.Sprintf("127.0.0.1:%d", port), grpc.WithInsecure())
			if err != nil {
				log.Println(err)
				return
			}
			defer conn.Close()

			cli := l2API.NewL2ServiceClient(conn)
			stream, err := cli.WatchStacks(context.Background(), &l2API.Empty{})
			if err != nil {
				log.Println(err)
				return
			}

			for {
				change, err := stream.Recv()
				if err != nil {
					log.Println(err)
					return
				}

				log.Printf("%d %s", change.VpcId, change.Action)
			}
		},
	}

	cmd.Flags().UintP("port", "p", 18254, "GRPC port")

	return cmd
}
