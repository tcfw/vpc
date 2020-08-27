package utils

import (
	"os"
	"os/signal"
	"syscall"
)

//BlockUntilSigTerm waits until the application receives a SIGTERM
//usually sent to the application via ctrl+c
func BlockUntilSigTerm() {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
}
