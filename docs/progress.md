# プロジェクト進捗状況

## 完了したタスク

### ユーザーサービス

- [x] モデルの実装
- [x] バリデーターの実装
- [x] リポジトリ層の実装
- [x] サービス層の実装
- [x] JWT サービスの実装
- [x] ハンドラーの実装
- [x] テストの実装
  - [x] サービステスト
  - [x] JWT サービステスト
  - [x] ハンドラーテスト

### タスクサービス

- [x] Protocol Buffers の定義
- [x] モデルの実装
- [x] リポジトリ層の実装

## 進行中のタスク

- [ ] タスクサービスの実装
  - [ ] サービス層の実装
  - [ ] gRPC ハンドラーの実装
  - [ ] テストの実装

## 今後のタスク

- [ ] Docker 関連ファイルの実装
  - [ ] Dockerfile
  - [ ] docker-compose.yml
- [ ] CI/CD 設定の実装
  - [ ] GitHub Actions
  - [ ] テスト自動化
  - [ ] デプロイメントパイプライン
- [ ] ドキュメンテーション
  - [ ] API 仕様書
  - [ ] セットアップガイド
  - [ ] 運用ガイド

## 技術スタック

- 言語: Go
- フレームワーク: Echo
- データベース: MongoDB
- 認証: JWT
- API: REST (ユーザーサービス), gRPC (タスクサービス)
- テスト: testify
- プロトコル: Protocol Buffers

## プロジェクト構造

```
.
├── cmd
│   └── user-service
│       └── main.go
├── internal
│   ├── pkg
│   │   └── validator
│   │       └── validator.go
│   ├── task
│   │   ├── model
│   │   │   └── task.go
│   │   ├── pb
│   │   │   ├── task.pb.go
│   │   │   └── task_grpc.pb.go
│   │   └── repository
│   │       └── task_repository.go
│   └── user
│       ├── auth
│       │   └── jwt.go
│       ├── handler
│       │   └── user_handler.go
│       ├── model
│       │   └── user.go
│       ├── repository
│       │   └── user_repository.go
│       └── service
│           └── user_service.go
├── proto
│   ├── task.proto
│   ├── task.pb.go
│   └── task_grpc.pb.go
├── test
│   └── user
│       ├── auth
│       │   └── jwt_test.go
│       ├── handler
│       │   └── user_handler_test.go
│       └── service
│           └── user_service_test.go
├── deployments
│   └── docker
│       └── docker-compose.yml
├── docs
│   └── progress.md
├── .env.example
├── go.mod
└── go.sum
```

## 技術スタック

- 言語: Go
- フレームワーク: Echo
- データベース: MongoDB
- 認証: JWT
- API: REST (ユーザーサービス), gRPC (タスクサービス)
- テスト: testify
- プロトコル: Protocol Buffers

## 次のステップ

1. タスクサービスのサービス層実装
2. gRPC ハンドラーの実装
3. タスクサービスのテスト実装
4. Docker 環境の構築
