.PHONY: proto clean

# Go関連の変数
GOPATH := $(shell go env GOPATH)
GO := go
GOBIN := $(GOPATH)/bin

# Protobufs関連の変数
PROTOC := protoc
PROTO_DIR := proto
GO_OUT_DIR := internal/pb

# 必要なツールのインストール
tools:
	$(GO) install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	$(GO) install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Protoファイルからコードを生成
proto: tools
	$(PROTOC) --proto_path=$(PROTO_DIR) \
		--go_out=$(GO_OUT_DIR) \
		--go_opt=paths=source_relative \
		--go-grpc_out=$(GO_OUT_DIR) \
		--go-grpc_opt=paths=source_relative \
		$(PROTO_DIR)/task.proto

# 生成されたコードを削除
clean:
	rm -f $(GO_OUT_DIR)/*.pb.go

# ビルド
build:
	$(GO) build -o bin/server cmd/server/main.go

# テスト
test:
	$(GO) test -v ./...

# 依存関係の更新
deps:
	$(GO) mod tidy
	$(GO) mod verify 


