package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/my-backend-project/internal/user/auth"
	"github.com/my-backend-project/internal/user/model"
	"github.com/my-backend-project/internal/user/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockValidator はecho.Validatorインターフェースを実装するモックバリデーターです
type MockValidator struct {
	mock.Mock
}

func (m *MockValidator) Validate(i interface{}) error {
	args := m.Called(i)
	return args.Error(0)
}

// MockUserService はUserServiceのモック実装です
type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) SignUp(ctx context.Context, req *model.SignUpRequest) (*model.AuthResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.AuthResponse), args.Error(1)
}

func (m *MockUserService) Login(ctx context.Context, req *model.LoginRequest) (*model.AuthResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.AuthResponse), args.Error(1)
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

// setupTest はテスト用の共通セットアップを行います
func setupTest(t *testing.T) (*echo.Echo, *MockUserService, *MockJWTService, *MockValidator) {
	e := echo.New()
	mockValidator := new(MockValidator)
	e.Validator = mockValidator

	mockService := new(MockUserService)
	mockJWT := new(MockJWTService)
	return e, mockService, mockJWT, mockValidator
}

func TestUserHandler_SignUp(t *testing.T) {
	e, mockService, mockJWT, mockValidator := setupTest(t)
	handler := NewUserHandler(mockService, mockJWT)

	tests := []struct {
		name         string
		request      *model.SignUpRequest
		setup        func()
		validateErr  error
		expectedCode int
		expectedErr  string
	}{
		{
			name: "successful signup",
			request: &model.SignUpRequest{
				Email:    "test@example.com",
				Password: "password123",
			},
			setup: func() {
				mockValidator.On("Validate", mock.AnythingOfType("*model.SignUpRequest")).Return(nil).Once()
				response := &model.AuthResponse{
					Token: "token123",
					User: model.User{
						Email: "test@example.com",
					},
				}
				mockService.On("SignUp", mock.Anything, mock.AnythingOfType("*model.SignUpRequest")).Return(response, nil).Once()
			},
			validateErr:  nil,
			expectedCode: http.StatusCreated,
		},
		{
			name: "validation error - empty email",
			request: &model.SignUpRequest{
				Email:    "",
				Password: "password123",
			},
			setup: func() {
				mockValidator.On("Validate", mock.AnythingOfType("*model.SignUpRequest")).Return(errors.New("validation error")).Once()
			},
			validateErr:  errors.New("validation error"),
			expectedCode: http.StatusBadRequest,
			expectedErr:  "validation error",
		},
		{
			name: "user already exists",
			request: &model.SignUpRequest{
				Email:    "existing@example.com",
				Password: "password123",
			},
			setup: func() {
				mockValidator.On("Validate", mock.AnythingOfType("*model.SignUpRequest")).Return(nil).Once()
				mockService.On("SignUp", mock.Anything, mock.AnythingOfType("*model.SignUpRequest")).Return(nil, service.ErrUserAlreadyExists).Once()
			},
			validateErr:  nil,
			expectedCode: http.StatusConflict,
			expectedErr:  service.ErrUserAlreadyExists.Error(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 各テストケースの前にモックをリセット
			mockValidator.ExpectedCalls = nil
			mockService.ExpectedCalls = nil
			mockJWT.ExpectedCalls = nil

			tt.setup()

			jsonBytes, _ := json.Marshal(tt.request)
			req := httptest.NewRequest(http.MethodPost, "/auth/signup", bytes.NewReader(jsonBytes))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			err := handler.SignUp(c)
			if tt.expectedErr != "" {
				if assert.Error(t, err) {
					he, ok := err.(*echo.HTTPError)
					if assert.True(t, ok, "expected HTTP error") {
						assert.Equal(t, tt.expectedCode, he.Code)
						assert.Equal(t, tt.expectedErr, he.Message)
					}
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedCode, rec.Code)
			}

			// モックの期待通りの呼び出しを確認
			mockValidator.AssertExpectations(t)
			mockService.AssertExpectations(t)
			mockJWT.AssertExpectations(t)
		})
	}
}

func TestUserHandler_Login(t *testing.T) {
	e, mockService, mockJWT, mockValidator := setupTest(t)
	handler := NewUserHandler(mockService, mockJWT)

	tests := []struct {
		name         string
		request      *model.LoginRequest
		setup        func()
		validateErr  error
		expectedCode int
		expectedErr  string
	}{
		{
			name: "successful login",
			request: &model.LoginRequest{
				Email:    "test@example.com",
				Password: "password123",
			},
			setup: func() {
				mockValidator.On("Validate", mock.AnythingOfType("*model.LoginRequest")).Return(nil).Once()
				response := &model.AuthResponse{
					Token: "token123",
					User: model.User{
						Email: "test@example.com",
					},
				}
				mockService.On("Login", mock.Anything, mock.AnythingOfType("*model.LoginRequest")).Return(response, nil).Once()
			},
			validateErr:  nil,
			expectedCode: http.StatusOK,
		},
		{
			name: "validation error - empty email",
			request: &model.LoginRequest{
				Email:    "",
				Password: "password123",
			},
			setup: func() {
				mockValidator.On("Validate", mock.AnythingOfType("*model.LoginRequest")).Return(errors.New("validation error")).Once()
			},
			validateErr:  errors.New("validation error"),
			expectedCode: http.StatusBadRequest,
			expectedErr:  "validation error",
		},
		{
			name: "invalid credentials",
			request: &model.LoginRequest{
				Email:    "test@example.com",
				Password: "wrongpassword",
			},
			setup: func() {
				mockValidator.On("Validate", mock.AnythingOfType("*model.LoginRequest")).Return(nil).Once()
				mockService.On("Login", mock.Anything, mock.AnythingOfType("*model.LoginRequest")).Return(nil, service.ErrInvalidCredentials).Once()
			},
			validateErr:  nil,
			expectedCode: http.StatusUnauthorized,
			expectedErr:  "Invalid email or password",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 各テストケースの前にモックをリセット
			mockValidator.ExpectedCalls = nil
			mockService.ExpectedCalls = nil
			mockJWT.ExpectedCalls = nil

			tt.setup()

			jsonBytes, _ := json.Marshal(tt.request)
			req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(jsonBytes))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			err := handler.Login(c)
			if tt.expectedErr != "" {
				if assert.Error(t, err) {
					he, ok := err.(*echo.HTTPError)
					if assert.True(t, ok, "expected HTTP error") {
						assert.Equal(t, tt.expectedCode, he.Code)
						assert.Equal(t, tt.expectedErr, he.Message)
					}
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedCode, rec.Code)
			}

			// モックの期待通りの呼び出しを確認
			mockValidator.AssertExpectations(t)
			mockService.AssertExpectations(t)
			mockJWT.AssertExpectations(t)
		})
	}
}

func TestUserHandler_AuthMiddleware(t *testing.T) {
	e, mockService, mockJWT, _ := setupTest(t)
	handler := NewUserHandler(mockService, mockJWT)

	tests := []struct {
		name         string
		setupAuth    func(req *http.Request)
		setup        func()
		expectedCode int
		expectedErr  string
	}{
		{
			name: "valid token",
			setupAuth: func(req *http.Request) {
				req.Header.Set(echo.HeaderAuthorization, "Bearer valid-token")
			},
			setup: func() {
				claims := &auth.JWTClaims{
					UserID: "user123",
					Email:  "test@example.com",
				}
				mockJWT.On("ValidateToken", "valid-token").Return(claims, nil).Once()
			},
			expectedCode: http.StatusOK,
		},
		{
			name: "missing token",
			setupAuth: func(req *http.Request) {
				// トークンを設定しない
			},
			setup:        func() {},
			expectedCode: http.StatusUnauthorized,
			expectedErr:  "Missing authorization token",
		},
		{
			name: "invalid token format",
			setupAuth: func(req *http.Request) {
				req.Header.Set(echo.HeaderAuthorization, "invalid-token")
			},
			setup:        func() {},
			expectedCode: http.StatusUnauthorized,
			expectedErr:  "Invalid token format",
		},
		{
			name: "expired token",
			setupAuth: func(req *http.Request) {
				req.Header.Set(echo.HeaderAuthorization, "Bearer expired-token")
			},
			setup: func() {
				mockJWT.On("ValidateToken", "expired-token").Return(nil, auth.ErrTokenExpired).Once()
			},
			expectedCode: http.StatusUnauthorized,
			expectedErr:  "Token has expired",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 各テストケースの前にモックをリセット
			mockJWT.ExpectedCalls = nil

			tt.setup()

			req := httptest.NewRequest(http.MethodGet, "/protected", nil)
			tt.setupAuth(req)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			middleware := handler.AuthMiddleware(func(c echo.Context) error {
				return c.NoContent(http.StatusOK)
			})

			err := middleware(c)
			if tt.expectedErr != "" {
				if assert.Error(t, err) {
					he, ok := err.(*echo.HTTPError)
					if assert.True(t, ok, "expected HTTP error") {
						assert.Equal(t, tt.expectedCode, he.Code)
						assert.Equal(t, tt.expectedErr, he.Message)
					}
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedCode, rec.Code)
			}

			// モックの期待通りの呼び出しを確認
			mockJWT.AssertExpectations(t)
		})
	}
}
