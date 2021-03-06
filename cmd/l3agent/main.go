package main

import (
	"fmt"
	"os"

	"github.com/tcfw/vpc/pkg/l3/cmd"
)

func main() {
	dcmd := cmd.NewDefaultCommand()

	if err := dcmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
