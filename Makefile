.PHONY: proto test build run clean

# Protocol Buffers
proto:
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		proto/task.proto

# テスト実行
test:
	go test -v ./...

# ビルド
build:
	go build -o bin/user-service ./cmd/user-service
	go build -o bin/task-service ./cmd/task-service

# サービス実行
run-user:
	go run ./cmd/user-service/main.go

run-task:
	go run ./cmd/task-service/main.go

# クリーンアップ
clean:
	rm -rf bin/
	rm -f internal/task/pb/*.pb.go 