package sbs

/*
Details taken originally from
https://github.com/abligh/gonbdserver/blob/master/nbd/protocol.go
*/

// this section is in essence a transcription of the protocol from
// NBD's proto.md; note that that file is *not* GPL. For details of
// what the options mean, see proto.md

// NBD commands
const (
	NBDCMDRead = iota
	NBDCMDWrite
	NBDCMDDisc
	NBDCMDFlush
	NBDCMDTrim
	NBDCMDCache
	NBDCMDWriteZeroes
	NBDCMDBlockStatus
)

// NBD command flags
const (
	NBDCMDFlagFUA = uint16(1 << iota)
	NBDCMDFlagMayTrim
	NBDCMDFlagDF
)

// NBD negotiation flags
const (
	NBDFLAGHasFlags = uint16(1 << iota)
	NBDFLAGReadOnly
	NBDFLAGSendFlush
	NBDFLAGSendFUA
	NBDFLAGRotational
	NBDFLAGSendTrim
	NBDFLAGSendWriteZeroes
	NBDFLAGSendDF
	NBDFLAGMultiCon
	NBDFLAGSendResize
	NBDFLAGSendCache
	NBDFLAGFastZero
)

// NBD magic numbers
const (
	NBDMAGIC                = 0x4e42444d41474943
	NBDMAGICRequest         = 0x25609513
	NBDMAGICReply           = 0x67446698
	NBDMAGICCliserv         = 0x00420281861253
	NBDMAGICOpts            = 0x49484156454F5054
	NBDMAGICRep             = 0x3e889045565a9
	NBDMAGICStructuredReply = 0x668e33ef

	// NBD default port
	NBDDefaultPort = 10809
)

// NBD options
const (
	NBDOPTExportName = iota + 1
	NBDOPTAbort
	NBDOPTList
	NBDOPTPeekExport
	NBDOPTStartTLS
	NBDOPTInfo
	NBDOPTGo
	NBDOPTStructuredReply

	// NBD option reply types
	NBDREPAck              = uint32(1)
	NBDREPServer           = uint32(2)
	NBDREPInfo             = uint32(3)
	NBDREPFlagError        = uint32(1 << 31)
	NBDREPErrUnsup         = uint32(1 | NBDREPFlagError)
	NBDREPErrPolicy        = uint32(2 | NBDREPFlagError)
	NBDREPErrInvalid       = uint32(3 | NBDREPFlagError)
	NBDREPErrPlatform      = uint32(4 | NBDREPFlagError)
	NBDREPErrTLSReqd       = uint32(5 | NBDREPFlagError)
	NBDREPErrUnknown       = uint32(6 | NBDREPFlagError)
	NBDREPErrShutdown      = uint32(7 | NBDREPFlagError)
	NBDREPErrBlockSizeReqd = uint32(8 | NBDREPFlagError)

	// NBD reply flags
	NBDReplyFlagDone = 1 << 0
)

// NBD reply types
const (
	NBDREPLYTYPENone = iota
	NBDREPLYTYPEError
	NBDREPLYTYPEErrorOffset
	NBDREPLYTYPEOffsetData
	NBDREPLYTYPEOffsetHole
)

// NBD hanshake flags
const (
	NBDFLAGFixedNewstyle = 1 << iota
	NBDFLAGNoZeroes
)

// NBD client flags
const (
	NBDFLAGCFixedNewstyle = 1 << iota
	NBDFLAGCNoZeroes

	// NBD errors
	NBDEPERM     = 1
	NBDEIO       = 5
	NBDENOMEM    = 12
	NBDEINVAL    = 22
	NBDENOSPC    = 28
	NBDEOVERFLOW = 75
)

// NBD info types
const (
	NBDINFOExport = iota
	NBDINFOName
	NBDINFODescription
	NBDINFOBlockSize
)

// NBD new style header
type nbdNewStyleHeader struct {
	NbdMagic       uint64
	NbdOptsMagic   uint64
	NbdGlobalFlags uint16
}

// NBD client flags
type nbdClientFlags struct {
	NbdClientFlags uint32
}

// NBD client options
type nbdClientOpt struct {
	NbdOptMagic uint64
	NbdOptID    uint32
	NbdOptLen   uint32
}

// NBD export details
type nbdExportDetails struct {
	NbdExportSize  uint64
	NbdExportFlags uint16
}

// NBD option reply
type nbdOptReply struct {
	NbdOptReplyMagic  uint64
	NbdOptID          uint32
	NbdOptReplyType   uint32
	NbdOptReplyLength uint32
}

// NBD request
type nbdRequest struct {
	NbdRequestMagic uint32
	NbdCommandFlags uint16
	NbdCommandType  uint16
	NbdHandle       uint64
	NbdOffset       uint64
	NbdLength       uint32
}

// NBD simple reply
type nbdReply struct {
	NbdReplyMagic uint32
	NbdError      uint32
	NbdHandle     uint64
}

type nbdStructuredReply struct {
	NbdReplyMagic uint32
	NbdFlags      uint16
	NbdType       uint16
	NbdHandle     uint64
	NbdLength     uint32
}

type nbdStructuredError struct {
	NbdError  uint32
	NbdLength uint16
}

// NBD info export
type nbdInfoExport struct {
	NbdInfoType          uint16
	NbdExportSize        uint64
	NbdTransmissionFlags uint16
}

// NBD info blocksize
type nbdInfoBlockSize struct {
	NbdInfoType           uint16
	NbdMinimumBlockSize   uint32
	NbdPreferredBlockSize uint32
	NbdMaximumBlockSize   uint32
}
