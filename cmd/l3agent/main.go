package main

import (
	"fmt"
	"os"

	"github.com/tcfw/vpc/pkg/l3/cmd"
)

func main() {
	cmd := cmd.NewDefaultCommand()

	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	// vpcID := int32(1)

	// stack, err := l2.GetVPCStack(vpcID)
	// if err != nil {
	// 	fmt.Printf("Failed to get stack: %s", err)
	// 	os.Exit(1)
	// }

	// r, err := l3.CreateRouter(stack, "a8s9d")
	// if err != nil {
	// 	fmt.Printf("Failed to create router: %s", err)
	// 	os.Exit(1)
	// }

	// waitForExit()

	// r.Delete()
}
