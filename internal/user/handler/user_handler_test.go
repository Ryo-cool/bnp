package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/my-backend-project/internal/pkg/validator"
	"github.com/my-backend-project/internal/user/auth"
	"github.com/my-backend-project/internal/user/model"
	"github.com/my-backend-project/internal/user/service"

	"github.com/golang-jwt/jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockUserService struct {
	mock.Mock
}

func (m *mockUserService) SignUp(ctx context.Context, req *model.SignUpRequest) (*model.AuthResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.AuthResponse), args.Error(1)
}

func (m *mockUserService) Login(ctx context.Context, req *model.LoginRequest) (*model.AuthResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.AuthResponse), args.Error(1)
}

type mockJWTService struct {
	mock.Mock
}

func (m *mockJWTService) GenerateToken(user *model.User) (string, error) {
	args := m.Called(user)
	return args.String(0), args.Error(1)
}

func (m *mockJWTService) ValidateToken(token string) (*auth.JWTClaims, error) {
	args := m.Called(token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.JWTClaims), args.Error(1)
}

func setupTest() (*echo.Echo, *mockUserService, *mockJWTService) {
	e := echo.New()
	e.Validator = validator.NewCustomValidator()
	mockService := new(mockUserService)
	mockJWT := new(mockJWTService)
	return e, mockService, mockJWT
}

func TestUserHandler_SignUp(t *testing.T) {
	e, mockService, mockJWT := setupTest()
	handler := NewUserHandler(mockService, mockJWT)

	t.Run("正常系：ユーザー登録成功", func(t *testing.T) {
		req := &model.SignUpRequest{
			Email:    "test@example.com",
			Password: "password123",
		}
		reqJSON, _ := json.Marshal(req)

		resp := &model.AuthResponse{
			Token: "test-token",
			User: model.User{
				Email: req.Email,
			},
		}

		mockService.On("SignUp", mock.Anything, req).Return(resp, nil).Once()

		c := newTestContext(e, http.MethodPost, "/signup", bytes.NewBuffer(reqJSON))
		err := handler.SignUp(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, c.Response().Status)
		mockService.AssertExpectations(t)
	})

	t.Run("異常系：既存ユーザー", func(t *testing.T) {
		req := &model.SignUpRequest{
			Email:    "existing@example.com",
			Password: "password123",
		}
		reqJSON, _ := json.Marshal(req)

		mockService.On("SignUp", mock.Anything, req).Return(nil, service.ErrUserAlreadyExists).Once()

		c := newTestContext(e, http.MethodPost, "/signup", bytes.NewBuffer(reqJSON))
		err := handler.SignUp(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusConflict, c.Response().Status)

		rec := c.Response().Writer.(*httptest.ResponseRecorder)
		var respBody map[string]string
		err = json.NewDecoder(rec.Body).Decode(&respBody)
		assert.NoError(t, err)
		assert.Equal(t, service.ErrUserAlreadyExists.Error(), respBody["error"])
		mockService.AssertExpectations(t)
	})

	t.Run("異常系：無効なメールアドレス", func(t *testing.T) {
		req := &model.SignUpRequest{
			Email:    "invalid-email",
			Password: "password123",
		}
		reqJSON, _ := json.Marshal(req)

		c := newTestContext(e, http.MethodPost, "/signup", bytes.NewBuffer(reqJSON))
		err := handler.SignUp(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, c.Response().Status)

		rec := c.Response().Writer.(*httptest.ResponseRecorder)
		var respBody map[string]string
		err = json.NewDecoder(rec.Body).Decode(&respBody)
		assert.NoError(t, err)
		assert.Equal(t, "Key: 'SignUpRequest.Email' Error:Field validation for 'Email' failed on the 'email' tag", respBody["error"])
	})
}

func TestUserHandler_Login(t *testing.T) {
	e, mockService, mockJWT := setupTest()
	handler := NewUserHandler(mockService, mockJWT)

	t.Run("正常系：ログイン成功", func(t *testing.T) {
		req := &model.LoginRequest{
			Email:    "test@example.com",
			Password: "password123",
		}
		reqJSON, _ := json.Marshal(req)

		resp := &model.AuthResponse{
			Token: "test-token",
			User: model.User{
				Email: req.Email,
			},
		}

		mockService.On("Login", mock.Anything, req).Return(resp, nil).Once()

		c := newTestContext(e, http.MethodPost, "/login", bytes.NewBuffer(reqJSON))
		err := handler.Login(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, c.Response().Status)
		mockService.AssertExpectations(t)
	})

	t.Run("異常系：無効な認証情報", func(t *testing.T) {
		req := &model.LoginRequest{
			Email:    "wrong@example.com",
			Password: "wrongpass",
		}
		reqJSON, _ := json.Marshal(req)

		mockService.On("Login", mock.Anything, req).Return(nil, service.ErrInvalidCredentials).Once()

		c := newTestContext(e, http.MethodPost, "/login", bytes.NewBuffer(reqJSON))
		err := handler.Login(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, c.Response().Status)

		rec := c.Response().Writer.(*httptest.ResponseRecorder)
		var respBody map[string]string
		err = json.NewDecoder(rec.Body).Decode(&respBody)
		assert.NoError(t, err)
		assert.Equal(t, "Invalid email or password", respBody["error"])
		mockService.AssertExpectations(t)
	})
}

func TestUserHandler_AuthMiddleware(t *testing.T) {
	e, mockService, mockJWT := setupTest()
	handler := NewUserHandler(mockService, mockJWT)

	t.Run("正常系：有効なトークン", func(t *testing.T) {
		token := "valid-token"
		claims := &auth.JWTClaims{
			UserID: "user123",
			Email:  "test@example.com",
		}

		mockJWT.On("ValidateToken", token).Return(claims, nil).Once()

		c := newTestContext(e, http.MethodGet, "/protected", nil)
		c.Request().Header.Set("Authorization", "Bearer "+token)

		nextCalled := false
		next := func(c echo.Context) error {
			nextCalled = true
			return nil
		}

		err := handler.AuthMiddleware(next)(c)
		assert.NoError(t, err)
		assert.True(t, nextCalled)
		assert.Equal(t, claims.UserID, c.Get("user_id"))
		assert.Equal(t, claims.Email, c.Get("email"))
		mockJWT.AssertExpectations(t)
	})

	t.Run("異常系：トークンなし", func(t *testing.T) {
		c := newTestContext(e, http.MethodGet, "/protected", nil)

		next := func(c echo.Context) error {
			t.Fatal("Next handler should not be called")
			return nil
		}

		err := handler.AuthMiddleware(next)(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, c.Response().Status)
	})

	t.Run("異常系：無効なトークン", func(t *testing.T) {
		token := "invalid-token"
		mockJWT.On("ValidateToken", token).Return(nil, jwt.ErrSignatureInvalid).Once()

		c := newTestContext(e, http.MethodGet, "/protected", nil)
		c.Request().Header.Set("Authorization", "Bearer "+token)

		next := func(c echo.Context) error {
			t.Fatal("Next handler should not be called")
			return nil
		}

		err := handler.AuthMiddleware(next)(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, c.Response().Status)
		mockJWT.AssertExpectations(t)
	})
}

func newTestContext(e *echo.Echo, method, url string, body *bytes.Buffer) echo.Context {
	var req *http.Request
	if body != nil {
		req = httptest.NewRequest(method, url, body)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	} else {
		req = httptest.NewRequest(method, url, nil)
	}
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetRequest(req.WithContext(context.Background()))
	return c
}
