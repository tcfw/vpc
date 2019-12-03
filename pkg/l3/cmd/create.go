package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"

	"github.com/spf13/cobra"
	l2API "github.com/tcfw/vpc/pkg/api/v1/l2"
	l3 "github.com/tcfw/vpc/pkg/l3"
	"google.golang.org/grpc"
)

//NewCreateCmd provides a command to create vpcs
func NewCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Creates a l3 router on a VPC",
		Run: func(cmd *cobra.Command, args []string) {
			vpcs, _ := cmd.Flags().GetIntSlice("vpc")
			port, _ := cmd.Flags().GetUint("port")

			for _, vpcID := range vpcs {
				if vpcID > 16777215 {
					fmt.Printf("VPC %d id out of range\n", vpcID)
					continue
				}

				conn, err := grpc.Dial(fmt.Sprintf("127.0.0.1:%d", port), grpc.WithInsecure())
				if err != nil {
					log.Println(err)
					return
				}
				defer conn.Close()

				l2cli := l2API.NewL2ServiceClient(conn)

				ctx := context.Background()

				var resp *l2API.StackResponse

				resp, err = l2cli.GetStack(ctx, &l2API.StackRequest{VpcId: int32(vpcID)})
				if err != nil {
					if strings.Contains(err.Error(), "Stack not created") {
						resp, err = l2cli.AddStack(ctx, &l2API.StackRequest{VpcId: int32(vpcID)})
						if err != nil {
							log.Println("Failed to create stack: ", err)
							return
						}
					} else {
						log.Println("Failed to fetch stack: ", err)
						return
					}
				}

				stack, err := l3.L2APIToStack(resp.Stack)
				if err != nil {
					log.Println("Failed to convert stack: ", err)
					return
				}

				id := l3.NewID()

				r, err := l3.CreateRouter(l2cli, stack, id)
				if err != nil {
					log.Println("Failed to create router: ", err)
					return
				}
				defer func() {
					r.Delete()
				}()

				fmt.Printf("Created Router %s on %d\n", r.ID, vpcID)
			}

			waitForExit()
		},
	}

	cmd.Flags().IntSlice("vpc", []int{}, "vpc to add local router too")
	cmd.Flags().UintP("port", "p", 18254, "GRPC port")

	cmd.MarkFlagRequired("vpc")

	return cmd
}

func waitForExit() {
	var signalChannel chan os.Signal
	signalChannel = make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt)
	<-signalChannel
}
