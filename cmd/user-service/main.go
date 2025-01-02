package main

import (
	"context"
	"log"
	"os"
	"time"

	"my-backend-project/internal/pkg/validator"
	"my-backend-project/internal/user/auth"
	"my-backend-project/internal/user/handler"
	"my-backend-project/internal/user/repository"
	"my-backend-project/internal/user/service"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	// 環境変数の読み込み
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found")
	}

	// MongoDBクライアントの初期化
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://localhost:27017"
	}

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer client.Disconnect(ctx)

	// データベースとコレクションの初期化
	db := client.Database(os.Getenv("MONGO_DB_NAME"))

	// 依存関係の初期化
	userRepo := repository.NewUserRepository(db)
	jwtExpiration, err := time.ParseDuration(os.Getenv("JWT_EXPIRATION"))
	if err != nil {
		jwtExpiration = 24 * time.Hour // デフォルト値
	}
	jwtService := auth.NewJWTService(os.Getenv("JWT_SECRET"), jwtExpiration)
	userService := service.NewUserService(userRepo, jwtService)
	userHandler := handler.NewUserHandler(userService, jwtService)

	// Echoインスタンスの作成
	e := echo.New()

	// ミドルウェアの設定
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// カスタムバリデーターの設定
	e.Validator = validator.NewCustomValidator()

	// ルーティングの設定
	auth := e.Group("/auth")
	{
		auth.POST("/signup", userHandler.SignUp)
		auth.POST("/login", userHandler.Login)
	}

	// 認証が必要なルートのグループ
	api := e.Group("/api")
	api.Use(userHandler.AuthMiddleware)
	{
		// 認証が必要なエンドポイントをここに追加
	}

	// サーバーの起動
	port := os.Getenv("PORT_USER_SERVICE")
	if port == "" {
		port = "8080"
	}
	if err := e.Start(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
