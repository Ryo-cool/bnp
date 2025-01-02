package auth

import (
	"time"

	"my-backend-project/internal/user/model"

	"github.com/golang-jwt/jwt/v4"
)

type JWTClaims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

type JWTService interface {
	GenerateToken(user *model.User) (string, error)
	ValidateToken(tokenString string) (*JWTClaims, error)
}

type jwtService struct {
	secretKey string
	expiresIn time.Duration
}

func NewJWTService(secretKey string, expiresIn time.Duration) JWTService {
	return &jwtService{
		secretKey: secretKey,
		expiresIn: expiresIn,
	}
}

func (s *jwtService) GenerateToken(user *model.User) (string, error) {
	claims := &JWTClaims{
		UserID: user.ID.Hex(),
		Email:  user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.expiresIn)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.secretKey))
}

func (s *jwtService) ValidateToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.secretKey), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, jwt.ErrSignatureInvalid
}
