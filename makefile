BINARIES=server device cli
BUILD_DIR=bin

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

build: proto $(BINARIES)

server:
	go build -o $(BUILD_DIR)/server ./cmd/server

device:
	go build -o $(BUILD_DIR)/device ./cmd/device

cli:
	go build -o $(BUILD_DIR)/cli ./cmd/cli

clean:
	rm -rf $(BUILD_DIR)
	rm -rf ./api/gen

run: build
	@echo "Running..."
	@$(BUILD_DIR)/server & echo $$! > $(BUILD_DIR)/server.pid
	@$(BUILD_DIR)/device & echo $$! > $(BUILD_DIR)/device.pid
	@$(BUILD_DIR)/cli & echo $$! > $(BUILD_DIR)/cli.pid
	@echo "Complete"

stop:
	@echo "Stopping..."
	@kill `cat $(BUILD_DIR)/server.pid` || true
	@kill `cat $(BUILD_DIR)/device.pid` || true
	@kill `cat $(BUILD_DIR)/cli.pid` || true
	@rm -f $(BUILD_DIR)/*.pid
	@echo "Complete"