#!/bin/bash

services='machinery vpc l2 hyper volumes l2Controller'

for service in ${services}
do
	echo "Generating service ${service}...";
	mkdir -p pkg/api/v1/${service}
	protoc --proto_path=api/proto/v1 --go_out=pkg/api/v1/${service} --go-grpc_out=pkg/api/v1/${service} ${service}.proto
	protoc --proto_path=api/proto/v1 --swagger_out=logtostderr=true:api/swagger/v1 ${service}.proto
done