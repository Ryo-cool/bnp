package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/my-backend-project/internal/user/model"
)

var (
	// エラー定数
	ErrInvalidToken   = errors.New("invalid token")
	ErrTokenExpired   = errors.New("token has expired")
	ErrTokenMalformed = errors.New("token is malformed")
)

// JWTClaims はJWTトークンのクレーム情報を表します
type JWTClaims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

// JWTService はJWTトークンの生成と検証を行うインターフェースです
type JWTService interface {
	GenerateToken(user *model.User) (string, error)
	ValidateToken(token string) (*JWTClaims, error)
}

// jwtService はJWTServiceの実装です
type jwtService struct {
	secretKey []byte
}

// NewJWTService は新しいJWTServiceインスタンスを作成します
func NewJWTService(secretKey string) JWTService {
	return &jwtService{
		secretKey: []byte(secretKey),
	}
}

// GenerateToken はユーザー情報からJWTトークンを生成します
func (s *jwtService) GenerateToken(user *model.User) (string, error) {
	claims := &JWTClaims{
		UserID: user.ID.Hex(),
		Email:  user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secretKey)
}

// ValidateToken はJWTトークンを検証し、クレーム情報を返します
func (s *jwtService) ValidateToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return s.secretKey, nil
	})

	if err != nil {
		if ve, ok := err.(*jwt.ValidationError); ok {
			if ve.Errors&jwt.ValidationErrorExpired != 0 {
				return nil, ErrTokenExpired
			}
			if ve.Errors&jwt.ValidationErrorMalformed != 0 {
				return nil, ErrTokenMalformed
			}
		}
		return nil, ErrInvalidToken
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, ErrInvalidToken
}
