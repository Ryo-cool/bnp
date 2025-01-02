package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/my-backend-project/internal/pkg/validator"
	"github.com/my-backend-project/internal/user/auth"
	"github.com/my-backend-project/internal/user/handler"
	"github.com/my-backend-project/internal/user/model"
	"github.com/my-backend-project/internal/user/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) SignUp(ctx echo.Context, req *model.SignUpRequest) (*model.AuthResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.AuthResponse), args.Error(1)
}

func (m *MockUserService) Login(ctx echo.Context, req *model.LoginRequest) (*model.AuthResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.AuthResponse), args.Error(1)
}

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

func setupTest() (*echo.Echo, *MockUserService, *MockJWTService) {
	e := echo.New()
	e.Validator = validator.NewCustomValidator()
	mockService := new(MockUserService)
	mockJWT := new(MockJWTService)
	return e, mockService, mockJWT
}

func TestUserHandler_SignUp(t *testing.T) {
	e, mockService, mockJWT := setupTest()
	h := handler.NewUserHandler(mockService, mockJWT)

	tests := []struct {
		name           string
		requestBody    interface{}
		mockBehavior   func()
		expectedStatus int
		expectedError  string
	}{
		{
			name: "正常系：ユーザー登録成功",
			requestBody: model.SignUpRequest{
				Email:    "test@example.com",
				Password: "password123",
			},
			mockBehavior: func() {
				mockService.On("SignUp", mock.Anything, mock.AnythingOfType("*model.SignUpRequest")).
					Return(&model.AuthResponse{
						Token: "dummy.token.string",
						User: model.User{
							Email: "test@example.com",
						},
					}, nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "異常系：既存ユーザー",
			requestBody: model.SignUpRequest{
				Email:    "existing@example.com",
				Password: "password123",
			},
			mockBehavior: func() {
				mockService.On("SignUp", mock.Anything, mock.AnythingOfType("*model.SignUpRequest")).
					Return(nil, service.ErrUserAlreadyExists)
			},
			expectedStatus: http.StatusConflict,
			expectedError:  service.ErrUserAlreadyExists.Error(),
		},
		{
			name: "異常系：無効なメールアドレス",
			requestBody: model.SignUpRequest{
				Email:    "invalid-email",
				Password: "password123",
			},
			mockBehavior:   func() {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockBehavior()

			jsonBody, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/auth/signup", bytes.NewBuffer(jsonBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			err := h.SignUp(c)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)

			if tt.expectedError != "" {
				var response map[string]string
				json.Unmarshal(rec.Body.Bytes(), &response)
				assert.Equal(t, tt.expectedError, response["error"])
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestUserHandler_Login(t *testing.T) {
	e, mockService, mockJWT := setupTest()
	h := handler.NewUserHandler(mockService, mockJWT)

	tests := []struct {
		name           string
		requestBody    interface{}
		mockBehavior   func()
		expectedStatus int
		expectedError  string
	}{
		{
			name: "正常系：ログイン成功",
			requestBody: model.LoginRequest{
				Email:    "test@example.com",
				Password: "password123",
			},
			mockBehavior: func() {
				mockService.On("Login", mock.Anything, mock.AnythingOfType("*model.LoginRequest")).
					Return(&model.AuthResponse{
						Token: "dummy.token.string",
						User: model.User{
							Email: "test@example.com",
						},
					}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "異常系：無効な認証情報",
			requestBody: model.LoginRequest{
				Email:    "wrong@example.com",
				Password: "wrongpassword",
			},
			mockBehavior: func() {
				mockService.On("Login", mock.Anything, mock.AnythingOfType("*model.LoginRequest")).
					Return(nil, service.ErrInvalidCredentials)
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Invalid email or password",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockBehavior()

			jsonBody, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(jsonBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			err := h.Login(c)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)

			if tt.expectedError != "" {
				var response map[string]string
				json.Unmarshal(rec.Body.Bytes(), &response)
				assert.Equal(t, tt.expectedError, response["error"])
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestUserHandler_AuthMiddleware(t *testing.T) {
	e, _, mockJWT := setupTest()
	h := handler.NewUserHandler(nil, mockJWT)

	tests := []struct {
		name           string
		setupAuth      func(req *http.Request)
		mockBehavior   func()
		expectedStatus int
		expectedError  string
	}{
		{
			name: "正常系：有効なトークン",
			setupAuth: func(req *http.Request) {
				req.Header.Set(echo.HeaderAuthorization, "Bearer valid.token.string")
			},
			mockBehavior: func() {
				mockJWT.On("ValidateToken", "valid.token.string").
					Return(&auth.JWTClaims{
						UserID: "user123",
						Email:  "test@example.com",
					}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "異常系：トークンなし",
			setupAuth: func(req *http.Request) {
				// トークンを設定しない
			},
			mockBehavior:   func() {},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Missing authorization token",
		},
		{
			name: "異常系：無効なトークン",
			setupAuth: func(req *http.Request) {
				req.Header.Set(echo.HeaderAuthorization, "Bearer invalid.token")
			},
			mockBehavior: func() {
				mockJWT.On("ValidateToken", "invalid.token").
					Return(nil, echo.ErrUnauthorized)
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Invalid token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockBehavior()

			req := httptest.NewRequest(http.MethodGet, "/api/protected", nil)
			tt.setupAuth(req)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			middleware := h.AuthMiddleware(func(c echo.Context) error {
				return c.NoContent(http.StatusOK)
			})

			err := middleware(c)
			if err != nil {
				// エラーハンドリング
				he, ok := err.(*echo.HTTPError)
				if ok {
					assert.Equal(t, tt.expectedStatus, he.Code)
				}
			} else {
				assert.Equal(t, tt.expectedStatus, rec.Code)
			}

			mockJWT.AssertExpectations(t)
		})
	}
}
