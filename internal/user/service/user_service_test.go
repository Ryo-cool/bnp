package service

import (
	"context"
	"testing"

	"github.com/my-backend-project/internal/user/auth"
	"github.com/my-backend-project/internal/user/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

// MockUserRepository はUserRepositoryのモック実装です
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

// MockJWTService はJWTServiceのモック実装です
type MockJWTService struct {
	mock.Mock
}

func (m *MockJWTService) GenerateToken(user *model.User) (string, error) {
	args := m.Called(user)
	return args.String(0), args.Error(1)
}

func (m *MockJWTService) ValidateToken(token string) (*auth.JWTClaims, error) {
	args := m.Called(token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.JWTClaims), args.Error(1)
}

func TestUserService_SignUp(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockUserRepository)
	mockJWT := new(MockJWTService)
	service := NewUserService(mockRepo, mockJWT)

	tests := []struct {
		name    string
		req     *model.SignUpRequest
		setup   func()
		wantErr error
	}{
		{
			name: "successful signup",
			req: &model.SignUpRequest{
				Email:    "test@example.com",
				Password: "password123",
			},
			setup: func() {
				mockRepo.On("FindByEmail", ctx, "test@example.com").Return(nil, nil)
				mockRepo.On("Create", ctx, mock.AnythingOfType("*model.User")).Return(nil)
				mockJWT.On("GenerateToken", mock.AnythingOfType("*model.User")).Return("token123", nil)
			},
			wantErr: nil,
		},
		{
			name: "user already exists",
			req: &model.SignUpRequest{
				Email:    "existing@example.com",
				Password: "password123",
			},
			setup: func() {
				existingUser := &model.User{
					ID:       primitive.NewObjectID(),
					Email:    "existing@example.com",
					Password: "hashedpassword",
				}
				mockRepo.On("FindByEmail", ctx, "existing@example.com").Return(existingUser, nil)
			},
			wantErr: ErrUserAlreadyExists,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			resp, err := service.SignUp(ctx, tt.req)
			if tt.wantErr != nil {
				assert.Equal(t, tt.wantErr, err)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.NotEmpty(t, resp.Token)
				assert.Equal(t, tt.req.Email, resp.User.Email)
			}
		})
	}
}

func TestUserService_Login(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockUserRepository)
	mockJWT := new(MockJWTService)
	service := NewUserService(mockRepo, mockJWT)

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	existingUser := &model.User{
		ID:       primitive.NewObjectID(),
		Email:    "test@example.com",
		Password: string(hashedPassword),
	}

	tests := []struct {
		name    string
		req     *model.LoginRequest
		setup   func()
		wantErr error
	}{
		{
			name: "successful login",
			req: &model.LoginRequest{
				Email:    "test@example.com",
				Password: "password123",
			},
			setup: func() {
				mockRepo.On("FindByEmail", ctx, "test@example.com").Return(existingUser, nil)
				mockJWT.On("GenerateToken", existingUser).Return("token123", nil)
			},
			wantErr: nil,
		},
		{
			name: "user not found",
			req: &model.LoginRequest{
				Email:    "nonexistent@example.com",
				Password: "password123",
			},
			setup: func() {
				mockRepo.On("FindByEmail", ctx, "nonexistent@example.com").Return(nil, nil)
			},
			wantErr: ErrInvalidCredentials,
		},
		{
			name: "invalid password",
			req: &model.LoginRequest{
				Email:    "test@example.com",
				Password: "wrongpassword",
			},
			setup: func() {
				mockRepo.On("FindByEmail", ctx, "test@example.com").Return(existingUser, nil)
			},
			wantErr: ErrInvalidCredentials,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			resp, err := service.Login(ctx, tt.req)
			if tt.wantErr != nil {
				assert.Equal(t, tt.wantErr, err)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.NotEmpty(t, resp.Token)
				assert.Equal(t, tt.req.Email, resp.User.Email)
			}
		})
	}
}
