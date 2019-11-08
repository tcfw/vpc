package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"

	"github.com/tcfw/vpc/l2"
	"github.com/tcfw/vpc/l3"
)

func main() {
	flag.Parse()

	vpcID := int32(1)

	stack, err := l2.GetVPCStack(vpcID)
	if err != nil {
		fmt.Printf("Failed to get stack: %s", err)
		os.Exit(1)
	}

	r, err := l3.CreateRouter(stack, "a8s9d")
	if err != nil {
		fmt.Printf("Failed to create router: %s", err)
		os.Exit(1)
	}

	waitForExit()

	r.Delete()
}

func waitForExit() {
	var sigChan chan os.Signal
	sigChan = make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	<-sigChan
}
