FLAGS=-ldflags="-w -s"
GO=go
GO_BUILD=${GO} build ${FLAGS}

all: l2 l3

.PHONY: l3
l3:
	${GO_BUILD} -o ./build/l3-agent ./cmd/l3agent

.PHONY: l2
l2:
	${GO_BUILD} -o ./build/l2-agent ./cmd/l2agent

.PHONY: vpc
vpc:
	${GO_BUILD} -o ./build/vpc ./cmd/vpc

.PHONY: strip
strip:
	strip --strip-unneeded ./build/l2-agent
	strip --strip-unneeded ./build/l3-agent
	strip --strip-unneeded ./build/vpc

protos:
	./scripts/protoc-gen.sh

clean:
	rm -f ./build/l3-agent
	rm -f ./build/l2-agent
	rm -f ./build/vpc