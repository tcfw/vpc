package sbs

import (
	"github.com/lucas-clemente/quic-go"
)

const (
	errHandshakeRejected quic.ErrorCode = iota
	errDisconnect
)

type quicError interface {
	IsApplicationError() bool
}
