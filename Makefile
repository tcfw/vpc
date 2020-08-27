FLAGS=-ldflags="-w -s"
GO=go
GO115=go1.15beta1
GO_BUILD=${GO} build ${FLAGS}
GO115_BUILD=${GO115} build ${FLAGS}

all: l2 l3 vpc sbs
115: l2_go115 l3_go115 sbs_go115 vpc

.PHONY: l3
l3:
	${GO_BUILD} -o ./build/l3-agent ./cmd/l3agent

l3_go115:
	${GO115_BUILD} -o ./build/l3-agent ./cmd/l3agent

.PHONY: l2
l2:
	${GO_BUILD} -o ./build/l2-agent ./cmd/l2agent

l2_go115:
	${GO115_BUILD} -o ./build/l2-agent ./cmd/l2agent

sbs_go115:
	${GO115_BUILD} -o ./build/sbs ./cmd/sbs

.PHONY: vpc
vpc:
	${GO_BUILD} -o ./build/vpc ./cmd/vpc

.PHONY: sbs
sbs:
	${GO_BUILD} -o ./build/sbs ./cmd/sbs

.PHONY: strip
strip:
	strip --strip-unneeded ./build/l2-agent
	strip --strip-unneeded ./build/l3-agent
	strip --strip-unneeded ./build/vpc
	strip --strip-unneeded ./build/sbs

protos:
	./scripts/protoc-gen.sh

clean:
	rm -f ./build/l3-agent
	rm -f ./build/l2-agent
	rm -f ./build/vpc
	rm -f ./build/sbs