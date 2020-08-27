package main

import (
	"fmt"
	"os"

	"github.com/tcfw/vpc/pkg/sbs/cmd"
)

func main() {
	cmd := cmd.NewDefaultCommand()

	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
