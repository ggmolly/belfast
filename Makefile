PROTO_DIR := internal/proto
PROTO_GO_DIR := internal/protobuf
PROTO_LUA_SCRIPT := internal/tools/proto_from_lua.py
PROTOC_GEN_GO := $(shell go env GOPATH)/bin/protoc-gen-go

.PHONY: lua-proto proto go all swag install-protoc-gen-go build build-belfast build-gateway clean fclean re

COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
LDFLAGS := -X github.com/ggmolly/belfast/internal/buildinfo.Commit=$(COMMIT)
BINARY_DIR ?= bin

lua-proto:
	python $(PROTO_LUA_SCRIPT)

install-protoc-gen-go:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest

proto: lua-proto install-protoc-gen-go
	protoc --plugin=protoc-gen-go=$(PROTOC_GEN_GO) --proto_path=$(PROTO_DIR) --go_out=$(PROTO_GO_DIR) --go_opt=paths=source_relative $(PROTO_DIR)/*.proto

go: proto

all: lua-proto proto

swag:
	go run github.com/swaggo/swag/cmd/swag init -g cmd/belfast/main.go

build: build-belfast build-gateway

build-belfast:
	@mkdir -p $(BINARY_DIR)
	go build -ldflags "$(LDFLAGS)" -o $(BINARY_DIR)/belfast ./cmd/belfast

build-gateway:
	@mkdir -p $(BINARY_DIR)
	go build -ldflags "$(LDFLAGS)" -o $(BINARY_DIR)/gateway ./cmd/gateway

clean:
	rm -rf $(PROTO_DIR)

fclean: clean
	rm -f $(PROTO_GO_DIR)/*.pb.go

re: fclean all
