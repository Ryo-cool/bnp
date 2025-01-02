package auth_test

import (
	"testing"
	"time"

	"github.com/my-backend-project/internal/user/auth"
	"github.com/my-backend-project/internal/user/model"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestJWTService_GenerateToken(t *testing.T) {
	// テスト用の秘密鍵とトークン有効期限
	secretKey := "test_secret_key"
	expiration := 24 * time.Hour

	jwtService := auth.NewJWTService(secretKey, expiration)

	tests := []struct {
		name        string
		user        *model.User
		shouldError bool
	}{
		{
			name: "正常系：トークン生成成功",
			user: &model.User{
				ID:    primitive.NewObjectID(),
				Email: "test@example.com",
			},
			shouldError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := jwtService.GenerateToken(tt.user)

			if tt.shouldError {
				assert.Error(t, err)
				assert.Empty(t, token)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, token)
			}
		})
	}
}

func TestJWTService_ValidateToken(t *testing.T) {
	// テスト用の秘密鍵とトークン有効期限
	secretKey := "test_secret_key"
	expiration := 24 * time.Hour

	jwtService := auth.NewJWTService(secretKey, expiration)

	// テスト用のユーザーを作成
	user := &model.User{
		ID:    primitive.NewObjectID(),
		Email: "test@example.com",
	}

	// 有効なトークンを生成
	validToken, err := jwtService.GenerateToken(user)
	assert.NoError(t, err)

	tests := []struct {
		name        string
		token       string
		shouldError bool
	}{
		{
			name:        "正常系：有効なトークン",
			token:       validToken,
			shouldError: false,
		},
		{
			name:        "異常系：無効なトークン",
			token:       "invalid.token.string",
			shouldError: true,
		},
		{
			name:        "異常系：空のトークン",
			token:       "",
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := jwtService.ValidateToken(tt.token)

			if tt.shouldError {
				assert.Error(t, err)
				assert.Nil(t, claims)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, claims)
				assert.Equal(t, user.ID.Hex(), claims.UserID)
				assert.Equal(t, user.Email, claims.Email)
			}
		})
	}
}
