FLAGS=-ldflags="-w -s"
GO=go
GO_BUILD=${GO} build ${FLAGS}

all: l2 l3

.PHONY: l2
l2:
	${GO_BUILD} -o ./build/l3-agent ./cmd/l3agent

.PHONY: l3
l3:
	${GO_BUILD} -o ./build/l2-agent ./cmd/l2agent

.PHONY: strip
strip:
	strip --strip-unneeded ./build/l2-agent
	strip --strip-unneeded ./build/l3-agent