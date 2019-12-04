package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"

	"github.com/spf13/viper"

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

	cmd.Flags().IntSlice("vpc", []int{}, "VPC to add local router too")
	cmd.Flags().UintP("port", "p", 18254, "GRPC port")
	cmd.Flags().StringP("bridge", "b", "virbr0", "bridge for public network")
	cmd.Flags().Bool("dhcp", false, "Enable DHCP")
	cmd.Flags().Bool("nat", true, "Enable NAT")
	cmd.Flags().StringSlice("dhcpdns", []string{"1.1.1.1", "1.0.0.1"}, "DHCP DNS servers")
	cmd.Flags().StringSlice("bgp", []string{"192.168.222.1"}, "BGP peers")
	cmd.Flags().StringSlice("ip", []string{"10.4.0.1/24:5", "10.4.1.1/24:6"}, `Subnets/IPs (format "cidr:vlan" e.g. "10.4.0.1/24:5)`)

	viper.BindPFlag("br", cmd.Flags().Lookup("bridge"))
	viper.BindPFlag("bgp", cmd.Flags().Lookup("bgp"))
	viper.BindPFlag("dhcp", cmd.Flags().Lookup("dhcp"))
	viper.BindPFlag("ips", cmd.Flags().Lookup("ip"))

	cmd.MarkFlagRequired("vpc")

	return cmd
}

func waitForExit() {
	var signalChannel chan os.Signal
	signalChannel = make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt)
	<-signalChannel
}
