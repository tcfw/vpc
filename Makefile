FLAGS=-ldflags="-w -s"
GO=go
GO115=go1.15beta1
GO_BUILD=${GO} build ${FLAGS}
GO115_BUILD=${GO115} build ${FLAGS}

all: l2 l3
115: l2_115 l3_115

.PHONY: l3
l3:
	${GO_BUILD} -o ./build/l3-agent ./cmd/l3agent

l3_115:
	${GO115_BUILD} -o ./build/l3-agent ./cmd/l3agent

.PHONY: l2
l2:
	${GO_BUILD} -o ./build/l2-agent ./cmd/l2agent

l2_115:
	${GO115_BUILD} -o ./build/l2-agent ./cmd/l2agent

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