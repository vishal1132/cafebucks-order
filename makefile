.Phony: protogo dev help

.DEFAULT_GOAL=help

help:
	@echo "protogo 				for compiling .proto file for go grpc plugin"
	@echo "dev					for running the server"
	@echo "buildImage   				for building docker image"

protogo:
	@protoc -I protos/ protos/order.proto --go_out=plugins=grpc:protos/order

dev:
	@go build -o order ./server && PORT=8080 ./order

buildImage:
	@docker build -t order .