.PHONY: all build run test clean proto-gen

all: build

build:
	@echo "Building..."
	@go build -v -o bin/server ./cmd/server

run:
	@echo "Running..."
	@./bin/server

test:
	@echo "Testing..."
	@go test -v ./...

clean:
	@echo "Cleaning..."
	@rm -f bin/server
	@rm -f pkg/mikrotik/*.pb.go

proto-gen:
	@echo "Generating protobuf code..."
	@protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    proto/mikrotik.proto

