BINARIES=server device cli
BUILD_DIR=bin
FILEEXT=".exe"

ifeq ($(OS), Windows_NT)
	FILEEXT=".exe"
else
	FILEEXT=""
endif

.PHONY: all configure proto build clean run stop $(BINARIES)

all: configure build

proto:
	mkdir -p ./api/gen/go/management ./api/gen/go/control
	protoc -I=api/proto management.proto \
		--go_out=api/gen/go/management \
		--go_opt=paths=source_relative \
		--go-grpc_out=api/gen/go/management \
		--go-grpc_opt=paths=source_relative
		
	protoc -I=api/proto control.proto \
		--go_out=api/gen/go/control \
		--go_opt=paths=source_relative \
		--go-grpc_out=api/gen/go/control \
		--go-grpc_opt=paths=source_relative

configure:
	go mod tidy
	go mod download
	proto

build: clean proto $(BINARIES)

server:
	go build -o $(BUILD_DIR)/server$(FILEEXT) ./cmd/server

device:
	go build -o $(BUILD_DIR)/device$(FILEEXT) ./cmd/device

cli:
	go build -o $(BUILD_DIR)/cli$(FILEEXT) ./cmd/cli

clean:
	rm -rf $(BUILD_DIR)
	rm -rf ./api/gen
