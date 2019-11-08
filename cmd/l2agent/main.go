package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"

	l2 "github.com/tcfw/vpc/l2"
)

func main() {
	flag.Parse()

	vpcID := int32(1)

	stack, err := l2.CreateVPCStack(vpcID)
	if err != nil {
		fmt.Printf("Failed to create stack: %s", err)
		return
	}

	waitForExit()

	if err := l2.DeleteVPCStack(stack); err != nil {
		fmt.Printf("Failed to delete stack: %s", err)
		return
	}
}

func waitForExit() {
	var sigChan chan os.Signal
	sigChan = make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	<-sigChan
}
