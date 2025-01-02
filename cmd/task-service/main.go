package main

import (
	"context"
	"fmt"
	"os"

	"log"
	"net"

	"github.com/my-backend-project/internal/pb"
	"github.com/my-backend-project/internal/task/handler"
	"github.com/my-backend-project/internal/task/interceptor"
	"github.com/my-backend-project/internal/task/repository"
	"github.com/my-backend-project/internal/task/service"
	"github.com/my-backend-project/internal/user/auth"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
)

func main() {
	// 環境変数の読み込み
	jwtSecretKey := os.Getenv("JWT_SECRET_KEY")
	if jwtSecretKey == "" {
		log.Fatal("JWT_SECRET_KEY is not set")
	}

	grpcPort := os.Getenv("GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "50051"
	}

	// MongoDB接続設定
	mongoClient, err := connectMongoDB()
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer mongoClient.Disconnect(context.Background())

	// リポジトリの初期化
	taskRepo := repository.NewTaskRepository(mongoClient.Database("task"))

	// サービスの初期化
	taskService := service.NewTaskService(taskRepo)

	// JWT サービスの初期化
	jwtService := auth.NewJWTService(jwtSecretKey)

	// 認証インターセプターの初期化
	authInterceptor := interceptor.NewAuthInterceptor(jwtService)

	// gRPCサーバーの初期化
	server := grpc.NewServer(
		grpc.UnaryInterceptor(authInterceptor.Unary()),
	)

	// タスクハンドラーの登録
	taskHandler := handler.NewTaskHandler(taskService)
	pb.RegisterTaskServiceServer(server, taskHandler)

	// サーバーの起動
	lis, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	log.Printf("Starting gRPC server on :%s", grpcPort)
	if err := server.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}

func connectMongoDB() (*mongo.Client, error) {
	ctx := context.Background()
	mongoURI := os.Getenv("MONGODB_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://localhost:27017"
	}

	clientOptions := options.Client().ApplyURI(mongoURI)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %v", err)
	}

	// 接続確認
	err = client.Ping(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %v", err)
	}

	return client, nil
}
