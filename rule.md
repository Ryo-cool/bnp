以下では「Go + gRPC + MongoDB」を学習・実践するためのサンプルプロジェクトを想定した要件定義とディレクトリ構造を提案します。
Kubernetes などの大規模オーケストレーションはスコープ外とし、Docker Compose ベースでローカル環境を作れる構成に絞ります。

1. 要件定義

1.1 システム概要
• 目的
Go と gRPC・MongoDB を用いたバックエンドを学習すること
• Web(REST) + gRPC 両方の実装・運用を体験し、使い分けを把握する
• MongoDB でのデータモデリングやトランザクションを簡単に触れる
• Docker Compose を使ってローカル開発環境を立ち上げる
• CI (GitHub Actions 等) で lint やユニットテスト、コンテナビルドを自動化
• サンプルドメイン
以下のようなシンプルなドメインを想定し、API を作成します。 1. ユーザー管理機能（認証・認可も含む） 2. タスク管理機能（CRUD）
• 最終的に学びたいトピック 1. HTTP (REST) と gRPC の使い分け 2. JWT 認証の実装 3. MongoDB を使ったスキーマ設計・CRUD 4. Docker Compose でのローカル環境構築 5. CI/CD (テスト・ビルド・コンテナプッシュ) の基本フロー

1.2 ユースケース・機能一覧 1. ユーザー管理 (User Service - REST API で実装)
• サインアップ（新規登録）
• ログイン（JWT 発行）
• ログアウト（トークン無効化もしくはクライアント側で破棄）
• (オプション) パスワードリセット など 2. タスク管理 (Task Service - gRPC で実装)
• CreateTask（タスクの作成）
• GetTask / ListTasks（単一のタスク取得・タスクリスト取得）
• UpdateTask（タスク更新）
• DeleteTask（タスク削除） 3. セキュリティ
• ユーザー操作時には JWT を使用して認証する
• gRPC 側もリクエストメタデータに JWT を含む設計で保護 4. DB 設計 (MongoDB)
• user コレクション：ユーザーアカウント情報 (email, hashedPassword, createdAt 等)
• task コレクション：タスク情報 (title, description, status, userID 等) 5. 運用要件
• Docker Compose で user-service・task-service・mongo を起動
• テストと Linter (golangci-lint 等) を GitHub Actions や他の CI サービスで実行
• (オプション) ビルド後のコンテナを Docker Hub などにプッシュできるように設定

2. ディレクトリ構造

下記は一例ですが、Go での一般的な構成例を意識しつつ、REST サービス (User Service) と gRPC サービス (Task Service) を分割しています。必要に応じてフォルダ・ファイルは調整してください。

my-backend-project/
├── cmd/
│ ├── user-service/
│ │ └── main.go // ユーザー管理用の REST API サーバーのエントリーポイント
│ └── task-service/
│ └── main.go // タスク管理用の gRPC サーバーのエントリーポイント
│
├── internal/
│ ├── user/
│ │ ├── handler/ // HTTP ハンドラー (Echo, Gin, net/http など)
│ │ │ └── user_handler.go
│ │ ├── service/ // ビジネスロジック (UseCase)
│ │ │ └── user_service.go
│ │ ├── repository/ // DB とのやり取り (MongoDB)
│ │ │ └── user_repository.go
│ │ └── auth/
│ │ └── jwt.go // JWT の発行、検証周り
│ │
│ ├── task/
│ │ ├── service/
│ │ │ └── task_service.go
│ │ ├── repository/
│ │ │ └── task_repository.go
│ │ └── interceptor/
│ │ └── auth_interceptor.go // gRPC の unary interceptor など
│ │
│ └── pkg/ // 共通で使うユーティリティ、ヘルパー等
│ ├── logger/
│ └── config/ // 設定ファイルの読み込みや環境変数管理
│
├── proto/
│ └── task.proto // Task Service 用の .proto 定義
│
├── test/
│ ├── user/
│ │ ├── user_handler_test.go
│ │ ├── user_service_test.go
│ │ └── ...
│ ├── task/
│ │ ├── task_service_test.go
│ │ ├── task_repository_test.go
│ │ └── ...
│ └── integration/
│ └── ...
│
├── deployments/
│ ├── docker/
│ │ ├── user-service/
│ │ │ └── Dockerfile
│ │ ├── task-service/
│ │ │ └── Dockerfile
│ │ └── docker-compose.yml
│ └── ci/
│ └── github-actions.yml // もしくは .github/workflows/xxx.yml に配置
│
├── go.mod
├── go.sum
├── .env.example // ローカル起動用の環境変数サンプル
└── README.md

ディレクトリ構造におけるポイント

1. cmd/
   各サービスのエントリーポイント。go build -o user-service ./cmd/user-service のようにビルド対象を明確化。
2. internal/
   • user, task のように、機能ごとにまとめる。
   • HTTP or gRPC のハンドラー (あるいは server) と、ビジネスロジック(UseCase)、DB リポジトリを分けることで、責務を明確にし、テストをしやすくする。
   • auth/ など認証関連はユーザーサービスだけでなくタスク側でも使うなら共通化しても良いが、初期のうちは分かりやすい形でも OK。
3. proto/
   • gRPC の .proto 定義を配置。
   • protoc + protoc-gen-go + protoc-gen-go-grpc などで生成される .pb.go は internal/task/pb/ などに生成して、依存を限定させるケースが多い。 4. test/
   • 単体テスト (unit) と 結合テスト (integration) を分けると管理しやすい。
   • テストは各機能別に配置し、Xxx_test.go 形式で分割。 5. deployments/
   • docker/ フォルダ配下で Dockerfile や docker-compose.yml を管理。
   • CI 設定 (GitHub Actions や CircleCI) は .github/workflows/xxx.yml などに置くか、deployments/ci/ などで一元管理しても構わない。 6. .env.example
   • 環境変数をまとめ、実運用では .env (git 管理外) に書き換えて使う運用を想定。
   • 例: MONGO_URI=mongodb://mongo:27017 / JWT_SECRET=your_jwt_secret / APP_ENV=dev 等

4. 開発フロー (例)

   1. proto ファイル定義 → gRPC コード生成
      • task.proto を定義し、protoc で Go 用のスタブ生成 (make generate などのスクリプト化が便利)
   2. User Service (REST)
      • cmd/user-service/main.go から HTTP サーバ起動
      • /auth/signup, /auth/login, /auth/logout などエンドポイント
      • DB (MongoDB) との通信は repository 層を通じて実行
      • 認証トークン (JWT) 発行と検証を実装
   3. Task Service (gRPC)
      • cmd/task-service/main.go で gRPC サーバ起動
      • RPC: CreateTask, GetTask, ListTasks, UpdateTask, DeleteTask
      • interceptor で JWT 検証を行う (認可処理)
   4. Docker Compose で起動
      • docker-compose up -d で user-service, task-service, mongo が一括起動
      • ポート割り当て例:
      • User Service → localhost:8080
      • Task Service → localhost:50051 (gRPC)
      • Mongo → localhost:27017
   5. テストと CI/CD
      • go test ./... でローカルテスト
      • GitHub Actions などでプルリクエスト時に自動実行 (lint + test)
      • master/main ブランチにマージされたら自動的に Docker イメージビルドと Push (オプション)

5. まとめ
   • ドメイン (User/Task) ごとにディレクトリを分け、責務を明確にする
   • cmd/ でエントリーポイントを分離して、マイクロサービス化を意識
   • Docker Compose を使って、ローカルで複数サービス + DB を起動できるようにする
   • CI/CD (lint, test, build) を自動化して開発効率とコード品質を向上

この要件定義とディレクトリ構造をベースに、実際にコードを書いてみると、Go、gRPC、MongoDB の連携や JWT 認証などの主要バックエンド技術を一通り学べます。小さくスコープを区切りながら実装を進め、都度テストして挙動を確認し、エラー対応やデバッグを繰り返すことで理解が深まるはずです。ぜひ参考にしてみてください。
