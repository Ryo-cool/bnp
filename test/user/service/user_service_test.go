package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/my-backend-project/internal/user/auth"
	"github.com/my-backend-project/internal/user/model"
	"github.com/my-backend-project/internal/user/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// モックリポジトリ
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *model.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepository) FindByID(ctx context.Context, id string) (*model.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

// モックJWTサービス
type MockJWTService struct {
	mock.Mock
}

func (m *MockJWTService) GenerateToken(user *model.User) (string, error) {
	args := m.Called(user)
	return args.String(0), args.Error(1)
}

func (m *MockJWTService) ValidateToken(tokenString string) (*auth.JWTClaims, error) {
	args := m.Called(tokenString)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.JWTClaims), args.Error(1)
}

func TestUserService_SignUp(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockJWT := new(MockJWTService)
	userService := service.NewUserService(mockRepo, mockJWT)

	tests := []struct {
		name          string
		input         *model.SignUpRequest
		mockBehavior  func()
		expectedError error
	}{
		{
			name: "正常系：ユーザー登録成功",
			input: &model.SignUpRequest{
				Email:    "test@example.com",
				Password: "password123",
			},
			mockBehavior: func() {
				mockRepo.On("FindByEmail", mock.Anything, "test@example.com").Return(nil, nil)
				mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*model.User")).Return(nil)
				mockJWT.On("GenerateToken", mock.AnythingOfType("*model.User")).Return("dummy.token.string", nil)
			},
			expectedError: nil,
		},
		{
			name: "異常系：既存ユーザー",
			input: &model.SignUpRequest{
				Email:    "existing@example.com",
				Password: "password123",
			},
			mockBehavior: func() {
				existingUser := &model.User{
					ID:        primitive.NewObjectID(),
					Email:     "existing@example.com",
					Password:  "hashedpassword",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}
				mockRepo.On("FindByEmail", mock.Anything, "existing@example.com").Return(existingUser, nil)
			},
			expectedError: service.ErrUserAlreadyExists,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockBehavior()

			resp, err := userService.SignUp(context.Background(), tt.input)

			if tt.expectedError != nil {
				assert.Equal(t, tt.expectedError, err)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.NotEmpty(t, resp.Token)
				assert.Equal(t, tt.input.Email, resp.User.Email)
			}

			mockRepo.AssertExpectations(t)
			mockJWT.AssertExpectations(t)
		})
	}
}

func TestUserService_Login(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockJWT := new(MockJWTService)
	userService := service.NewUserService(mockRepo, mockJWT)

	hashedPassword := "$2a$10$..." // 実際のハッシュ値

	tests := []struct {
		name          string
		input         *model.LoginRequest
		mockBehavior  func()
		expectedError error
	}{
		{
			name: "正常系：ログイン成功",
			input: &model.LoginRequest{
				Email:    "test@example.com",
				Password: "password123",
			},
			mockBehavior: func() {
				user := &model.User{
					ID:       primitive.NewObjectID(),
					Email:    "test@example.com",
					Password: hashedPassword,
				}
				mockRepo.On("FindByEmail", mock.Anything, "test@example.com").Return(user, nil)
				mockJWT.On("GenerateToken", mock.AnythingOfType("*model.User")).Return("dummy.token.string", nil)
			},
			expectedError: nil,
		},
		{
			name: "異常系：ユーザーが存在しない",
			input: &model.LoginRequest{
				Email:    "nonexistent@example.com",
				Password: "password123",
			},
			mockBehavior: func() {
				mockRepo.On("FindByEmail", mock.Anything, "nonexistent@example.com").Return(nil, nil)
			},
			expectedError: service.ErrInvalidCredentials,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockBehavior()

			resp, err := userService.Login(context.Background(), tt.input)

			if tt.expectedError != nil {
				assert.Equal(t, tt.expectedError, err)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.NotEmpty(t, resp.Token)
				assert.Equal(t, tt.input.Email, resp.User.Email)
			}

			mockRepo.AssertExpectations(t)
			mockJWT.AssertExpectations(t)
		})
	}
}
