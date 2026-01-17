PROTO_DIR := internal/proto
PROTO_GO_DIR := internal/protobuf
PROTO_LUA_SCRIPT := internal/tools/proto_from_lua.py
PROTOC_GEN_GO := $(shell go env GOPATH)/bin/protoc-gen-go

.PHONY: lua-proto proto go all install-protoc-gen-go clean fclean re

lua-proto:
	python $(PROTO_LUA_SCRIPT)

install-protoc-gen-go:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest

proto: lua-proto install-protoc-gen-go
	protoc --plugin=protoc-gen-go=$(PROTOC_GEN_GO) --proto_path=$(PROTO_DIR) --go_out=$(PROTO_GO_DIR) --go_opt=paths=source_relative $(PROTO_DIR)/*.proto

go: proto

all: lua-proto proto

clean:
	rm -rf $(PROTO_DIR)

fclean: clean
	rm -f $(PROTO_GO_DIR)/*.pb.go

re: fclean all
